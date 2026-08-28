"""工作流工具函数"""
import json
from typing import Dict, Any, List

from .node_types import (
    Workflow, WorkflowInstance, WorkflowResult,
    Node, Edge, NodeType
)


def dict_to_workflow(data: dict) -> Workflow:
    nodes_data = data.get("nodes", [])
    if isinstance(nodes_data, str):
        nodes_data = json.loads(nodes_data)
    
    edges_data = data.get("edges", [])
    if isinstance(edges_data, str):
        edges_data = json.loads(edges_data)
    
    nodes = []
    for n in nodes_data:
        ntype = NodeType(n.get("type", "start"))
        config = n.get("config", {})
        
        nodes.append(Node(
            id=n.get("id", ""),
            type=ntype,
            config=config,
            position=n.get("position", {"x": 0, "y": 0})
        ))

    edges = []
    for e in edges_data:
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