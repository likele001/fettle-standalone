"""MCP 协议实现"""
import logging
import json
from typing import Dict, Any, List, Optional
from dataclasses import dataclass, field
from enum import Enum

logger = logging.getLogger(__name__)


class MCPRequestType(str, Enum):
    """MCP 请求类型"""
    DESCRIBE = "describe"
    EXECUTE = "execute"
    SHUTDOWN = "shutdown"


@dataclass
class MCPDescribeRequest:
    """MCP 描述请求"""
    type: str = "MCPDescribeRequest"


@dataclass
class MCPExecuteRequest:
    """MCP 执行请求"""
    type: str = "MCPExecuteRequest"
    name: str = ""
    arguments: Dict[str, Any] = field(default_factory=dict)


@dataclass
class MCPDescribeResponse:
    """MCP 描述响应"""
    name: str = ""
    version: str = ""
    description: str = ""
    tools: List[Dict[str, Any]] = field(default_factory=list)


@dataclass
class MCPExecuteResponse:
    """MCP 执行响应"""
    type: str = ""
    content: str = ""
    is_error: bool = False
    error_type: str = ""


@dataclass
class MCPServerInfo:
    """MCP 服务器信息"""
    id: str = ""
    name: str = ""
    url: str = ""
    description: str = ""
    health: str = "unknown"
    tools: List[Dict[str, Any]] = field(default_factory=list)
    created_at: str = ""


class MCPProtocol:
    """MCP 协议处理"""

    @staticmethod
    def build_describe_request() -> str:
        """构建描述请求"""
        return json.dumps({
            "jsonrpc": "2.0",
            "id": 1,
            "method": "mcp/describe",
            "params": {}
        })

    @staticmethod
    def build_execute_request(tool_name: str, arguments: Dict[str, Any]) -> str:
        """构建执行请求"""
        return json.dumps({
            "jsonrpc": "2.0",
            "id": 1,
            "method": "mcp/execute",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        })

    @staticmethod
    def parse_response(response: str) -> Optional[Dict[str, Any]]:
        """解析响应"""
        try:
            data = json.loads(response)
            if "error" in data:
                return {"error": data["error"]}
            return data.get("result")
        except json.JSONDecodeError:
            logger.error(f"Failed to parse MCP response: {response}")
            return None

    @staticmethod
    def build_describe_response(
        name: str,
        version: str,
        description: str,
        tools: List[Dict[str, Any]]
    ) -> str:
        """构建描述响应"""
        return json.dumps({
            "jsonrpc": "2.0",
            "id": 1,
            "result": {
                "name": name,
                "version": version,
                "description": description,
                "tools": tools
            }
        })

    @staticmethod
    def build_execute_response(content: str, is_error: bool = False, error_type: str = "") -> str:
        """构建执行响应"""
        result = {
            "content": content,
            "is_error": is_error
        }
        if is_error:
            result["error_type"] = error_type

        return json.dumps({
            "jsonrpc": "2.0",
            "id": 1,
            "result": result
        })

    @staticmethod
    def build_error_response(message: str, code: int = -1) -> str:
        """构建错误响应"""
        return json.dumps({
            "jsonrpc": "2.0",
            "id": 1,
            "error": {
                "code": code,
                "message": message
            }
        })