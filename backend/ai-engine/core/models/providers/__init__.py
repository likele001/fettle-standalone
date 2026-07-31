"""提供商工厂"""
import logging
from typing import Optional

from .base import BaseProvider, ChatMessage, ChatResponse, EmbeddingResponse
from .qwen import QwenProvider
from .deepseek import DeepSeekProvider
from .openai_provider import OpenAIProvider
from .minimax_provider import MiniMaxProvider
from config.settings import settings

logger = logging.getLogger(__name__)


# provider_code 映射到 ai_providers.code 的别名（兼容 yaml 用的 qwen/openai/deepseek/minimax）
PROVIDER_CODE_ALIASES = {
    "qwen": ["qwen", "aliyun", "dashscope"],
    "openai": ["openai"],
    "deepseek": ["deepseek"],
    "minimax": ["minimax"],
    "moonshot": ["moonshot"],
    "zhipu": ["zhipu"],
}


class ProviderFactory:
    """模型提供商工厂"""

    _providers = {
        'qwen': QwenProvider,
        'deepseek': DeepSeekProvider,
        'openai': OpenAIProvider,
        'minimax': MiniMaxProvider,
    }

    _instances = {}

    @classmethod
    def _resolve_class(cls, provider_name: str):
        """把 ai_providers.code (aliyun/dashscope/moonshot/...) 映射到内部 provider class"""
        if provider_name in cls._providers:
            return cls._providers[provider_name], provider_name
        # 反向别名映射: ali_provider_code -> internal
        for internal, aliases in PROVIDER_CODE_ALIASES.items():
            if provider_name in aliases and internal in cls._providers:
                return cls._providers[internal], internal
        return None, provider_name

    @classmethod
    def get_provider(cls, provider_name: str) -> Optional[BaseProvider]:
        """获取提供商实例（同步版本，从全局 .env 读取 API key）"""
        if provider_name in cls._instances:
            return cls._instances[provider_name]

        provider_class, internal_code = cls._resolve_class(provider_name)
        if not provider_class:
            logger.warning(f"Unknown provider: {provider_name}")
            return None

        # 从全局 settings 读 API key (用 internal_code)
        api_key = getattr(settings, f"{internal_code}_api_key", None)
        if not api_key:
            logger.warning(f"No API key for provider: {provider_name}")
            return None

        instance = provider_class(api_key=api_key)
        cls._instances[provider_name] = instance
        return instance

    @classmethod
    async def get_provider_async(
        cls,
        provider_name: str,
        tenant_id: Optional[str] = None,
    ) -> Optional[BaseProvider]:
        """获取提供商实例（异步版本，优先读租户 API key，回退到全局 .env key）。

        参数:
            provider_name: 厂商代码（qwen / openai / deepseek / minimax）
            tenant_id:     租户 ID（从 gRPC request.tenant_id 传入），用于从 DB 读租户自己的 API key

        优先级:
            1. DB tenant_api_keys（status=active）
            2. 全局 .env (settings.{provider}_api_key)
        """
        # 缓存 key
        cache_key = f"{tenant_id}:{provider_name}" if tenant_id else provider_name
        if cache_key in cls._instances:
            return cls._instances[cache_key]

        provider_class, internal_code = cls._resolve_class(provider_name)
        if not provider_class:
            logger.warning(f"Unknown provider: {provider_name}")
            return None

        api_key = None
        base_url = None
        custom_headers = {}

        # 1) 优先从数据库读取租户 API key
        if tenant_id:
            try:
                from core.models.db_config import ai_config_db
                provider = None
                for code in PROVIDER_CODE_ALIASES.get(provider_name, [provider_name]):
                    provider = await ai_config_db.get_provider(provider_code=code)
                    if provider:
                        break
                if provider:
                    tk = await ai_config_db.get_tenant_api_key(tenant_id, provider.id)
                    if tk and tk.api_key_value:
                        api_key = tk.api_key_value
                        base_url = tk.custom_base_url or None
                        custom_headers = tk.custom_headers or {}
                        logger.info(
                            f"Loaded tenant API key for provider={provider_name}, tenant={tenant_id}"
                        )
            except Exception as e:
                logger.warning(f"Failed to load tenant API key for {provider_name}/{tenant_id}: {e}")

        # 2) Fallback 到全局 .env
        if not api_key:
            api_key = getattr(settings, f"{internal_code}_api_key", None)
            if api_key:
                logger.info(f"Using global .env API key for provider={provider_name}")

        if not api_key:
            logger.warning(
                f"No API key for provider: {provider_name} (tenant={tenant_id or '-'})"
            )
            return None

        try:
            kwargs = {}
            if base_url:
                kwargs["base_url"] = base_url
            if custom_headers:
                kwargs["custom_headers"] = custom_headers
            instance = provider_class(api_key, **kwargs)
            cls._instances[cache_key] = instance
            logger.info(
                f"Provider instance created: provider={provider_name}, tenant={tenant_id or '-'}"
            )
            return instance
        except Exception as e:
            logger.error(f"Failed to create provider {provider_name}: {e}")
            return None

    @classmethod
    async def close_all(cls):
        """关闭所有提供商"""
        for instance in cls._instances.values():
            try:
                await instance.close()
            except Exception:
                pass
        cls._instances.clear()
