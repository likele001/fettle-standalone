"""skill-service 执行桥接：租户已安装技能注册为可执行工具。

- Code 技能：临时脚本 + subprocess 沙箱执行（www 用户权限 + 超时保护）
- 约定：技能 Code 为 main(params) 函数体（4 空格缩进）
- 调用：对话 / 工作流入口 ensure_tenant_skills(tenant_id)（缓存 60s，失败不阻断）
"""
import json
import logging
import os
import shutil
import signal
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

# 技能代码是租户可控的任意 Python，与 ai-engine 同 uid（www）运行，因此只能靠
# 资源上限 + 输出截断 + 进程组回收来兜底：单次执行不得吃满内存/CPU/磁盘，
# 超时后其派生的所有子进程必须一起回收。
# 已验证边界：setsid 脱离进程组的孤儿无法靠 killpg 回收，但它仍继承本次的 CPU 时间
# 与 nproc 上限，最长存活时间受 timeout 约束；environ 已过滤，但同 uid 仍可读
# /proc/<ai-engine>/environ（DB 口令等），彻底堵住需要独立的沙箱用户（见交付说明）。
# 注：本机 unprivileged user namespace 被禁（uid_map 写入 EPERM），无法用
# unshare 做网络隔离；www 也无 CAP_SETUID，无法再降权到独立用户。
_SKILL_MEM_MB = int(os.environ.get("SKILL_SANDBOX_MEM_MB", "2048"))
_SKILL_NPROC = int(os.environ.get("SKILL_SANDBOX_NPROC", "1024"))
_SKILL_FSIZE_BLOCKS = int(os.environ.get("SKILL_SANDBOX_FSIZE_BLOCKS", "204800"))  # ~100MB
_SKILL_OUTPUT_CAP = 64 * 1024  # 只回读这么多字节，其余丢弃


def _sandbox_shell(cpu_sec: int) -> str:
    """构造 ulimit 包装。用 shell 而非 preexec_fn：后者在多线程进程里 fork 不安全。"""
    return (
        "ulimit -v {mem} -t {cpu} -u {nproc} -f {fsize} -c 0 || exit 97\n"
        'exec "$@"\n'
    ).format(mem=_SKILL_MEM_MB * 1024, cpu=cpu_sec, nproc=_SKILL_NPROC, fsize=_SKILL_FSIZE_BLOCKS)


def _read_capped(path: str) -> str:
    try:
        size = os.path.getsize(path)
        with open(path, "r", encoding="utf-8", errors="replace") as f:
            if size > _SKILL_OUTPUT_CAP:
                f.seek(size - _SKILL_OUTPUT_CAP)
            return f.read().strip()
    except OSError:
        return ""


def _kill_group(proc: subprocess.Popen) -> None:
    """连坐回收：技能脚本可以再 fork，只杀直接子进程会留下跑满额度的孤儿。
    start_new_session=True 使 proc.pid 就是进程组 ID，因此 leader 已回收后仍可击杀存活后代。"""
    try:
        os.killpg(proc.pid, signal.SIGKILL)
    except OSError:
        pass
    try:
        proc.kill()
    except OSError:
        pass


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
    """受限执行：ulimit 资源上限 + 私有工作目录 + 进程组回收 + 输出截断。"""
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
    workdir = None
    proc = None
    try:
        workdir = tempfile.mkdtemp(prefix="skill-")
        tmp = os.path.join(workdir, "skill.py")
        with open(tmp, "w", encoding="utf-8") as f:
            f.write(script)

        out_path = os.path.join(workdir, "stdout")
        err_path = os.path.join(workdir, "stderr")
        # cwd/HOME/TMPDIR 全部指进私有目录：技能落地的临时文件随执行结束一起消失，
        # 不会写进项目目录被下一个租户读到。
        env = {
            "PATH": os.environ.get("PATH", "/usr/bin:/bin"),
            "LANG": "C.UTF-8",
            "LC_ALL": "C.UTF-8",
            "HOME": workdir,
            "TMPDIR": workdir,
        }
        try:
            with open(out_path, "wb") as out_f, open(err_path, "wb") as err_f:
                proc = subprocess.Popen(
                    ["/bin/sh", "-c", _sandbox_shell(timeout), "skill-sh", sys.executable, tmp],
                    stdin=subprocess.PIPE,
                    stdout=out_f,
                    stderr=err_f,
                    cwd=workdir,
                    env=env,
                    start_new_session=True,
                )
                try:
                    proc.communicate(json.dumps(params or {}, ensure_ascii=False).encode("utf-8"),
                                     timeout=timeout)
                except subprocess.TimeoutExpired:
                    _kill_group(proc)
                    proc.wait(timeout=5)
                    return f"技能执行超时（>{timeout}s）"

            out = _read_capped(out_path)
            if "__ERROR__" in out:
                err = out.split("__ERROR__:", 1)[1].strip()
                logger.error("skill %s exec error: %s", name, err[:300])
                return f"技能执行出错：{err[-500:]}"
            if "__RESULT__:" in out:
                return out.split("__RESULT__:", 1)[1].strip()[:2000]
            if proc.returncode != 0 and not out:
                # 触发 ulimit 时内核直接杀进程，没有 traceback，只能回吐 stderr
                err_text = _read_capped(err_path)
                if err_text:
                    return f"技能执行出错：{err_text[-500:]}"
                return f"技能执行出错：进程异常退出（code={proc.returncode}），可能超出资源限制"
            return out[:2000]
        finally:
            # 正常返回也要收组：技能后台起的进程不能活到下一个租户的执行里。
            if proc is not None:
                _kill_group(proc)
    except Exception as e:
        logger.error("skill %s exec exception: %s", name, e)
        return f"技能执行异常：{e}"
    finally:
        if workdir:
            shutil.rmtree(workdir, ignore_errors=True)


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
