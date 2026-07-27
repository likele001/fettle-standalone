"""MCP 服务端"""
import logging
import json
from typing import Dict, Any, List
from fastapi import FastAPI, HTTPException, Request

from .tool_registry import tool_registry
from .mcp_protocol import MCPProtocol

logger = logging.getLogger(__name__)


class MCPServer:
    """MCP 服务端"""

    def __init__(self, name: str = "Fettle MCP Server", version: str = "1.0.0", description: str = ""):
        self.name = name
        self.version = version
        self.description = description
        self.app = FastAPI(title=name, version=version)
        self._setup_routes()

    def _setup_routes(self):
        """设置路由"""
        @self.app.post("/mcp")
        async def mcp_handler(request: Request):
            try:
                body = await request.json()
                method = body.get("method", "")
                params = body.get("params", {})

                if method == "mcp/describe":
                    return await self._handle_describe()
                elif method == "mcp/execute":
                    return await self._handle_execute(params)
                else:
                    return MCPProtocol.build_error_response(f"Unknown method: {method}")

            except Exception as e:
                logger.error(f"MCP handler error: {e}")
                return MCPProtocol.build_error_response(str(e))

        @self.app.get("/health")
        async def health_check():
            return {"status": "healthy", "name": self.name, "version": self.version}

    async def _handle_describe(self):
        """处理描述请求"""
        tools = tool_registry.list_tools()
        return MCPProtocol.build_describe_response(
            name=self.name,
            version=self.version,
            description=self.description,
            tools=tools
        )

    async def _handle_execute(self, params: Dict[str, Any]):
        """处理执行请求"""
        tool_name = params.get("name")
        arguments = params.get("arguments", {})

        if not tool_name:
            return MCPProtocol.build_error_response("Tool name is required")

        try:
            result = await tool_registry.execute_tool(tool_name, **arguments)
            return MCPProtocol.build_execute_response(str(result), is_error=False)

        except Exception as e:
            logger.error(f"Tool execution failed: {tool_name}: {e}")
            return MCPProtocol.build_execute_response(str(e), is_error=True, error_type="execution_error")

    def run(self, host: str = "0.0.0.0", port: int = 8000):
        """启动服务"""
        import uvicorn
        logger.info(f"Starting MCP server at {host}:{port}")
        uvicorn.run(self.app, host=host, port=port)

    async def run_async(self, host: str = "0.0.0.0", port: int = 8000):
        """异步启动服务"""
        import uvicorn
        logger.info(f"Starting MCP server at {host}:{port}")
        config = uvicorn.Config(self.app, host=host, port=port)
        server = uvicorn.Server(config)
        await server.serve()