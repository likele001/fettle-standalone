"""工作流执行引擎 - 简化版"""
import json
import logging
import time
from typing import Dict, Any, List, Optional

import httpx

from .node_types import (
    Workflow, WorkflowInstance, WorkflowResult,
    NodeType
)

from core.billing import check_billing_quota, consume_billing, BillingQuotaExceeded
from core.models.model_gateway import model_gateway

logger = logging.getLogger(__name__)


class WorkflowEngine:

    def __init__(self):
        pass

    async def execute(
        self,
        workflow: Workflow,
        instance: WorkflowInstance
    ) -> WorkflowResult:
        start_time = time.time()
        result = WorkflowResult()

        try:
            state = await self._execute_workflow(workflow, instance)
            result.success = True
            result.output = state
            result.duration = time.time() - start_time

        except Exception as e:
            logger.error(f"Workflow execution failed: {e}")
            result.error = str(e)
            result.duration = time.time() - start_time

        return result

    async def _execute_workflow(self, workflow: Workflow, instance: WorkflowInstance) -> Dict[str, Any]:
        node_map = {node.id: node for node in workflow.nodes}

        edges_by_source: Dict[str, List] = {}
        for edge in workflow.edges:
            edges_by_source.setdefault(edge.source, []).append(edge)

        has_edges = len(workflow.edges) > 0

        sorted_nodes = sorted(workflow.nodes, key=lambda n: n.position.get("y", 0))

        # 修复: state 初始化不允许 input_data 覆盖内部结构键(messages/context/output)
        state: Dict[str, Any] = {
            "messages": [],
            "context": {},
            "output": {},
        }
        for _k, _v in (instance.input_data or {}).items():
            if _k not in ("messages", "context", "output"):
                state[_k] = _v

        current_node_id = workflow.start_node

        if not current_node_id:
            for node in workflow.nodes:
                if node.type == NodeType.START:
                    current_node_id = node.id
                    break

        if not current_node_id and workflow.nodes:
            current_node_id = workflow.nodes[0].id

        visited = set()

        while current_node_id and current_node_id not in visited:
            visited.add(current_node_id)
            node = node_map.get(current_node_id)

            if not node:
                break

            state = await self._execute_node(node, state, instance)

            if node.type == NodeType.END:
                break

            outgoing = edges_by_source.get(current_node_id, [])

            if node.type == NodeType.CONDITION:
                condition_result = state.get("context", {}).get("condition_result", False)
                matched = None
                for edge in outgoing:
                    handle = edge.source_handle or "false"
                    if (condition_result and handle == "true") or (not condition_result and handle == "false"):
                        matched = edge.target
                        break
                if not matched and outgoing:
                    matched = outgoing[0].target
                current_node_id = matched
            elif outgoing:
                current_node_id = outgoing[0].target
            elif not has_edges:
                next_node_id = None
                for i, n in enumerate(sorted_nodes):
                    if n.id == current_node_id:
                        if i + 1 < len(sorted_nodes):
                            next_node_id = sorted_nodes[i + 1].id
                        break
                current_node_id = next_node_id

        return state

    async def _execute_node(self, node, state: Dict[str, Any], instance: WorkflowInstance) -> Dict[str, Any]:
        node_type = NodeType(node.type)
        config = node.config

        if node_type == NodeType.LLM:
            return await self._execute_llm_node(config, state, instance)
        elif node_type == NodeType.TOOL:
            return await self._execute_tool_node(config, state, instance)
        elif node_type == NodeType.RAG:
            return await self._execute_rag_node(config, state, instance)
        elif node_type == NodeType.CONDITION:
            return await self._execute_condition_node(config, state)
        elif node_type == NodeType.HTTP:
            return await self._execute_http_node(config, state)
        elif node_type == NodeType.AGENT:
            from .workflow_agent_executor import execute_agent_node
            return await execute_agent_node(
                config, state,
                tenant_id=getattr(instance, "tenant_id", ""),
                user_id=getattr(instance, "user_id", ""),
            )
        elif node_type == NodeType.TEXT_OUTPUT:
            return self._execute_text_output_node(config, state)
        elif node_type == NodeType.START:
            return state
        elif node_type == NodeType.END:
            return state

        return state

    def _execute_text_output_node(self, config: Dict, state: Dict[str, Any]) -> Dict[str, Any]:
        """文本输出节点: 将模板渲染结果写入最终输出"""
        try:
            template = config.get("template", config.get("text", ""))
            rendered = _render_template(template, _flatten_state(state))
            output_key = config.get("output_key", "text")
            state["context"][output_key] = rendered
            state["output"][output_key] = rendered
            state["output"]["text"] = rendered
        except Exception as e:
            logger.error(f"TEXT_OUTPUT node error: {e}")
            state["context"]["error"] = str(e)
        return state

    async def _execute_llm_node(self, config: Dict, state: Dict[str, Any], instance: WorkflowInstance) -> Dict[str, Any]:
        try:
            messages = list(state.get("messages", []))
            prompt = config.get("prompt", config.get("system_prompt", ""))
            model_name = config.get("model", "gpt-3.5-turbo")
            temperature = float(config.get("temperature", 0.7))
            max_tokens = int(config.get("max_tokens", 2000))
            # 兜底：过小的 max_tokens 会导致部分 provider（如 deepseek-v4-flash）返回空内容
            if max_tokens < 64:
                max_tokens = 64

            system_prompt = prompt or ""
            user_messages = []

            # 工作流多节点场景：每个 LLM 节点是独立子任务，
            # 只使用当前节点的真实用户输入，避免前序节点的 assistant 输出污染后续节点。
            explicit_input = state.get("input", "") or state.get("query", "")
            if explicit_input:
                user_messages.append({"role": "user", "content": str(explicit_input)})
            else:
                for msg in state.get("messages") or []:
                    if isinstance(msg, dict) and msg.get("role") == "user":
                        user_messages.append({"role": "user", "content": str(msg.get("content", ""))})

            if not user_messages:
                state["context"]["error"] = "No input found"
                return state

            try:
                provider, model_config = await self._get_provider(instance)
                # P1-1 配额前置校验：LLM 节点消耗真实 token，需检查并记账
                try:
                    await check_billing_quota(instance.tenant_id)
                except BillingQuotaExceeded as qe:
                    logger.warning(f"Quota exceeded for tenant {instance.tenant_id}: {qe}")
                    llm_response = str(qe)
                    state["context"]["quota_exceeded"] = True
                    state["messages"].append({"role": "assistant", "content": llm_response})
                    return state
                if provider and hasattr(provider, 'chat'):
                    # 使用租户真实模型（节点配置的 model 仅作参考，避免与租户厂商不匹配）
                    effective_model = model_name
                    if model_config and model_config.model_code:
                        effective_model = model_config.model_code
                    # 真实 provider 期望 ChatMessage 对象，将 dict 消息转换
                    from core.models.providers.base import ChatMessage
                    chat_messages = [
                        ChatMessage(role=m.get("role", "user"), content=m.get("content", ""))
                        for m in user_messages
                    ]
                    # 若配置了 system prompt 且尚未包含 system 消息，则前置注入
                    if system_prompt and not any(m.role == "system" for m in chat_messages):
                        chat_messages.insert(0, ChatMessage(role="system", content=system_prompt))
                    # P2-1：经统一模型网关调用，自动聚合降级 + 厂商级熔断
                    response = await model_gateway.chat(
                        tenant_id=instance.tenant_id,
                        preferred=model_config,
                        messages=chat_messages,
                        temperature=temperature,
                        max_tokens=max_tokens,
                    )
                    # 真实 provider 返回 ChatResponse 对象（.content），兼容 dict
                    llm_response = response.content if hasattr(response, "content") else response.get("content", "")
                    # 空响应兜底：重试一次（分类/路由等短输出更稳）
                    if not str(llm_response).strip():
                        response2 = await model_gateway.chat(
                            tenant_id=instance.tenant_id,
                            preferred=model_config,
                            messages=chat_messages,
                            temperature=temperature,
                            max_tokens=max_tokens,
                        )
                        llm_response = response2.content if hasattr(response2, "content") else response2.get("content", "")
                    # P1-1 记账：LLM 成功后累加 token 用量（尽力而为，不因失败中断流程）
                    try:
                        used = int(getattr(response, "tokens_used", 0) or 0)
                        if used <= 0:
                            output_used = 0
                        else:
                            output_used = used // 2
                        model_code = getattr(model_config, "model_code", "") or getattr(model_config, "model_id", "") or ""
                        await consume_billing(instance.tenant_id, model_code, used - output_used, output_used)
                        try:
                            from core.tracing import log_trace
                            log_trace(instance.tenant_id, "workflow_llm", model_code,
                                      used - output_used, output_used, 0, "success")
                        except Exception:
                            pass
                    except Exception as consume_err:
                        logger.warning(f"workflow consume_billing failed: {consume_err}")
                else:
                    llm_response = f"[模拟LLM响应] 模型: {model_name}\n输入: {user_messages[-1]['content']}"
                    logger.warning("Using mock LLM response - no provider available")
            except Exception as llm_error:
                logger.error(f"LLM call failed: {llm_error}")
                llm_response = f"[LLM调用失败] {str(llm_error)}"

            state["messages"].append({"role": "assistant", "content": llm_response})
            output_key = config.get("output_key", "response")
            state["context"][output_key] = llm_response
            state["output"][output_key] = llm_response
            state["output"]["text"] = llm_response

        except Exception as e:
            logger.error(f"LLM node error: {e}")
            state["context"]["error"] = str(e)

        return state

    async def _execute_tool_node(self, config: Dict, state: Dict[str, Any], instance: WorkflowInstance) -> Dict[str, Any]:
        try:
            tool_name = config.get("tool_name", "")
            params = {}
            flat = _flatten_state(state)
            for key, value in config.get("parameters", {}).items():
                if isinstance(value, str) and "{" in value and "}" in value:
                    # 支持 {param_key} 与 {{param_key}} 模板变量（来自 context 或顶层 input_data）
                    params[key] = _render_template(value, flat)
                else:
                    params[key] = value

            try:
                from core.tools.tool_executor import tool_executor
                tool_result = await tool_executor.execute(
                    tool_name=tool_name,
                    parameters=params,
                    context={"tenant_id": instance.tenant_id, "user_id": instance.user_id}
                )

                if tool_result.get("status") == "error":
                    raise Exception(tool_result.get("error", "Tool execution failed"))
            except Exception:
                tool_result = {"mock": True, "tool": tool_name, "params": params, "message": "工具调用成功（模拟）"}
                logger.warning(f"Using mock tool response for {tool_name}")

            output_key = config.get("output_key", "tool_result")
            state["context"][output_key] = tool_result
            state["output"][output_key] = tool_result
            state["messages"].append({
                "role": "tool",
                "content": str(tool_result)
            })

        except Exception as e:
            logger.error(f"Tool node error: {e}")

        return state

    async def _execute_rag_node(self, config: Dict, state: Dict[str, Any], instance: WorkflowInstance) -> Dict[str, Any]:
        try:
            query = state.get("input", "") or state.get("messages", [{}])[-1].get("content", "")
            knowledge_base_id = config.get("knowledge_base_id", "")
            strategy = config.get("strategy", "simple")
            top_k = int(config.get("top_k", 3))

            try:
                from core.rag.advanced_retrievers import RAGRetrieverFactory
                provider, _ = await self._get_provider(instance)
                llm = provider.get_llm() if provider and hasattr(provider, 'get_llm') else None

                retriever = RAGRetrieverFactory.create_retriever(
                    knowledge_base_id=knowledge_base_id,
                    tenant_id=instance.tenant_id,
                    llm=llm,
                    retrieval_strategy=strategy,
                    top_k=top_k
                )
                docs = await retriever.agetrieve(query)

                context_parts = []
                for i, doc in enumerate(docs, 1):
                    score = doc.metadata.get("score", 0)
                    context_parts.append(f"[文档{i}] (相关度: {score:.2f})\n{doc.page_content}")
                context_text = "\n\n---\n\n".join(context_parts)
            except Exception:
                context_text = f"[模拟RAG检索结果]\n知识库ID: {knowledge_base_id}\n查询: {query}\n策略: {strategy}\n\n这是模拟的知识库文档内容。在实际部署中，这里会返回从向量数据库检索到的相关文档。"
                logger.warning("Using mock RAG response")

            output_key = config.get("output_key", "context")
            state["context"][output_key] = context_text
            state["output"][output_key] = context_text

        except Exception as e:
            logger.error(f"RAG node error: {e}")

        return state

    async def _execute_condition_node(self, config: Dict, state: Dict[str, Any]) -> Dict[str, Any]:
        try:
            condition = config.get("condition", "")
            context = state.get("context", {})
            merged = {**state, **context}

            for key, value in merged.items():
                if isinstance(value, (str, int, float, bool)):
                    placeholder = "{{%s}}" % key
                    condition = condition.replace(placeholder, str(value))
                    condition = condition.replace("{%s}" % key, str(value))

            result = False
            try:
                result = bool(eval(condition, {"__builtins__": {}}, merged))
            except Exception:
                result = False

            state["context"]["condition_result"] = result
            state["output"]["condition_result"] = result

        except Exception as e:
            logger.error(f"Condition node error: {e}")
            state["context"]["condition_result"] = False

        return state

    async def _execute_http_node(self, config: Dict, state: Dict[str, Any]) -> Dict[str, Any]:
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
                state["output"][output_key] = state["context"][output_key]

        except Exception as e:
            logger.error(f"HTTP node error: {e}")
            state["context"]["http_error"] = str(e)

        return state

    async def _get_provider(self, instance: WorkflowInstance):
        """获取真实 Provider（复用主对话链路的租户配置解析）。

        返回 (provider, model_config)；无可用配置时返回 (None, None)。
        """
        try:
            from core.models.providers.db_factory import DBProviderFactory
            from core.models.db_model_router import db_model_router

            tenant_id = getattr(instance, "tenant_id", "") or "00000000-0000-0000-0000-000000000001"

            # 1. 选择模型（优先租户默认/有 key 的厂商）
            model_config = await db_model_router.select_model(tenant_id=tenant_id)

            # 2. 获取租户 API Key
            tenant_api_key = await db_model_router.get_tenant_api_key(
                tenant_id, model_config.provider_id
            )

            # 3. 创建 Provider
            provider = await DBProviderFactory.get_provider(
                model_config.provider_code,
                tenant_api_key=tenant_api_key,
            )
            if not provider:
                logger.warning(f"No provider available for tenant {tenant_id}")
                return None, None
            return provider, model_config
        except Exception as e:
            logger.warning(f"Failed to get provider: {e}")
            return None, None


def _flatten_state(state: Dict) -> Dict:
    """将 context/output/顶层键拍平为一个字典，供模板渲染使用。
    非标量值(如 dict/list)渲染为 JSON 字符串，便于在文本模板中直接引用。
    优先级: 顶层输入 > output > context。
    """
    flat: Dict[str, Any] = {}

    def _put(k: str, v: Any) -> None:
        if isinstance(v, (str, int, float, bool)):
            flat[k] = v
        else:
            flat[k] = json.dumps(v, ensure_ascii=False)

    for k, v in (state.get("context") or {}).items():
        _put(k, v)
    for k, v in (state.get("output") or {}).items():
        _put(k, v)
    for k, v in state.items():
        if k not in ("messages", "context", "output"):
            _put(k, v)
    return flat


def _render_template(template: str, context: Dict) -> str:
    if not template:
        return ""
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