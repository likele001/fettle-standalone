from typing import Dict, Any, List, Optional
from dataclasses import dataclass, field
from enum import Enum


class NodeType(str, Enum):
    LLM = "llm"
    TOOL = "tool"
    RAG = "rag"
    CONDITION = "condition"
    HTTP = "http"
    TEXT_OUTPUT = "text_output"
    CRON = "cron"
    WEBHOOK = "webhook"
    INPUT = "input"
    START = "start"
    END = "end"


@dataclass
class NodeConfig:
    node_type: NodeType
    name: str = ""
    description: str = ""
    config: Dict[str, Any] = field(default_factory=dict)


@dataclass
class LLMNodeConfig(NodeConfig):
    model: str = ""
    system_prompt: str = ""
    max_tokens: int = 2000
    temperature: float = 0.7
    output_key: str = "response"


@dataclass
class ToolNodeConfig(NodeConfig):
    tool_name: str = ""
    parameters: Dict[str, Any] = field(default_factory=dict)
    output_key: str = "tool_result"


@dataclass
class RAGNodeConfig(NodeConfig):
    knowledge_base_id: str = ""
    top_k: int = 3
    retrieval_strategy: str = "simple"
    output_key: str = "context"


@dataclass
class ConditionNodeConfig(NodeConfig):
    condition: str = ""
    true_next: str = ""
    false_next: str = ""


@dataclass
class HTTPNodeConfig(NodeConfig):
    url: str = ""
    method: str = "GET"
    headers: Dict[str, str] = field(default_factory=dict)
    body: str = ""
    output_key: str = "http_response"


@dataclass
class TextOutputNodeConfig(NodeConfig):
    template: str = "{response}"


@dataclass
class CronNodeConfig(NodeConfig):
    cron: str = "0 9 * * *"
    timezone: str = "Asia/Shanghai"


@dataclass
class WebhookNodeConfig(NodeConfig):
    path: str = ""
    secret: str = ""


@dataclass
class Node:
    id: str
    type: NodeType
    config: NodeConfig
    position: Dict[str, float] = field(default_factory=dict)


@dataclass
class Edge:
    source: str
    target: str
    source_handle: Optional[str] = None
    target_handle: Optional[str] = None


@dataclass
class Workflow:
    id: str
    name: str
    description: str = ""
    tenant_id: str = ""
    nodes: List[Node] = field(default_factory=list)
    edges: List[Edge] = field(default_factory=list)
    start_node: str = ""
    end_node: str = ""
    created_at: str = ""
    updated_at: str = ""


@dataclass
class WorkflowInstance:
    instance_id: str
    workflow_id: str
    tenant_id: str
    user_id: str
    status: str = "running"
    input_data: Dict[str, Any] = field(default_factory=dict)
    output_data: Dict[str, Any] = field(default_factory=dict)
    node_states: Dict[str, Any] = field(default_factory=dict)
    created_at: str = ""
    started_at: str = ""
    completed_at: str = ""


@dataclass
class WorkflowResult:
    success: bool = False
    output: Dict[str, Any] = field(default_factory=dict)
    error: str = ""
    steps: List[Dict[str, Any]] = field(default_factory=list)
    duration: float = 0.0
