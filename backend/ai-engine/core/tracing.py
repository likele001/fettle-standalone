"""LLM 调用链可观测性：结构化 JSONL trace（按租户/日期分文件）。

- 记录每次对话/工作流 LLM 调用的 token、耗时、模型、状态
- 供监控页/运维回溯（data/traces/<tenant>/<date>.jsonl）
- 同步接口，内部捕获异常，绝不阻断业务
"""
import json
import logging
import os
import time

logger = logging.getLogger(__name__)

_TRACE_DIR = os.getenv("AI_ENGINE_TRACE_DIR", "data/traces")


def _path(tenant_id: str) -> str:
    d = os.path.join(_TRACE_DIR, tenant_id)
    try:
        os.makedirs(d, exist_ok=True)
    except Exception:
        pass
    return os.path.join(d, time.strftime("%Y-%m-%d") + ".jsonl")


def log_trace(tenant_id: str, trace_type: str, model: str, input_tokens: int,
              output_tokens: int, latency_ms: float, status: str, extra: dict = None) -> None:
    """写一条调用链记录（尽力而为）。"""
    try:
        rec = {
            "ts": round(time.time(), 3),
            "type": trace_type,
            "tenant_id": tenant_id,
            "model": model or "",
            "input_tokens": int(input_tokens or 0),
            "output_tokens": int(output_tokens or 0),
            "latency_ms": round(latency_ms or 0, 1),
            "status": status,
            "extra": extra or {},
        }
        with open(_path(tenant_id), "a", encoding="utf-8") as f:
            f.write(json.dumps(rec, ensure_ascii=False) + "\n")
    except Exception as e:
        logger.warning("trace log failed: %s", e)
