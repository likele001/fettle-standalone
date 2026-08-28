"""P1-1/P1-2 商业化配额模块（供 chat / workflow 复用）。

能力：
  - check_quota：调用 LLM 前校验租户配额；超限抛 QuotaExceededError。
     （订阅 token_limit 用尽、余额不足时拒绝）
  - circuit breaker（P1-2）：并发放行后仍大量失败/超配额时，Redis 短期熔断，
     避免在被拒租户上反复穿透 DB 查询。
  - usage alert（P1-2）：用量达到配额阈值时记录告警（日志 + Redis 时间窗去重）。
  - consume_quota：LLM 成功后记账（累加 token、扣减余额）。
"""
import logging
import os
import time
from typing import Dict, Any, Optional

from core.models.db_config import ai_config_db

logger = logging.getLogger(__name__)

# 本地计费开关：业务已统一走 billing-service 记账，本地表默认不再读写
LOCAL_BILLING_ENABLED = os.getenv("AI_ENGINE_LOCAL_BILLING", "0") == "1"

# 熔断/告警配置（可通过环境变量覆盖）
CIRCUIT_ENABLED = os.getenv("QUOTA_CIRCUIT_ENABLED", "true").lower() == "true"
CIRCUIT_FAIL_THRESHOLD = int(os.getenv("QUOTA_CIRCUIT_FAIL_THRESHOLD", "10"))
CIRCUIT_OPEN_SECONDS = int(os.getenv("QUOTA_CIRCUIT_OPEN_SECONDS", "30"))
ALERT_RATIO = float(os.getenv("QUOTA_ALERT_RATIO", "0.9"))  # 用量占限额>90%告警


class QuotaExceededError(Exception):
    """配额不足异常，message 为用户可见提示"""
    pass


class QuotaCircuitOpen(QuotaExceededError):
    """熔断打开时抛出的内部异常（可被上层当作配额拒绝向用户展示）"""
    pass


def _redis():
    """懒创建 Redis 客户端，失败返回 None（熔断/告警为尽力而为）。"""
    try:
        from config.settings import settings
        import redis.asyncio as aioredis
        client = aioredis.from_url(settings.redis_url, decode_responses=True)
        return client
    except Exception:
        return None


def _circuit_notify_key(tenant_id: str) -> str:
    return f"quota:circuit:notify:{tenant_id}"


def _circuit_open_key(tenant_id: str) -> str:
    return f"quota:circuit:open:{tenant_id}"


async def _record_failure(tenant_id: str) -> bool:
    """记录一次配额失败；达到阈值打开熔断。返回是否本次打开熔断。"""
    if not CIRCUIT_ENABLED:
        return False
    client = _redis()
    if client is None:
        return False
    try:
        key = f"quota:circuit:fail:{tenant_id}"
        n = await client.incr(key)
        await client.expire(key, CIRCUIT_OPEN_SECONDS)
        if n >= CIRCUIT_FAIL_THRESHOLD:
            await client.set(_circuit_open_key(tenant_id), "1", ex=CIRCUIT_OPEN_SECONDS)
            # 防止每个请求都打告警日志
            nf = await client.set(_circuit_notify_key(tenant_id), "1", nx=True, ex=60)
            if nf:
                logger.warning(f"QUOTA CIRCUIT OPEN: tenant={tenant_id} after {n} consecutive failures")
            return True
    except Exception as e:
        logger.warning(f"quota circuit record_failure error: {e}")
    finally:
        try:
            await client.aclose()
        except Exception:
            pass
    return False


async def _is_open(tenant_id: str) -> bool:
    """判断租户熔断是否打开。"""
    if not CIRCUIT_ENABLED:
        return False
    client = _redis()
    if client is None:
        return False
    try:
        return bool(await client.exists(_circuit_open_key(tenant_id)))
    except Exception:
        return False
    finally:
        try:
            await client.aclose()
        except Exception:
            pass


async def _fire_alert(tenant_id: str, ratio: float, used: int, limit: int) -> None:
    """用量达到阈值时记录告警（日志 + Redis 时间窗去重，避免刷屏）。"""
    if ratio < ALERT_RATIO:
        return
    client = _redis()
    if client is not None:
        try:
            key = f"quota:alert:{tenant_id}"
            fired = await client.set(key, "1", nx=True, ex=300)  # 5 分钟去重窗口
            await client.aclose()
            if not fired:
                return
        except Exception:
            pass
    logger.warning(f"QUOTA ALERT: tenant={tenant_id} usage={used}/{limit} ratio={ratio:.1%}")


async def check_quota(tenant_id: str) -> Dict[str, Any]:
    """
    在调用 LLM 前校验租户配额；若超限抛 QuotaExceededError。
    策略：
      0. 若该租户熔断已打开 -> 直接拒绝
      1. 订阅套餐 enforces token_limit：tokens_used >= token_limit -> 拒绝
      2. 无订阅且余额记录 <= 0（且曾充值过） -> 拒绝（余额不足）
      3. 达标【接近限额】-> 触发用量告警
      4. 否则放行（免费/未启用计费租户）。
    """
    if not LOCAL_BILLING_ENABLED:
        return None
    # 熔断快速失败
    if await _is_open(tenant_id):
        raise QuotaCircuitOpen("服务暂时繁忙，请稍后重试。")

    q = await ai_config_db.get_tenant_quota(tenant_id)

    # 策略1：订阅 token 限额
    if q.get('enforced') and q.get('token_limit'):
        used = q.get('tokens_used') or 0
        limit = int(q['token_limit'])
        if used >= limit:
            await _record_failure(tenant_id)
            raise QuotaExceededError("当前套餐的 Token 用量已达上限，请升级套餐或等待下个周期重置。")
        # 接近限额告警
        await _fire_alert(tenant_id, used / limit, used, limit)
        # 通过则清零失败计数（成功放行）
        if CIRCUIT_ENABLED:
            client = _redis()
            if client is not None:
                try:
                    await client.delete(f"quota:circuit:fail:{tenant_id}")
                except Exception:
                    pass
                finally:
                    try:
                        await client.aclose()
                    except Exception:
                        pass

    # 策略2：余额不足（有余额记录且从未充值充足）
    balance = q.get('balance')
    has_active_sub = q.get('has_active_sub')
    if balance is not None and not has_active_sub:
        total_recharged = q.get('total_recharged') or 0
        if float(total_recharged) > 0 and float(balance) <= 0:
            await _record_failure(tenant_id)
            raise QuotaExceededError("账户余额不足，请充值后继续使用 AI 服务。")

    return q


async def consume_quota(tenant_id: str, total_tokens: int, total_cost: float) -> None:
    """LLM 调用成功后记账：累加已用 token，扣减余额。"""
    if not LOCAL_BILLING_ENABLED:
        logger.warning("consume_quota skipped: AI_ENGINE_LOCAL_BILLING=0 (billing-service 统一记账)")
        return

    try:
        await ai_config_db.consume_usage(tenant_id, total_tokens, total_cost)
    except Exception as e:
        logger.error(f"consume_quota failed for tenant {tenant_id}: {e}")