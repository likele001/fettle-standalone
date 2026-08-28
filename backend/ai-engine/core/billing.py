"""billing-service 内部接口客户端（平台托管计费统一入口）。

业务约定（2026-08-20 收敛）：
  - billing-service 是唯一商业化计费中枢（余额/资源包/用量明细）。
  - 对话 / 工作流均在 ai-engine 侧调用本模块完成「余额校验 + 扣费」。
  - 租户配置了自有 API Key 时不走平台托管计费（由调用方判断）。
"""
import logging

import httpx

from config.settings import settings

logger = logging.getLogger(__name__)


class BillingQuotaExceeded(Exception):
    """余额不足异常，message 为用户可见提示"""
    pass


def _base_url() -> str:
    return getattr(settings, "billing_service_url", "") or "http://localhost:9600"


def _headers() -> dict:
    token = getattr(settings, "billing_internal_token", "") or ""
    h = {"Content-Type": "application/json"}
    if token:
        h["X-Internal-Token"] = token
    return h


async def check_billing_quota(tenant_id: str) -> dict:
    """调用 LLM 前校验租户余额；余额不足抛 BillingQuotaExceeded。

    判断口径：资源包剩余 token 与余额（金额）同时为 0 时拒绝。
    计费服务不可用时放行（尽力而为，避免阻断业务）。
    """
    try:
        async with httpx.AsyncClient(timeout=5) as client:
            resp = await client.get(
                f"{_base_url()}/billing/internal/balance",
                params={"tenant_id": tenant_id},
                headers=_headers(),
            )
    except Exception as e:
        logger.warning("billing balance check error: %s", e)
        return {"available": True}
    if resp.status_code != 200:
        logger.warning("billing balance check failed: %s %s", resp.status_code, resp.text[:200])
        return {"code": resp.status_code, "available": True}
    data = (resp.json().get("data") or {}) if resp.status_code == 200 else {}
    remaining_tokens = int(data.get("remaining_tokens") or 0)
    bal = (data.get("balance") or {}).get("balance") or 0
    if remaining_tokens <= 0 and float(bal) <= 0:
        raise BillingQuotaExceeded("账户余额不足，请充值后继续使用 AI 服务。")
    return data


async def consume_billing(tenant_id: str, model_id: str, input_tokens: int, output_tokens: int) -> bool:
    """LLM 成功后扣费（billing-service internal，FIFO 资源包 + 余额）。"""
    try:
        async with httpx.AsyncClient(timeout=5) as client:
            resp = await client.post(
                f"{_base_url()}/billing/internal/deduct-ai-cost",
                json={
                    "tenant_id": tenant_id,
                    "model_id": model_id or "",
                    "input_tokens": int(input_tokens or 0),
                    "output_tokens": int(output_tokens or 0),
                },
                headers=_headers(),
            )
        if resp.status_code != 200:
            logger.warning("billing deduct failed: %s %s", resp.status_code, resp.text[:200])
            return False
        return (resp.json().get("code") or -1) == 0
    except Exception as e:
        logger.warning("billing deduct error: %s", e)
        return False
