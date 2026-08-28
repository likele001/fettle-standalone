"""AI Engine 内部服务鉴权中间件。

ai-engine 只允许被网关与本机可信服务调用。
除健康检查等白名单路径外，所有业务端点必须携带 X-Internal-Token，
令牌值从环境变量 AI_ENGINE_INTERNAL_TOKEN 读取（与网关/agent-service 共享），
并兼容通过 .env 文件注入。
"""
import logging
import os

from fastapi import Request
from fastapi.responses import JSONResponse

logger = logging.getLogger(__name__)


def _load_token() -> str:
    """优先环境变量，其次 .env 文件。"""
    token = os.getenv("AI_ENGINE_INTERNAL_TOKEN", "").strip()
    if token:
        return token

    # 兼容从 .env 读取（与 config/settings.py 的 env_file 目录一致）
    try:
        from dotenv import load_dotenv
        base = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))  # ai-engine/
        load_dotenv(os.path.join(base, ".env"))
        token = os.getenv("AI_ENGINE_INTERNAL_TOKEN", "").strip()
    except Exception:
        token = ""

    return token


INTERNAL_TOKEN = _load_token()

# 白名单：不需要内部令牌即可访问（健康探针/根节点）
PUBLIC_PATHS = {"/", "/health", "/ready", "/docs", "/openapi.json", "/redoc"}

if not INTERNAL_TOKEN:
    logger.warning("AI_ENGINE_INTERNAL_TOKEN 未配置，内部鉴权被禁用！请在生产环境注入该环境变量。")


async def internal_auth_middleware(request: Request, call_next):
    """校验 X-Internal-Token，失败返回 401。"""
    path = request.url.path

    # 放行白名单
    if path in PUBLIC_PATHS or path.startswith("/api/v1/health") or path.startswith("/health"):
        return await call_next(request)

    # 未配置令牌时：返回 503 拒绝一切业务请求，宁可失败也不裸奔
    if not INTERNAL_TOKEN:
        return JSONResponse(
            status_code=503,
            content={"code": 1006, "message": "ai-engine internal auth not configured"},
        )

    token = request.headers.get("X-Internal-Token", "")
    if token != INTERNAL_TOKEN:
        logger.warning("ai-engine internal auth rejected path=%s", path)
        return JSONResponse(
            status_code=401,
            content={"code": 1002, "message": "invalid internal token"},
        )

    return await call_next(request)