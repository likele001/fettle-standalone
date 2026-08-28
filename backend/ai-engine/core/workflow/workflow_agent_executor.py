"""AGENT 节点执行器：调用子 Agent（agent-service 内部端点），回复写入 context。

多 Agent 协作（方案 A）：工作流 = 协作图，AGENT 节点 = 专职岗位。
子 Agent 自身的模型/知识库/工作流由其配置生效（agent-service TestChat 逻辑）。
"""
import logging
from typing import Any, Dict

import httpx

from config.settings import settings

from .workflow_engine import _render_template, _flatten_state

logger = logging.getLogger(__name__)


def _agent_url() -> str:
    return getattr(settings, "agent_service_url", "") or "http://localhost:9300"


def _internal_token() -> str:
    return getattr(settings, "billing_internal_token", "") or ""


async def execute_agent_node(config: Dict[str, Any], state: Dict[str, Any],
                             tenant_id: str = "", user_id: str = "") -> Dict[str, Any]:
    """执行 AGENT 节点：调用子 Agent 对话，回复写入 state.context[output_key]。

    config: {agent_id, input_template, output_key, timeout}
    """
    output_key = config.get("output_key") or "agent_reply"
    agent_id = (config.get("agent_id") or "").strip()
    if not agent_id:
        state["context"][output_key] = "[AGENT节点] 未配置 agent_id"
        state["context"][output_key + "_ok"] = False
        return state

    # 渲染输入模板（state 上下文）
    template = config.get("input_template") or "{input}"
    try:
        message = _render_template(template, _flatten_state(state))
    except Exception:
        message = template

    t_id = tenant_id or (state.get("context") or {}).get("tenant_id") or "default"
    u_id = user_id or (state.get("context") or {}).get("user_id") or ""

    timeout = int(config.get("timeout") or 60)
    headers = {"Content-Type": "application/json"}
    token = _internal_token()
    if token:
        headers["X-Internal-Token"] = token

    try:
        async with httpx.AsyncClient(timeout=timeout) as client:
            resp = await client.post(
                f"{_agent_url()}/internal/agents/test",
                json={
                    "tenant_id": t_id,
                    "user_id": u_id,
                    "agent_id": agent_id,
                    "message": message,
                },
                headers=headers,
            )
        if resp.status_code == 200:
            data = resp.json()
            reply = ((data.get("data") or {}).get("reply") or "").strip()
            state["context"][output_key] = reply or "[AGENT节点] 子 Agent 无回复"
            state["context"][output_key + "_ok"] = bool(reply)
        else:
            state["context"][output_key] = f"[AGENT节点] 子 Agent 调用失败({resp.status_code})"
            state["context"][output_key + "_ok"] = False
    except Exception as e:
        logger.warning("agent node failed: %s", e)
        state["context"][output_key] = f"[AGENT节点] 子 Agent 调用异常：{str(e)[:200]}"
        state["context"][output_key + "_ok"] = False
    return state
