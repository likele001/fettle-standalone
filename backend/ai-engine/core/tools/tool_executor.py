"""工具执行器"""
import logging
from typing import Dict, Any, List, Optional
import json

from .tool_registry import tool_registry

logger = logging.getLogger(__name__)


class ToolExecutor:
    """工具执行器"""
    
    def __init__(self, registry=None):
        self.registry = registry or tool_registry
    
    async def execute(
        self,
        tool_name: str,
        parameters: Dict[str, Any],
        context: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """执行工具"""
        try:
            logger.info(f"Executing tool: {tool_name} with params: {parameters}")
            
            # 合并上下文
            if context:
                parameters.update(context)
            
            # 执行工具
            result = await self.registry.execute_tool(tool_name, **parameters)
            
            return {
                "status": "success",
                "tool": tool_name,
                "result": result
            }
        
        except Exception as e:
            logger.error(f"Tool execution failed: {tool_name}: {e}")
            return {
                "status": "error",
                "tool": tool_name,
                "error": str(e)
            }
    
    async def execute_plan(
        self,
        steps: List[Dict[str, Any]],
        context: Optional[Dict[str, Any]] = None
    ) -> List[Dict[str, Any]]:
        """执行任务计划"""
        results = []
        current_context = context or {}
        
        for i, step in enumerate(steps):
            tool_name = step.get('tool')
            parameters = step.get('parameters', {})
            
            if not tool_name:
                logger.warning(f"Step {i} has no tool, skipping")
                continue
            
            # 执行工具
            result = await self.execute(tool_name, parameters, current_context)
            results.append(result)
            
            # 更新上下文
            if result['status'] == 'success':
                current_context[f'step_{i}_result'] = result['result']
            
            # 如果失败，根据策略决定是否继续
            if result['status'] == 'error':
                logger.error(f"Step {i} failed, stopping execution")
                break
        
        return results
    
    def parse_tool_call(self, llm_output: str) -> Optional[Dict[str, Any]]:
        """解析 LLM 输出的工具调用"""
        try:
            # 尝试解析 JSON 格式的工具调用
            data = json.loads(llm_output)
            
            if 'tool' in data or 'action' in data:
                return {
                    'tool': data.get('tool') or data.get('action'),
                    'parameters': data.get('parameters', {})
                }
        
        except json.JSONDecodeError:
            # 尝试从文本中提取工具调用
            # 简单实现：查找 "tool: xxx" 或 "action: xxx" 模式
            import re
            
            tool_match = re.search(r'(?:tool|action):\s*(\w+)', llm_output, re.IGNORECASE)
            if tool_match:
                tool_name = tool_match.group(1)
                
                # 尝试提取参数
                params = {}
                param_matches = re.findall(r'(\w+):\s*([^\n,]+)', llm_output)
                for key, value in param_matches:
                    if key.lower() not in ['tool', 'action']:
                        params[key] = value.strip()
                
                return {
                    'tool': tool_name,
                    'parameters': params
                }
        
        return None


# 全局工具执行器
tool_executor = ToolExecutor()
