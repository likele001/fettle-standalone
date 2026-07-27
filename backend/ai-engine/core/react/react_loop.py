"""ReAct 循环 - 使用 LangGraph 工作流引擎"""
import json
import logging
import time
from typing import List, Dict, Any, Optional
from dataclasses import dataclass, field
from uuid import uuid4

from core.workflow import WorkflowBuilder, workflow_engine, Workflow, WorkflowInstance
from core.models.model_router import model_router, TaskType
from core.models.providers import ProviderFactory, ChatMessage

logger = logging.getLogger(__name__)


@dataclass
class Task:
    """任务定义"""
    id: str
    tenant_id: str
    user_id: str
    agent_id: str
    intent: str
    entities: Dict[str, str] = field(default_factory=dict)
    goal: str = ""
    available_tools: List[str] = field(default_factory=list)
    context: Dict[str, str] = field(default_factory=dict)


@dataclass
class Step:
    """执行步骤"""
    step_num: int = 0
    thought: str = ""
    action: str = ""
    action_input: Dict[str, str] = field(default_factory=dict)
    observation: str = ""
    is_final: bool = False
    final_answer: str = ""


@dataclass
class TaskResult:
    """任务执行结果"""
    success: bool = False
    steps: List[Step] = field(default_factory=list)
    final_answer: str = ""
    total_tokens: int = 0
    duration: float = 0.0


class ReactLoop:
    """ReAct 循环核心 - 使用 LangGraph"""

    def __init__(self, max_steps: int = 15, timeout: int = 300):
        self.max_steps = max_steps
        self.timeout = timeout

    async def execute(self, task: Task) -> TaskResult:
        """执行 ReAct 循环 - 使用 LangGraph 工作流"""
        start_time = time.time()
        result = TaskResult()

        try:
            workflow = await self._build_react_workflow(task)
            
            instance = WorkflowInstance(
                instance_id=f"task_{uuid4().hex[:8]}",
                workflow_id=workflow.id,
                tenant_id=task.tenant_id,
                user_id=task.user_id,
                input_data={
                    "input": task.goal,
                    "intent": task.intent,
                    "entities": task.entities,
                    "context": task.context
                }
            )

            workflow_result = await workflow_engine.execute(workflow, instance)

            if workflow_result.success:
                result.success = True
                result.final_answer = str(workflow_result.output.get("output", {}).get("text", "") or workflow_result.output.get("context", {}).get("response", ""))
            else:
                result.final_answer = f"执行失败: {workflow_result.error}"

            result.duration = time.time() - start_time

        except Exception as e:
            logger.error(f"ReAct execution failed: {e}")
            result.final_answer = f"执行失败: {str(e)}"
            result.duration = time.time() - start_time

        return result

    async def _build_react_workflow(self, task: Task) -> Workflow:
        """构建 ReAct 工作流"""
        builder = WorkflowBuilder()

        llm_node_id = builder.add_llm_node(
            name="ReAct Agent",
            system_prompt=self._build_react_prompt(task),
            max_tokens=2000,
            temperature=0.7,
            output_key="response"
        )

        output_node_id = builder.add_text_output_node(
            name="输出结果",
            template="{response}"
        )

        end_node_id = builder.add_end_node()

        builder.set_start(llm_node_id)
        builder.connect(llm_node_id, output_node_id)
        builder.connect(output_node_id, end_node_id)

        workflow = builder.build(
            workflow_id=f"react_{task.id}",
            name=f"ReAct Workflow - {task.intent}",
            description="基于 ReAct 的任务执行工作流",
            tenant_id=task.tenant_id
        )

        return workflow

    def _build_react_prompt(self, task: Task) -> str:
        """构建 ReAct 提示词"""
        tools_str = ", ".join(task.available_tools) if task.available_tools else "none"
        
        return f"""你是一个智能任务执行助手。请按照 ReAct 框架进行思考和行动。

用户目标: {task.goal}
意图: {task.intent}
实体: {task.entities}
可用工具: {tools_str}
上下文: {task.context}

请直接给出最终答案或执行计划。"""


# 全局实例
react_loop = ReactLoop()