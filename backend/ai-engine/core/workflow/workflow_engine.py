"""LangGraph 工作流引擎"""
import json
import logging
import time
from typing import Dict, Any, List, Optional

import httpx

from langgraph.graph import StateGraph, END
from langgraph.checkpoint.memory import MemorySaver

from .node_types import (
    Workflow, WorkflowInstance, WorkflowResult,
    NodeType, LLMNodeConfig, ToolNodeConfig, RAGNodeConfig,
    ConditionNodeConfig, HTTPNodeConfig, TextOutputNodeConfig
)
from core.models.providers.langchain_provider import LangChainProvider
from core.tools.tool_executor import tool_executor
from core.rag.advanced_retrievers import RAGRetrieverFactory

logger = logging.getLogger(__name__)


class WorkflowEngine:

    def __init__(self):
        self.checkpointer = MemorySaver()

    async def execute(
        self,
        workflow: Workflow,
        instance: WorkflowInstance
    ) -> WorkflowResult:
        start_time = time.time()
        result = WorkflowResult()

        try:
            graph = await self._build_graph(workflow, instance)
            if graph is None:
                result.error = "Failed to build workflow graph"
                result.duration = time.time() - start_time
                return result

            config = {"configurable": {"thread_id": instance.instance_id}}
            output = await graph.ainvoke(instance.input_data, config)

            result.success = True
            result.output = output
            result.duration = time.time() - start_time

        except Exception as e:
            logger.error(f"Workflow execution failed: {e}")
            result.error = str(e)
            result.duration = time.time() - start_time

        return result

    async def stream_execute(
        self,
        workflow: Workflow,
        instance: WorkflowInstance
    ):
        try:
            graph = await self._build_graph(workflow, instance)
            if graph is None:
                yield {"error": "Failed to build workflow graph", "is_final": True}
                return

            config = {"configurable": {"thread_id": instance.instance_id}}
            async for event in graph.astream(instance.input_data, config):
                yield event

        except Exception as e:
            logger.error(f"Stream workflow execution failed: {e}")
            yield {"error": str(e), "is_final": True}

    async def _build_graph(self, workflow: Workflow, instance: WorkflowInstance):
        try:
            workflow_state = {
                "messages": [],
                "context": {},
                "output": {},
                **instance.input_data
            }

            class WorkflowState(Dict):
                pass

            builder = StateGraph(WorkflowState)

            node_map = {node.id: node for node in workflow.nodes}
            for node in workflow.nodes:
                node_func = await self._create_node_function(node, instance)
                if node_func:
                    builder.add_node(node.id, node_func)

            edges_by_source: Dict[str, List] = {}
            for edge in workflow.edges:
                edges_by_source.setdefault(edge.source, []).append(edge)

            for source_id, outgoing in edges_by_source.items():
                source_node = node_map.get(source_id)
                is_condition = source_node and source_node.type == NodeType.CONDITION

                if is_condition:
                    true_targets = [e.target for e in outgoing if e.source_handle == "true"]
                    false_targets = [e.target for e in outgoing if e.source_handle in (None, "", "false")]

                    true_target = true_targets[0] if true_targets else None
                    false_target = false_targets[0] if false_targets else None

                    if true_target or false_target:
                        builder.add_conditional_edges(
                            source_id,
                            _make_condition_router(true_target, false_target)
                        )
                else:
                    for edge in outgoing:
                        builder.add_edge(edge.source, edge.target)

            if workflow.start_node:
                builder.set_entry_point(workflow.start_node)

            graph = builder.compile(checkpointer=self.checkpointer)
            return graph

        except Exception as e:
            logger.error(f"Failed to build graph: {e}")
            return None

    async def _create_node_function(self, node, instance: WorkflowInstance):
        node_type = NodeType(node.type)
        config = node.config

        if node_type == NodeType.LLM:
            return await self._create_llm_node(config, instance)
        elif node_type == NodeType.TOOL:
            return await self._create_tool_node(config, instance)
        elif node_type == NodeType.RAG:
            return await self._create_rag_node(config, instance)
        elif node_type == NodeType.CONDITION:
            return await self._create_condition_node(config)
        elif node_type == NodeType.HTTP:
            return await self._create_http_node(config)
        elif node_type == NodeType.TEXT_OUTPUT:
            return await self._create_text_output_node(config)
        elif node_type == NodeType.END:
            return lambda state: {**state, "__end__": True}
        elif node_type == NodeType.START:
            return lambda state: state

        return None

    async def _create_llm_node(self, config: LLMNodeConfig, instance: WorkflowInstance):
        async def llm_node(state: Dict[str, Any]) -> Dict[str, Any]:
            try:
                provider = await self._get_langchain_provider(instance)
                if not provider or not provider.llm:
                    return state

                llm = provider.get_llm(config.model if config.model else None)
                if not llm:
                    return state

                messages = list(state.get("messages", []))
                prompt = config.get("prompt", config.get("system_prompt", ""))

                history = state.get("messages", [])
                if prompt:
                    history.insert(0, {"role": "system", "content": prompt})
                if history:
                    messages = history

                response = await llm.ainvoke(messages)
                state["messages"].append({"role": "assistant", "content": response.content})
                output_key = config.get("output_key", "response")
                state["context"][output_key] = response.content

                return state
            except Exception as e:
                logger.error(f"LLM node error: {e}")
                state["context"]["error"] = str(e)
                return state

        return llm_node

    async def _create_tool_node(self, config: ToolNodeConfig, instance: WorkflowInstance):
        async def tool_node(state: Dict[str, Any]) -> Dict[str, Any]:
            try:
                params = {}
                for key, value in config.get("parameters", {}).items():
                    if isinstance(value, str) and value.startswith("{") and value.endswith("}"):
                        param_key = value[1:-1]
                        params[key] = state.get("context", {}).get(param_key, value)
                    else:
                        params[key] = value

                result = await tool_executor.execute(
                    tool_name=config.get("tool_name", ""),
                    parameters=params,
                    context={"tenant_id": instance.tenant_id, "user_id": instance.user_id}
                )

                output_key = config.get("output_key", "tool_result")
                state["context"][output_key] = result
                state["messages"].append({
                    "role": "tool",
                    "content": str(result)
                })

                return state
            except Exception as e:
                logger.error(f"Tool node error: {e}")
                return state

        return tool_node

    async def _create_rag_node(self, config: RAGNodeConfig, instance: WorkflowInstance):
        async def rag_node(state: Dict[str, Any]) -> Dict[str, Any]:
            try:
                query = state.get("input", "") or state.get("messages", [{}])[-1].get("content", "")

                provider = await self._get_langchain_provider(instance)
                llm = provider.get_llm() if provider else None

                retriever = RAGRetrieverFactory.create_retriever(
                    knowledge_base_id=config.get("knowledge_base_id", ""),
                    tenant_id=instance.tenant_id,
                    llm=llm,
                    retrieval_strategy=config.get("retrieval_strategy", "simple"),
                    top_k=config.get("top_k", 3)
                )

                docs = await retriever.agetrieve(query)

                context_parts = []
                for i, doc in enumerate(docs, 1):
                    score = doc.metadata.get("score", 0)
                    context_parts.append(f"[文档{i}] (相关度: {score:.2f})\n{doc.page_content}")

                context_text = "\n\n---\n\n".join(context_parts)
                output_key = config.get("output_key", "context")
                state["context"][output_key] = context_text

                return state
            except Exception as e:
                logger.error(f"RAG node error: {e}")
                return state

        return rag_node

    async def _create_condition_node(self, config: ConditionNodeConfig):
        async def condition_node(state: Dict[str, Any]) -> Dict[str, Any]:
            try:
                condition = config.get("condition", "")
                context = state.get("context", {})
                merged = {**state, **context}

                for key, value in merged.items():
                    if isinstance(value, (str, int, float, bool)):
                        placeholder = "{{%s}}" % key
                        condition = condition.replace(placeholder, str(value))
                        condition = condition.replace("{%s}" % key, str(value))

                result = bool(eval(condition, {"__builtins__": {}}, merged))
                state["context"]["condition_result"] = result

                return state
            except Exception as e:
                logger.error(f"Condition node error: {e}")
                state["context"]["condition_result"] = False
                return state

        return condition_node

    async def _create_http_node(self, config: HTTPNodeConfig):
        async def http_node(state: Dict[str, Any]) -> Dict[str, Any]:
            try:
                url = config.get("url", "")
                method = config.get("method", "GET").upper()
                headers = config.get("headers", {})
                body_str = config.get("body", "")

                body = None
                if body_str:
                    body_str = _render_template(body_str, state)
                    try:
                        body = json.loads(body_str)
                    except (json.JSONDecodeError, TypeError):
                        body = body_str

                async with httpx.AsyncClient(timeout=30) as client:
                    if method == "GET":
                        resp = await client.get(url, headers=headers)
                    elif method == "POST":
                        resp = await client.post(url, headers=headers, json=body)
                    elif method == "PUT":
                        resp = await client.put(url, headers=headers, json=body)
                    elif method == "DELETE":
                        resp = await client.delete(url, headers=headers)
                    else:
                        resp = await client.get(url, headers=headers)

                    output_key = config.get("output_key", "http_response")
                    state["context"][output_key] = {
                        "status": resp.status_code,
                        "headers": dict(resp.headers),
                        "body": _try_parse_json(resp.text),
                    }

                return state
            except Exception as e:
                logger.error(f"HTTP node error: {e}")
                state["context"]["http_error"] = str(e)
                return state

        return http_node

    async def _create_text_output_node(self, config: TextOutputNodeConfig):
        async def text_output_node(state: Dict[str, Any]) -> Dict[str, Any]:
            try:
                template = config.get("template", "{response}")
                context = state.get("context", {})
                merged = {**state, **context}
                result = _render_template(template, merged)
                state["output"]["text"] = result
                return state
            except Exception as e:
                logger.error(f"Text output node error: {e}")
                return state

        return text_output_node

    async def _get_langchain_provider(self, instance: WorkflowInstance) -> Optional[LangChainProvider]:
        from core.models.providers.db_factory import DBProviderFactory

        try:
            provider = await DBProviderFactory.get_provider(
                provider_code="openai",
                api_key="",
                use_langchain=True
            )
            return provider
        except Exception as e:
            logger.error(f"Failed to get LangChain provider: {e}")
            return None


def _make_condition_router(true_target: Optional[str], false_target: Optional[str]):
    def router(state: Dict) -> str:
        if state.get("context", {}).get("condition_result", False):
            return true_target or END
        return false_target or END
    return router


def _render_template(template: str, context: Dict) -> str:
    result = template
    for key, value in context.items():
        if isinstance(value, (str, int, float, bool)):
            result = result.replace("{{%s}}" % key, str(value))
            result = result.replace("{%s}" % key, str(value))
    return result


def _try_parse_json(text: str) -> Any:
    try:
        return json.loads(text)
    except (json.JSONDecodeError, TypeError):
        return text


workflow_engine = WorkflowEngine()
