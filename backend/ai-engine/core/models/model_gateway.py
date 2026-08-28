"""P2-1 模型统一代理/网关层。

在三方面增强现有「按租户 key 的 DB 模型路由 + Provider 工厂」：
  1) key 托管：租户未配置 key 时，回退到平台级托管 key（环境变量/平台配置）。
  2) 聚合降级：主模型失败后，自动按候选列表顺序尝试备选 provider/model。
  3) provider 级熔断：Redis 记录某 provider+model 连续失败，打开冷却窗口，期间跳过该候选。

设计上作为「薄代理」包裹现有 db_model_router / DBProviderFactory，
不替换底层 provider 实现，主调用方可零侵入地获得降级与熔断能力。
"""
import logging
import os
import time
from typing import List, Dict, Any, Optional, AsyncIterator

from .providers.base import BaseProvider, ChatMessage, ChatResponse
from .db_model_router import db_model_router, ModelConfig

logger = logging.getLogger(__name__)

# 平台级托管 key（环境变量）。与 config/settings.py 各厂商 key 字段对应。
_PLATFORM_KEYS = {
    "aliyun": "aliyun_api_key",
    "qwen": "qwen_api_key",
    "openai": "openai_api_key",
    "deepseek": "deepseek_api_key",
    "minimax": "minimax_api_key",
}

# 熔断参数（环境变量可覆盖）
_BREAKER_ENABLED = os.getenv("MODEL_BREAKER_ENABLED", "true").lower() == "true"
_BREAKER_FAIL_THRESHOLD = int(os.getenv("MODEL_BREAKER_FAIL_THRESHOLD", "5"))
_BREAKER_COOLDOWN = int(os.getenv("MODEL_BREAKER_COOLDOWN", "60"))  # 秒


def _get_platform_api_key(provider_code: str) -> Optional[str]:
    """从 env/settings 读取平台级托管 key。"""
    key = _PLATFORM_KEYS.get(provider_code or "")
    if not key:
        return None
    return os.getenv(key.upper()) or None


def _redis():
    try:
        from config.settings import settings
        import redis.asyncio as aioredis
        return aioredis.from_url(settings.redis_url, decode_responses=True)
    except Exception:
        return None


def _breaker_fail_key(provider, model):
    return f"mbreak:fail:{provider}:{model}"


def _breaker_open_key(provider, model):
    return f"mbreak:open:{provider}:{model}"


async def _breaker_should_try(provider_code: str, model: str) -> bool:
    """查询该 provider+model 是否处于熔断打开状态。"""
    if not _BREAKER_ENABLED:
        return True
    client = _redis()
    if client is None:
        return True
    try:
        if await client.exists(_breaker_open_key(provider_code, model)):
            return False
        return True
    except Exception:
        return True
    finally:
        try:
            await client.aclose()
        except Exception:
            pass


async def _breaker_record(provider_code: str, model: str, success: bool) -> None:
    """吞吐后记录失败/清零成功。"""
    if not _BREAKER_ENABLED:
        return
    client = _redis()
    if client is None:
        return
    try:
        if success:
            await client.delete(_breaker_fail_key(provider_code, model))
            return
        k = _breaker_fail_key(provider_code, model)
        n = await client.incr(k)
        await client.expire(k, _BREAKER_COOLDOWN)
        if n >= _BREAKER_FAIL_THRESHOLD:
            await client.set(
                _breaker_open_key(provider_code, model), "1", ex=_BREAKER_COOLDOWN
            )
            await client.delete(k)
            logger.warning(
                f"MODEL BREAKER OPEN: provider={provider_code} model={model} "
                f"after {n} consecutive failures"
            )
    except Exception:
        pass
    finally:
        try:
            await client.aclose()
        except Exception:
            pass


async def resolve_effective_key(
    tenant_id: str,
    provider_code: str,
    provider_id: str,
    prefer_tenant: bool = True,
) -> Optional[str]:
    """key 托管：优先租户 key（仅对首选候选安全，其 provider_id 与租户 key 对应），
    失败回退平台级托管 key。prefer_tenant=False 时（降级候选）直接走平台 key，
    避免把首选厂商的租户 key 错配给其他厂商。"""
    # 1. 租户 key（仅首选候选）
    if prefer_tenant:
        try:
            tenant_key = await db_model_router.get_tenant_api_key(tenant_id, provider_id)
            if tenant_key and tenant_key.api_key_value:
                return tenant_key.api_key_value
        except Exception as e:
            logger.warning(f"resolve tenant key failed: {e}")

    # 2. 平台级托管 key
    platform_key = _get_platform_api_key(provider_code)
    if platform_key:
        logger.info(f"using platform-managed key for provider={provider_code}")
        return platform_key
    return None


async def _build_candidate_providers(
    tenant_id: str,
    preferred: ModelConfig,
) -> List[tuple]:
    """构造候选 (provider_code, model_code, base_url) 列表，按优先级排序。

    优先级：
      1) 首选 provider 的本 model
      2) 首选 provider 的其他 chat model
      3) 租户关键可用 provider 的 chat model
      4) 平台托管 key 可用的 provider 的 chat model
    熔断已打开的候选会被跳过。
    """
    candidates: List[tuple] = []
    seen = set()

    try:
        models = await db_model_router.list_models()

        def chat_capable(m: Dict) -> bool:
            return "chat" in (m.get("capabilities") or [])

        def append(pcode, mcode):
            key = (pcode, mcode)
            if key not in seen:
                seen.add(key)
                candidates.append((pcode, mcode))

        # 1. 首选
        append(preferred.provider_code, preferred.model_code)

        # 2. 首选 provider 的其他 chat model
        for m in models:
            if m.get("provider_code") == preferred.provider_code and chat_capable(m):
                append(preferred.provider_code, m.get("model_code"))

        # 3. 平台托管 key 可用 provider 的 chat model（租户无 key 时的兜底通道）
        for m in models:
            if chat_capable(m) and _get_platform_api_key(m.get("provider_code")):
                append(m.get("provider_code"), m.get("model_code"))
    except Exception as e:
        logger.warning(f"build candidates failed: {e}")
        if not candidates:
            candidates.append((preferred.provider_code, preferred.model_code))

    return candidates


class ModelGateway:
    """统一模型网关：聚合降级 + 熔断 + key 托管。"""

    async def chat(
        self,
        tenant_id: str,
        preferred: ModelConfig,
        messages: List[ChatMessage],
        max_tokens: int = 2000,
        temperature: float = 0.7,
    ) -> ChatResponse:
        """带降级/熔断的 chat 调用。"""
        from .providers.db_factory import DBProviderFactory

        errors: List[str] = []
        candidates = await _build_candidate_providers(tenant_id, preferred)

        for idx, (provider_code, model_code) in enumerate(candidates):
            is_preferred = (idx == 0)
            # 熔断快速跳过
            if not await _breaker_should_try(provider_code, model_code):
                errors.append(f"{provider_code}/{model_code} breaker open")
                continue

            provider = None
            try:
                api_key = await resolve_effective_key(
                    tenant_id, provider_code, preferred.provider_id, prefer_tenant=is_preferred
                )
                if not api_key:
                    errors.append(f"{provider_code}/{model_code} no key")
                    _breaker_record(provider_code, model_code, True)  # 无 key 不算熔断
                    continue

                # 仅首选传租户 key；降级候选只走平台 key，避免厂商 key 错配
                tenant_api_key_arg = None
                if is_preferred:
                    try:
                        tenant_api_key_arg = await db_model_router.get_tenant_api_key(
                            tenant_id, preferred.provider_id
                        )
                    except Exception:
                        tenant_api_key_arg = None

                provider = await DBProviderFactory.get_provider(
                    provider_code,
                    tenant_api_key=tenant_api_key_arg,
                    api_key=api_key,
                    use_langchain=False,
                )
                if not provider:
                    errors.append(f"{provider_code}/{model_code} provider unavailable")
                    continue

                resp = await provider.chat(
                    messages=messages,
                    model=model_code,
                    max_tokens=max_tokens,
                    temperature=temperature,
                )
                await _breaker_record(provider_code, model_code, True)
                # 异步关闭避免连接泄漏
                try:
                    await provider.close()
                except Exception:
                    pass
                return resp
            except Exception as e:
                err = f"{provider_code}/{model_code}: {e}"
                errors.append(err)
                logger.warning(f"chat fallback skip {err}")
                await _breaker_record(provider_code, model_code, False)
                if provider is not None:
                    try:
                        await provider.close()
                    except Exception:
                        pass

        raise RuntimeError("; ".join(errors) or "no provider available")

    async def stream_chat(
        self,
        tenant_id: str,
        preferred: ModelConfig,
        messages: List[ChatMessage],
        max_tokens: int = 2000,
        temperature: float = 0.7,
        usage_holder: dict = None,
    ) -> AsyncIterator[str]:
        """带降级/熔断的流式调用（聚合降级时逐候选重试）。"""
        from .providers.db_factory import DBProviderFactory

        candidates = await _build_candidate_providers(tenant_id, preferred)
        errors: List[str] = []

        for idx, (provider_code, model_code) in enumerate(candidates):
            is_preferred = (idx == 0)
            if not await _breaker_should_try(provider_code, model_code):
                errors.append(f"{provider_code}/{model_code} breaker open")
                continue

            provider = None
            try:
                api_key = await resolve_effective_key(
                    tenant_id, provider_code, preferred.provider_id, prefer_tenant=is_preferred
                )
                if not api_key:
                    errors.append(f"{provider_code}/{model_code} no key")
                    continue

                tenant_api_key_arg = None
                if is_preferred:
                    try:
                        tenant_api_key_arg = await db_model_router.get_tenant_api_key(
                            tenant_id, preferred.provider_id
                        )
                    except Exception:
                        tenant_api_key_arg = None

                provider = await DBProviderFactory.get_provider(
                    provider_code,
                    tenant_api_key=tenant_api_key_arg,
                    api_key=api_key,
                    use_langchain=False,
                )
                if not provider:
                    errors.append(f"{provider_code}/{model_code} provider unavailable")
                    continue

                collected = []
                provider_gen = provider.stream_chat(
                    messages=messages,
                    model=model_code,
                    max_tokens=max_tokens,
                    temperature=temperature,
                )
                holder = {}
                try:
                    async for chunk in provider.stream_chat(
                        messages=messages,
                        model=model_code,
                        max_tokens=max_tokens,
                        temperature=temperature,
                        usage_holder=holder,
                    ):
                        collected.append(chunk)
                        yield chunk
                except Exception:
                    await _breaker_record(provider_code, model_code, False)
                    raise
                if usage_holder is not None:
                    usage_holder["input"] = holder.get("input", 0)
                    usage_holder["output"] = holder.get("output", 0)
                await _breaker_record(provider_code, model_code, True)
                try:
                    await provider.close()
                except Exception:
                    pass
                return
            except Exception as e:
                err = f"{provider_code}/{model_code}: {e}"
                errors.append(err)
                logger.warning(f"stream fallback skip {err}")
                await _breaker_record(provider_code, model_code, False)
                if provider is not None:
                    try:
                        await provider.close()
                    except Exception:
                        pass

        if errors:
            raise RuntimeError("; ".join(errors))


# 全局网关实例
model_gateway = ModelGateway()