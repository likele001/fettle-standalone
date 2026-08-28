"""MCP 客户端"""
import logging
import httpx
from typing import Dict, Any, List, Optional

from .mcp_protocol import MCPProtocol, MCPServerInfo

logger = logging.getLogger(__name__)


class MCPClient:
    """MCP 客户端"""

    def __init__(self, server_url: str, timeout: int = 30):
        self.server_url = server_url
        self.timeout = timeout
        self.client = httpx.AsyncClient(timeout=timeout)
        self.server_info: Optional[MCPServerInfo] = None
        self._tools_cache: Optional[List[Dict[str, Any]]] = None

    async def describe(self) -> Optional[MCPServerInfo]:
        """获取服务器描述"""
        try:
            request = MCPProtocol.build_describe_request()
            response = await self.client.post(
                self.server_url,
                content=request,
                headers={"Content-Type": "application/json"}
            )

            result = MCPProtocol.parse_response(response.text)
            if result:
                self.server_info = MCPServerInfo(
                    name=result.get("name", ""),
                    version=result.get("version", ""),
                    description=result.get("description", ""),
                    tools=result.get("tools", []),
                    url=self.server_url
                )
                self._tools_cache = result.get("tools", [])
                return self.server_info

        except Exception as e:
            logger.error(f"MCP describe failed: {e}")

        return None

    async def execute(self, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """执行工具"""
        try:
            request = MCPProtocol.build_execute_request(tool_name, arguments)
            response = await self.client.post(
                self.server_url,
                content=request,
                headers={"Content-Type": "application/json"}
            )

            result = MCPProtocol.parse_response(response.text)
            if result:
                return {
                    "status": "success",
                    "content": result.get("content", ""),
                    "is_error": result.get("is_error", False),
                    "error_type": result.get("error_type", "")
                }

            return {"status": "error", "error": "Invalid response"}

        except Exception as e:
            logger.error(f"MCP execute failed: {e}")
            return {"status": "error", "error": str(e)}

    async def list_tools(self) -> List[Dict[str, Any]]:
        """列出所有工具"""
        if self._tools_cache is None:
            await self.describe()

        return self._tools_cache or []

    async def get_tool(self, tool_name: str) -> Optional[Dict[str, Any]]:
        """获取工具信息"""
        tools = await self.list_tools()
        for tool in tools:
            if tool.get("name") == tool_name:
                return tool
        return None

    async def health_check(self) -> str:
        """健康检查"""
        try:
            info = await self.describe()
            return "healthy" if info else "unhealthy"
        except Exception:
            return "unhealthy"

    async def close(self):
        """关闭连接"""
        await self.client.aclose()


class MCPToolManager:
    """MCP 工具管理器"""

    def __init__(self):
        self.servers: Dict[str, MCPClient] = {}

    async def register_server(self, server_id: str, server_url: str) -> bool:
        """注册 MCP 服务器"""
        try:
            client = MCPClient(server_url)
            info = await client.describe()
            if info:
                self.servers[server_id] = client
                logger.info(f"Registered MCP server: {server_id} -> {server_url}")
                return True
            return False
        except Exception as e:
            logger.error(f"Failed to register MCP server: {e}")
            return False

    async def unregister_server(self, server_id: str):
        """注销 MCP 服务器"""
        if server_id in self.servers:
            await self.servers[server_id].close()
            del self.servers[server_id]
            logger.info(f"Unregistered MCP server: {server_id}")

    async def execute_tool(self, server_id: str, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """执行 MCP 工具"""
        if server_id not in self.servers:
            return {"status": "error", "error": f"MCP server not found: {server_id}"}

        client = self.servers[server_id]
        return await client.execute(tool_name, arguments)

    async def list_all_tools(self) -> List[Dict[str, Any]]:
        """列出所有 MCP 工具"""
        all_tools = []
        for server_id, client in self.servers.items():
            tools = await client.list_tools()
            for tool in tools:
                tool["server_id"] = server_id
                all_tools.append(tool)
        return all_tools

    async def get_server_info(self, server_id: str) -> Optional[MCPServerInfo]:
        """获取服务器信息"""
        if server_id in self.servers:
            return await self.servers[server_id].describe()
        return None

    async def health_check_all(self) -> Dict[str, str]:
        """检查所有服务器健康状态"""
        results = {}
        for server_id, client in self.servers.items():
            results[server_id] = await client.health_check()
        return results

    async def register_official_server(self, server_id: str, type_: str = "stdio", command: str = "",
                                       args: Optional[List[str]] = None, env: Optional[Dict[str, str]] = None,
                                       url: str = "") -> bool:
        """注册官方 MCP 服务器（stdio/SSE）。"""
        try:
            client = OfficialMCPClient(server_id, type_=type_, command=command, args=args, env=env, url=url)
            if await client.health_check() == "healthy":
                self.servers[server_id] = client
                logger.info(f"Registered official MCP server: {server_id} type={type_}")
                return True
            await client.close()
            return False
        except Exception as e:
            logger.error(f"Failed to register official MCP server: {e}")
            return False



class OfficialMCPClient:
    """官方 MCP 协议客户端（stdio / SSE），兼容 Anthropic MCP 规范。

    与自研 MCPClient 保持相同接口（describe/list_tools/execute/health_check/close），
    可被 MCPToolManager 混用（server 字典按 protocol 区分）。
    """

    def __init__(self, server_id: str, type_: str = "stdio", command: str = "",
                 args: Optional[List[str]] = None, env: Optional[Dict[str, str]] = None,
                 url: str = "", timeout: int = 60):
        self.server_id = server_id
        self.type = type_
        self.command = command
        self.args = args or []
        self.env = env
        self.url = url
        self.timeout = timeout
        self._session = None
        self._stack = None

    async def _ensure_connected(self):
        if self._session is not None:
            return
        from contextlib import AsyncExitStack
        from mcp import ClientSession
        stack = AsyncExitStack()
        if self.type == "sse":
            from mcp.client.sse import sse_client
            read, write = await stack.enter_async_context(sse_client(self.url))
        else:
            from mcp import StdioServerParameters
            from mcp.client.stdio import stdio_client
            params = StdioServerParameters(command=self.command, args=self.args, env=self.env)
            read, write = await stack.enter_async_context(stdio_client(params))
        session = await stack.enter_async_context(ClientSession(read, write))
        await session.initialize()
        self._session = session
        self._stack = stack

    async def describe(self) -> Optional[dict]:
        """兼容接口：返回服务器信息（官方协议无统一 describe，返回 tool 列表摘要）"""
        try:
            tools = await self.list_tools()
            return {"name": self.server_id, "version": "1.0", "description": "", "tools": tools, "url": self.url or self.command}
        except Exception as e:
            logger.error(f"official MCP describe failed: {e}")
            return None

    async def list_tools(self) -> List[Dict[str, Any]]:
        await self._ensure_connected()
        res = await self._session.list_tools()
        tools = []
        for t in res.tools:
            schema = getattr(t, "inputSchema", None) or {}
            tools.append({
                "name": t.name,
                "description": t.description or "",
                "inputSchema": schema,
                "protocol": "mcp-official",
            })
        return tools

    async def get_tool(self, tool_name: str) -> Optional[Dict[str, Any]]:
        for t in await self.list_tools():
            if t.get("name") == tool_name:
                return t
        return None

    async def execute(self, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        await self._ensure_connected()
        try:
            res = await self._session.call_tool(tool_name, arguments or {})
            contents = []
            for c in (res.content or []):
                text = getattr(c, "text", None)
                contents.append(text if text is not None else str(c))
            return {
                "status": "success",
                "content": "\n".join(contents),
                "is_error": bool(getattr(res, "isError", False)),
            }
        except Exception as e:
            logger.error(f"official MCP execute failed: {e}")
            return {"status": "error", "error": str(e)}

    async def health_check(self) -> str:
        try:
            await self._ensure_connected()
            return "healthy"
        except Exception as e:
            logger.warning(f"official MCP health check failed: {e}")
            return "unhealthy"

    async def close(self):
        if self._stack is not None:
            try:
                await self._stack.aclose()
            except Exception:
                pass
            self._stack = None
            self._session = None



# 全局 MCP 工具管理器
mcp_tool_manager = MCPToolManager()