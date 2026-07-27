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


# 全局 MCP 工具管理器
mcp_tool_manager = MCPToolManager()