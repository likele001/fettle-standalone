"""工作流构建器"""
import logging
from typing import Dict, Any, List, Optional
from uuid import uuid4

from .node_types import (
    Workflow, Node, Edge, NodeType,
    LLMNodeConfig, ToolNodeConfig, RAGNodeConfig,
    ConditionNodeConfig, TextOutputNodeConfig
)

logger = logging.getLogger(__name__)


class WorkflowBuilder:
    """工作流构建器"""

    def __init__(self):
        self.nodes: List[Node] = []
        self.edges: List[Edge] = []
        self.start_node_id: str = ""
        self.end_node_id: str = ""

    def add_llm_node(
        self,
        name: str,
        model: str = "",
        system_prompt: str = "",
        max_tokens: int = 2000,
        temperature: float = 0.7,
        output_key: str = "response",
        position: Optional[Dict[str, float]] = None
    ) -> str:
        """添加 LLM 节点"""
        node_id = f"llm_{uuid4().hex[:8]}"
        config = LLMNodeConfig(
            node_type=NodeType.LLM,
            name=name,
            model=model,
            system_prompt=system_prompt,
            max_tokens=max_tokens,
            temperature=temperature,
            output_key=output_key
        )
        node = Node(
            id=node_id,
            type=NodeType.LLM,
            config=config,
            position=position or {"x": 0, "y": 0}
        )
        self.nodes.append(node)
        return node_id

    def add_tool_node(
        self,
        name: str,
        tool_name: str,
        parameters: Optional[Dict[str, Any]] = None,
        output_key: str = "tool_result",
        position: Optional[Dict[str, float]] = None
    ) -> str:
        """添加工具节点"""
        node_id = f"tool_{uuid4().hex[:8]}"
        config = ToolNodeConfig(
            node_type=NodeType.TOOL,
            name=name,
            tool_name=tool_name,
            parameters=parameters or {},
            output_key=output_key
        )
        node = Node(
            id=node_id,
            type=NodeType.TOOL,
            config=config,
            position=position or {"x": 0, "y": 0}
        )
        self.nodes.append(node)
        return node_id

    def add_rag_node(
        self,
        name: str,
        knowledge_base_id: str,
        top_k: int = 3,
        retrieval_strategy: str = "simple",
        output_key: str = "context",
        position: Optional[Dict[str, float]] = None
    ) -> str:
        """添加 RAG 节点"""
        node_id = f"rag_{uuid4().hex[:8]}"
        config = RAGNodeConfig(
            node_type=NodeType.RAG,
            name=name,
            knowledge_base_id=knowledge_base_id,
            top_k=top_k,
            retrieval_strategy=retrieval_strategy,
            output_key=output_key
        )
        node = Node(
            id=node_id,
            type=NodeType.RAG,
            config=config,
            position=position or {"x": 0, "y": 0}
        )
        self.nodes.append(node)
        return node_id

    def add_condition_node(
        self,
        name: str,
        condition: str,
        true_next: str = "",
        false_next: str = "",
        position: Optional[Dict[str, float]] = None
    ) -> str:
        """添加条件节点"""
        node_id = f"condition_{uuid4().hex[:8]}"
        config = ConditionNodeConfig(
            node_type=NodeType.CONDITION,
            name=name,
            condition=condition,
            true_next=true_next,
            false_next=false_next
        )
        node = Node(
            id=node_id,
            type=NodeType.CONDITION,
            config=config,
            position=position or {"x": 0, "y": 0}
        )
        self.nodes.append(node)
        return node_id

    def add_text_output_node(
        self,
        name: str,
        template: str = "{response}",
        position: Optional[Dict[str, float]] = None
    ) -> str:
        """添加文本输出节点"""
        node_id = f"output_{uuid4().hex[:8]}"
        config = TextOutputNodeConfig(
            node_type=NodeType.TEXT_OUTPUT,
            name=name,
            template=template
        )
        node = Node(
            id=node_id,
            type=NodeType.TEXT_OUTPUT,
            config=config,
            position=position or {"x": 0, "y": 0}
        )
        self.nodes.append(node)
        return node_id

    def add_end_node(
        self,
        position: Optional[Dict[str, float]] = None
    ) -> str:
        """添加结束节点"""
        node_id = f"end_{uuid4().hex[:8]}"
        config = TextOutputNodeConfig(
            node_type=NodeType.END,
            name="结束"
        )
        node = Node(
            id=node_id,
            type=NodeType.END,
            config=config,
            position=position or {"x": 0, "y": 0}
        )
        self.nodes.append(node)
        self.end_node_id = node_id
        return node_id

    def connect(self, source_id: str, target_id: str):
        """连接两个节点"""
        edge = Edge(
            source=source_id,
            target=target_id
        )
        self.edges.append(edge)

    def set_start(self, node_id: str):
        """设置起始节点"""
        self.start_node_id = node_id

    def build(
        self,
        workflow_id: str,
        name: str,
        description: str = "",
        tenant_id: str = ""
    ) -> Workflow:
        """构建工作流"""
        workflow = Workflow(
            id=workflow_id,
            name=name,
            description=description,
            tenant_id=tenant_id,
            nodes=self.nodes,
            edges=self.edges,
            start_node=self.start_node_id,
            end_node=self.end_node_id,
            created_at="",
            updated_at=""
        )
        return workflow

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "WorkflowBuilder":
        """从字典创建构建器"""
        builder = cls()

        nodes = data.get("nodes", [])
        for node_data in nodes:
            node_type = NodeType(node_data.get("type", "llm"))
            node_id = node_data.get("id", "")
            position = node_data.get("position", {"x": 0, "y": 0})
            config_data = node_data.get("config", {})

            if node_type == NodeType.LLM:
                config = LLMNodeConfig(
                    node_type=NodeType.LLM,
                    name=config_data.get("name", ""),
                    model=config_data.get("model", ""),
                    system_prompt=config_data.get("system_prompt", ""),
                    max_tokens=config_data.get("max_tokens", 2000),
                    temperature=config_data.get("temperature", 0.7),
                    output_key=config_data.get("output_key", "response")
                )
            elif node_type == NodeType.TOOL:
                config = ToolNodeConfig(
                    node_type=NodeType.TOOL,
                    name=config_data.get("name", ""),
                    tool_name=config_data.get("tool_name", ""),
                    parameters=config_data.get("parameters", {}),
                    output_key=config_data.get("output_key", "tool_result")
                )
            elif node_type == NodeType.RAG:
                config = RAGNodeConfig(
                    node_type=NodeType.RAG,
                    name=config_data.get("name", ""),
                    knowledge_base_id=config_data.get("knowledge_base_id", ""),
                    top_k=config_data.get("top_k", 3),
                    retrieval_strategy=config_data.get("retrieval_strategy", "simple"),
                    output_key=config_data.get("output_key", "context")
                )
            elif node_type == NodeType.CONDITION:
                config = ConditionNodeConfig(
                    node_type=NodeType.CONDITION,
                    name=config_data.get("name", ""),
                    condition=config_data.get("condition", ""),
                    true_next=config_data.get("true_next", ""),
                    false_next=config_data.get("false_next", "")
                )
            elif node_type == NodeType.TEXT_OUTPUT:
                config = TextOutputNodeConfig(
                    node_type=NodeType.TEXT_OUTPUT,
                    name=config_data.get("name", ""),
                    template=config_data.get("template", "{response}")
                )
            else:
                continue

            node = Node(
                id=node_id,
                type=node_type,
                config=config,
                position=position
            )
            builder.nodes.append(node)

        edges = data.get("edges", [])
        for edge_data in edges:
            edge = Edge(
                source=edge_data.get("source", ""),
                target=edge_data.get("target", ""),
                source_handle=edge_data.get("source_handle"),
                target_handle=edge_data.get("target_handle")
            )
            builder.edges.append(edge)

        builder.start_node_id = data.get("start_node", "")
        builder.end_node_id = data.get("end_node", "")

        return builder