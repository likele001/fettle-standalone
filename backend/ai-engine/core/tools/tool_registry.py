"""工具注册中心 - 支持 MCP 工具"""
import logging
from typing import Dict, Any, List, Optional, Callable
from dataclasses import dataclass
from abc import ABC, abstractmethod

from .mcp_client import mcp_tool_manager

logger = logging.getLogger(__name__)


@dataclass
class ToolDefinition:
    """工具定义"""
    name: str
    description: str
    parameters: Dict[str, Any]
    handler: Callable
    category: str = "general"
    is_mcp: bool = False
    server_id: str = ""


class BaseTool(ABC):
    """工具基类"""
    
    @property
    @abstractmethod
    def name(self) -> str:
        """工具名称"""
        pass
    
    @property
    @abstractmethod
    def description(self) -> str:
        """工具描述"""
        pass
    
    @property
    @abstractmethod
    def parameters(self) -> Dict[str, Any]:
        """参数定义"""
        pass
    
    @abstractmethod
    async def execute(self, **kwargs) -> Any:
        """执行工具"""
        pass


class ToolRegistry:
    """工具注册中心"""
    
    def __init__(self):
        self.tools: Dict[str, ToolDefinition] = {}
        self.categories: Dict[str, List[str]] = {}
    
    def register(
        self,
        name: str,
        description: str,
        parameters: Dict[str, Any],
        handler: Callable,
        category: str = "general",
        is_mcp: bool = False,
        server_id: str = ""
    ):
        """注册工具"""
        tool = ToolDefinition(
            name=name,
            description=description,
            parameters=parameters,
            handler=handler,
            category=category,
            is_mcp=is_mcp,
            server_id=server_id
        )
        
        self.tools[name] = tool
        
        if category not in self.categories:
            self.categories[category] = []
        self.categories[category].append(name)
        
        logger.info(f"Registered tool: {name} in category: {category} (MCP: {is_mcp})")
    
    def register_tool_class(self, tool_class: type):
        """注册工具类"""
        instance = tool_class()
        self.register(
            name=instance.name,
            description=instance.description,
            parameters=instance.parameters,
            handler=instance.execute,
            category=getattr(instance, 'category', 'general')
        )
    
    def register_mcp_tool(self, server_id: str, tool_info: Dict[str, Any]):
        """注册 MCP 工具"""
        async def mcp_handler(**kwargs):
            result = await mcp_tool_manager.execute_tool(server_id, tool_info['name'], kwargs)
            if result.get("status") == "success":
                return result.get("content", "")
            return f"MCP Error: {result.get('error', '')}"

        self.register(
            name=f"mcp_{server_id}_{tool_info['name']}",
            description=tool_info.get("description", ""),
            parameters=tool_info.get("parameters", {}),
            handler=mcp_handler,
            category=f"mcp_{server_id}",
            is_mcp=True,
            server_id=server_id
        )
    
    def get_tool(self, name: str) -> Optional[ToolDefinition]:
        """获取工具"""
        return self.tools.get(name)
    
    def list_tools(self, category: str = None) -> List[Dict[str, Any]]:
        """列出工具"""
        if category:
            tool_names = self.categories.get(category, [])
            return [
                {
                    "name": name,
                    "description": self.tools[name].description,
                    "parameters": self.tools[name].parameters,
                    "category": self.tools[name].category,
                    "is_mcp": self.tools[name].is_mcp
                }
                for name in tool_names
                if name in self.tools
            ]
        
        return [
            {
                "name": tool.name,
                "description": tool.description,
                "parameters": tool.parameters,
                "category": tool.category,
                "is_mcp": tool.is_mcp,
                "server_id": tool.server_id
            }
            for tool in self.tools.values()
        ]
    
    def list_categories(self) -> List[str]:
        """列出所有类别"""
        return list(self.categories.keys())
    
    async def execute_tool(self, name: str, **kwargs) -> Any:
        """执行工具"""
        tool = self.get_tool(name)
        if not tool:
            raise ValueError(f"Tool not found: {name}")
        
        try:
            result = await tool.handler(**kwargs)
            logger.info(f"Executed tool: {name}")
            return result
        except Exception as e:
            logger.error(f"Tool execution failed: {name}: {e}")
            raise


# 全局工具注册中心
tool_registry = ToolRegistry()