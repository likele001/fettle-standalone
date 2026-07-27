"""工作流工具函数"""
from typing import Dict, Any, List

from .node_types import (
    Workflow, WorkflowInstance, WorkflowResult,
    Node, Edge, NodeType,
    LLMNodeConfig, ToolNodeConfig, RAGNodeConfig,
    ConditionNodeConfig, HTTPNodeConfig,
    TextOutputNodeConfig, CronNodeConfig, WebhookNodeConfig
)


def dict_to_workflow(data: dict) -> Workflow:
    nodes = []
    for n in data.get("nodes", []):
        ntype = NodeType(n.get("type", "start"))
        cfg = n.get("config", {})
        if isinstance(cfg, dict):
            config_map = {
                NodeType.LLM: ("LLMNodeConfig", LLMNodeConfig),
                NodeType.TOOL: ("ToolNodeConfig", ToolNodeConfig),
                NodeType.RAG: ("RAGNodeConfig", RAGNodeConfig),
                NodeType.CONDITION: ("ConditionNodeConfig", ConditionNodeConfig),
                NodeType.HTTP: ("HTTPNodeConfig", HTTPNodeConfig),
                NodeType.TEXT_OUTPUT: ("TextOutputNodeConfig", TextOutputNodeConfig),
                NodeType.CRON: ("CronNodeConfig", CronNodeConfig),
                NodeType.WEBHOOK: ("WebhookNodeConfig", WebhookNodeConfig),
            }
            entry = config_map.get(ntype)
            if entry:
                _, config_cls = entry
                config = config_cls(node_type=ntype, **cfg)
            else:
                config = cfg
        else:
            config = cfg

        nodes.append(Node(
            id=n.get("id", ""),
            type=ntype,
            config=config,
            position=n.get("position", {"x": 0, "y": 0})
        ))

    edges = []
    for e in data.get("edges", []):
        edges.append(Edge(
            source=e.get("source", ""),
            target=e.get("target", ""),
            source_handle=e.get("sourceHandle"),
            target_handle=e.get("targetHandle"),
        ))

    return Workflow(
        id=data.get("id", ""),
        name=data.get("name", "Untitled"),
        description=data.get("description", ""),
        tenant_id=data.get("tenant_id", ""),
        nodes=nodes,
        edges=edges,
        start_node=data.get("start_node", ""),
    )
