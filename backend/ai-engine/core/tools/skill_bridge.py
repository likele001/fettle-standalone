"""skill-service 执行桥接：租户已安装技能注册为可执行工具。

- Code 技能：临时脚本 + subprocess 沙箱执行（www 用户权限 + 超时保护）
- 约定：技能 Code 为 main(params) 函数体（4 空格缩进）
- 调用：对话 / 工作流入口 ensure_tenant_skills(tenant_id)（缓存 60s，失败不阻断）
"""
import json
import logging
import os
import subprocess
import sys
import tempfile
import time
from typing import Any, Dict, List

import httpx

from config.settings import settings

from .tool_registry import tool_registry, ToolDefinition

logger = logging.getLogger(__name__)

_CACHE: Dict[str, float] = {}
_CACHE_TTL = 60


def _skill_url() -> str:
    return getattr(settings, "skill_service_url", "") or "http://localhost:9500"


def _headers() -> dict:
    token = getattr(settings, "billing_internal_token", "") or ""
    h = {"Content-Type": "application/json"}
    if token:
        h["X-Internal-Token"] = token
    return h


async def fetch_tenant_skills(tenant_id: str) -> List[dict]:
    """拉取租户已安装技能（含 Code）。"""
    try:
        async with httpx.AsyncClient(timeout=8) as client:
            resp = await client.get(
                f"{_skill_url()}/internal/skills/installed",
                params={"tenant_id": tenant_id},
                headers=_headers(),
            )
        if resp.status_code != 200:
            logger.warning("fetch skills failed: %s %s", resp.status_code, resp.text[:200])
            return []
        return resp.json().get("items") or []
    except Exception as e:
        logger.warning("fetch skills error: %s", e)
        return []


async def _exec_python_code(code: str, name: str, params: dict, timeout: int = 30) -> str:
    """受限沙箱：临时脚本 + subprocess（www 用户权限，无额外特权），超时保护。"""
    script = (
        "import sys, json, traceback\n"
        "def main(params):\n"
        + code + "\n"
        "try:\n"
        "    params = json.loads(sys.stdin.read() or '{}')\n"
        "    result = main(params)\n"
        "    print('__RESULT__:' + json.dumps(result, ensure_ascii=False, default=str))\n"
        "except Exception:\n"
        "    print('__ERROR__:' + traceback.format_exc())\n"
    )
    tmp = ""
    try:
        fd, tmp = tempfile.mkstemp(suffix=".py")
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            f.write(script)
        proc = subprocess.run(
            [sys.executable, tmp],
            input=json.dumps(params or {}, ensure_ascii=False),
            capture_output=True,
            text=True,
            timeout=timeout,
            env={k: v for k, v in os.environ.items() if k in ("PATH", "LANG", "LC_ALL", "TMPDIR")},
        )
        out = (proc.stdout or "").strip()
        if proc.returncode != 0 and not out:
            err_text = (proc.stderr or "").strip()
            if err_text:
                return f"技能执行出错：{err_text[-500:]}"
        if "__ERROR__" in out:
            err = out.split("__ERROR__:", 1)[1].strip()
            logger.error("skill %s exec error: %s", name, err[:300])
            return f"技能执行出错：{err[-500:]}"
        if "__RESULT__:" in out:
            return out.split("__RESULT__:", 1)[1].strip()[:2000]
        return out[:2000]
    except subprocess.TimeoutExpired:
        return f"技能执行超时（>{timeout}s）"
    except Exception as e:
        logger.error("skill %s exec exception: %s", name, e)
        return f"技能执行异常：{e}"
    finally:
        if tmp:
            try:
                os.unlink(tmp)
            except Exception:
                pass


def _build_handler(code: str, name: str):
    async def handler(**kwargs) -> str:
        return await _exec_python_code(code, name, kwargs)
    return handler


async def ensure_tenant_skills(tenant_id: str) -> int:
    """拉取租户已安装技能并注册为可执行工具（缓存 60s）。返回注册数。"""
    now = time.time()
    if tenant_id in _CACHE and now - _CACHE[tenant_id] < _CACHE_TTL:
        return 0
    _CACHE[tenant_id] = now
    skills = await fetch_tenant_skills(tenant_id)
    n = 0
    for sk in skills:
        code = (sk.get("code") or "").strip()
        if not code or sk.get("is_mcp_tool"):
            continue  # 本阶段仅桥接 Code 技能；MCP 类技能待 MCP server 配置完善
        name = sk.get("name") or sk.get("id") or "skill"
        tname = "skill_" + "".join(c for c in name.lower() if c.isalnum() or c == "_")[:40]
        if not tname or tname == "skill_":
            tname = "skill_" + (sk.get("id") or "x")[:8]
        if tool_registry.get_tool(tname):
            continue
        tool_registry.register(
            name=tname,
            description=sk.get("description") or f"技能：{name}",
            parameters={"type": "object", "properties": {}, "additionalProperties": True},
            handler=_build_handler(code, tname),
            category="skill",
        )
        n += 1
    if n:
        logger.info("registered %d skill tools for tenant %s", n, tenant_id)
    return n
