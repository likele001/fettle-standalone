"""MCP 服务器管理 API（官方 MCP 兼容：stdio / SSE；兼容自研 HTTP 协议）"""
import json
import logging
import os
import uuid
from typing import Any, Dict, List, Optional

from fastapi import APIRouter
from pydantic import BaseModel

from core.tools.mcp_client import mcp_tool_manager

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/mcp-servers", tags=["mcp-servers"])

STORE = os.path.join(os.getenv("AI_ENGINE_DATA_DIR", "data"), "mcp_servers.json")


def _load() -> Dict[str, dict]:
    try:
        with open(STORE, encoding="utf-8") as f:
            return json.load(f)
    except Exception:
        return {}


def _save(data: Dict[str, dict]):
    os.makedirs(os.path.dirname(STORE), exist_ok=True)
    tmp = STORE + ".tmp"
    with open(tmp, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    os.replace(tmp, STORE)


class MCPServerCreate(BaseModel):
    name: str
    type: str = "stdio"  # stdio | sse | legacy(http)
    command: Optional[str] = ""
    args: Optional[List[str]] = []
    env: Optional[Dict[str, str]] = {}
    url: Optional[str] = ""
    description: Optional[str] = ""


class MCPServerUpdate(BaseModel):
    name: Optional[str] = None
    type: Optional[str] = None
    command: Optional[str] = None
    args: Optional[List[str]] = None
    env: Optional[Dict[str, str]] = None
    url: Optional[str] = None
    description: Optional[str] = None


def _register_runtime(sid: str, cfg: dict):
    """注册到运行时（失败仅告警，可后续 test 时重连）。"""
    try:
        if cfg.get("type") in ("stdio", "sse"):
            return mcp_tool_manager.register_official_server(
                sid, type_=cfg["type"], command=cfg.get("command") or "",
                args=cfg.get("args") or [], env=cfg.get("env") or {},
                url=cfg.get("url") or "",
            )
        elif cfg.get("type") == "legacy":
            from core.tools.mcp_client import MCPClient
            client = MCPClient(cfg.get("url") or "")
            info = await_legacy_describe(client)
            if info:
                mcp_tool_manager.servers[sid] = client
                return True
    except Exception as e:
        logger.warning("mcp server runtime register deferred: %s", e)
    return False


async def await_legacy_describe(client):
    info = await client.describe()
    return info


@router.get("")
async def list_servers():
    data = _load()
    items = [{"id": sid, **cfg} for sid, cfg in data.items()]
    return {"code": 0, "message": "success", "data": items}


@router.post("")
async def create_server(req: MCPServerCreate):
    data = _load()
    sid = uuid.uuid4().hex[:12]
    cfg = req.dict()
    cfg.setdefault("name", sid)
    data[sid] = cfg
    _save(data)
    await _register_runtime(sid, cfg)
    return {"code": 0, "message": "success", "data": {"id": sid, **cfg}}


@router.get("/{server_id}")
async def get_server(server_id: str):
    data = _load()
    cfg = data.get(server_id)
    if not cfg:
        return {"code": 1004, "message": "mcp server not found", "data": None}
    return {"code": 0, "message": "success", "data": {"id": server_id, **cfg}}


@router.put("/{server_id}")
async def update_server(server_id: str, req: MCPServerUpdate):
    data = _load()
    if server_id not in data:
        return {"code": 1004, "message": "mcp server not found", "data": None}
    merged = {**data[server_id], **{k: v for k, v in req.dict().items() if v is not None}}
    data[server_id] = merged
    _save(data)
    if server_id in mcp_tool_manager.servers:
        old = mcp_tool_manager.servers.pop(server_id)
        try:
            await old.close()
        except Exception:
            pass
    await _register_runtime(server_id, merged)
    return {"code": 0, "message": "success", "data": {"id": server_id, **merged}}


@router.delete("/{server_id}")
async def delete_server(server_id: str):
    data = _load()
    if server_id not in data:
        return {"code": 1004, "message": "mcp server not found", "data": None}
    del data[server_id]
    _save(data)
    if server_id in mcp_tool_manager.servers:
        old = mcp_tool_manager.servers.pop(server_id)
        try:
            await old.close()
        except Exception:
            pass
    return {"code": 0, "message": "success"}


@router.post("/{server_id}/test")
async def test_server(server_id: str):
    cfg = _load().get(server_id)
    if not cfg:
        return {"code": 1004, "message": "mcp server not found", "data": None}
    ok = await _register_runtime(server_id, cfg)
    tools = []
    if ok and server_id in mcp_tool_manager.servers:
        try:
            tools = await mcp_tool_manager.servers[server_id].list_tools()
        except Exception as e:
            logger.warning("list tools failed: %s", e)
    return {"code": 0 if ok else 5000, "message": "ok" if ok else "connection failed", "data": {"connected": ok, "tools": tools}}


@router.get("/{server_id}/tools")
async def list_server_tools(server_id: str):
    cfg = _load().get(server_id)
    if not cfg:
        return {"code": 1004, "message": "mcp server not found", "data": None}
    if server_id not in mcp_tool_manager.servers:
        await _register_runtime(server_id, cfg)
    tools = []
    if server_id in mcp_tool_manager.servers:
        try:
            tools = await mcp_tool_manager.servers[server_id].list_tools()
        except Exception as e:
            logger.warning("list tools failed: %s", e)
    return {"code": 0, "message": "success", "data": tools}
