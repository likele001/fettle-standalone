"""对话服务 - 核心对话逻辑"""
import logging
import time
from typing import List, Dict, Any, Optional, AsyncIterator
from dataclasses import dataclass

from core.models.model_router import model_router, TaskType
from core.models.providers import ProviderFactory, ChatMessage
from core.rag.knowledge_retriever import KnowledgeRetriever
from core.rag.advanced_retrievers import RAGRetrieverFactory
from core.workflow import Workflow, WorkflowInstance, workflow_engine

logger = logging.getLogger(__name__)


@dataclass
class ChatRequest:
    """对话请求"""
    tenant_id: str
    user_id: str
    input: str
    agent_id: str = ""
    conversation_id: str = ""
    history: List[Dict[str, str]] = None
    system_prompt: str = ""
    model: str = ""
    knowledge_base_id: str = ""
    context: Dict[str, str] = None
    workflow_id: str = ""
    workflow_config: Dict[str, Any] = None
    retrieval_strategy: str = "simple"


@dataclass
class ChatResponse:
    """对话响应"""
    reply: str
    model_used: str
    tokens_used: int
    latency_ms: float
    knowledge_chunks: List[Dict[str, Any]] = None


class ChatService:
    """对话服务"""

    def __init__(self):
        self.knowledge_retriever = KnowledgeRetriever()

    async def generate_reply(self, request: ChatRequest) -> ChatResponse:
        """生成回复"""
        start_time = time.time()
        logger.info(f"Generate reply: tenant={request.tenant_id}, input={request.input[:50]}")

        try:
            if request.workflow_id or request.workflow_config:
                return await self._generate_with_workflow(request, start_time)

            return await self._generate_direct(request, start_time)

        except Exception as e:
            logger.error(f"Generate reply failed: {e}")
            return ChatResponse(
                reply=f"生成回复失败：{str(e)}",
                model_used="",
                tokens_used=0,
                latency_ms=(time.time() - start_time) * 1000
            )

    async def _generate_with_workflow(self, request: ChatRequest, start_time: float) -> ChatResponse:
        """使用工作流生成回复"""
        try:
            workflow_config = request.workflow_config or {}
            
            workflow = Workflow(
                id=request.workflow_id or "temp_workflow",
                name="Dynamic Workflow",
                nodes=[],
                edges=[],
                start_node="",
                tenant_id=request.tenant_id
            )

            instance = WorkflowInstance(
                instance_id=request.conversation_id or f"conv_{time.time()}",
                workflow_id=workflow.id,
                tenant_id=request.tenant_id,
                user_id=request.user_id,
                input_data={
                    "input": request.input,
                    "history": request.history or [],
                    "context": request.context or {},
                    "knowledge_base_id": request.knowledge_base_id
                }
            )

            result = await workflow_engine.execute(workflow, instance)

            latency_ms = (time.time() - start_time) * 1000

            if result.success:
                output_text = str(result.output.get("output", {}).get("text", "") or result.output.get("context", {}).get("response", ""))
                return ChatResponse(
                    reply=output_text,
                    model_used="workflow",
                    tokens_used=0,
                    latency_ms=latency_ms
                )
            else:
                return await self._generate_direct(request, start_time)

        except Exception as e:
            logger.error(f"Workflow generation failed: {e}")
            return await self._generate_direct(request, start_time)

    async def _generate_direct(self, request: ChatRequest, start_time: float) -> ChatResponse:
        """直接生成回复"""
        model_config = model_router.select_model(
            task_type=TaskType.KNOWLEDGE_QA if request.knowledge_base_id else TaskType.SIMPLE_CHAT,
            tenant_plan=request.context.get("tenant_plan") if request.context else None,
            model_id=request.model if request.model else None
        )

        provider = ProviderFactory.get_provider(model_config.provider)
        if not provider:
            return ChatResponse(
                reply="AI 服务暂时不可用，请稍后再试。",
                model_used="",
                tokens_used=0,
                latency_ms=(time.time() - start_time) * 1000
            )

        messages, knowledge_chunks = await self._build_messages(request)

        response = await provider.chat(
            messages=messages,
            model=model_config.model_name,
            max_tokens=model_config.max_tokens,
            temperature=model_config.temperature
        )

        latency_ms = (time.time() - start_time) * 1000

        return ChatResponse(
            reply=response.content,
            model_used=response.model,
            tokens_used=response.tokens_used,
            latency_ms=latency_ms,
            knowledge_chunks=knowledge_chunks
        )

    async def stream_reply(self, request: ChatRequest) -> AsyncIterator[Dict[str, Any]]:
        """流式生成回复"""
        logger.info(f"Stream reply: tenant={request.tenant_id}, input={request.input[:50]}")

        try:
            if request.workflow_id or request.workflow_config:
                async for event in self._stream_with_workflow(request):
                    yield event
                return

            async for event in self._stream_direct(request):
                yield event

        except Exception as e:
            logger.error(f"Stream reply failed: {e}")
            yield {
                "chunk": f"生成回复失败：{str(e)}",
                "is_final": True,
                "model_used": "",
                "tokens_used": 0
            }

    async def _stream_with_workflow(self, request: ChatRequest):
        """使用工作流流式生成"""
        try:
            workflow = Workflow(
                id=request.workflow_id or "temp_workflow",
                name="Dynamic Workflow",
                nodes=[],
                edges=[],
                start_node="",
                tenant_id=request.tenant_id
            )

            instance = WorkflowInstance(
                instance_id=request.conversation_id or f"conv_{time.time()}",
                workflow_id=workflow.id,
                tenant_id=request.tenant_id,
                user_id=request.user_id,
                input_data={
                    "input": request.input,
                    "history": request.history or [],
                    "context": request.context or {},
                    "knowledge_base_id": request.knowledge_base_id
                }
            )

            async for event in workflow_engine.stream_execute(workflow, instance):
                yield {"chunk": str(event), "is_final": False}

            yield {"chunk": "", "is_final": True}

        except Exception as e:
            logger.error(f"Stream workflow failed: {e}")
            yield {"chunk": str(e), "is_final": True}

    async def _stream_direct(self, request: ChatRequest):
        """直接流式生成"""
        model_config = model_router.select_model(
            task_type=TaskType.KNOWLEDGE_QA if request.knowledge_base_id else TaskType.SIMPLE_CHAT,
            tenant_plan=request.context.get("tenant_plan") if request.context else None,
            model_id=request.model if request.model else None
        )

        provider = ProviderFactory.get_provider(model_config.provider)
        if not provider:
            yield {
                "chunk": "AI 服务暂时不可用，请稍后再试。",
                "is_final": True,
                "model_used": "",
                "tokens_used": 0
            }
            return

        messages, knowledge_chunks = await self._build_messages(request)

        tokens_used = 0
        full_content = []

        async for chunk in provider.stream_chat(
            messages=messages,
            model=model_config.model_name,
            max_tokens=model_config.max_tokens,
            temperature=model_config.temperature
        ):
            tokens_used += 1
            full_content.append(chunk)
            yield {
                "chunk": chunk,
                "is_final": False,
                "model_used": model_config.model_name,
                "tokens_used": tokens_used
            }

        yield {
            "chunk": "",
            "is_final": True,
            "model_used": model_config.model_name,
            "tokens_used": tokens_used,
            "knowledge_chunks": knowledge_chunks
        }

    async def _build_messages(self, request: ChatRequest):
        """构建消息列表"""
        messages = []
        knowledge_chunks = []

        system_parts = []

        if request.system_prompt:
            system_parts.append(request.system_prompt)

        if request.knowledge_base_id:
            context_text, chunks = await self._retrieve_knowledge(
                knowledge_base_id=request.knowledge_base_id,
                query=request.input,
                tenant_id=request.tenant_id,
                strategy=request.retrieval_strategy
            )
            if context_text:
                system_parts.append(f"请基于以下知识库内容回答问题，如果内容不足以回答，请直接说明。\n\n知识库内容：\n{context_text}")
            knowledge_chunks = chunks

        if system_parts:
            full_system_prompt = "\n\n".join(system_parts)
            messages.append(ChatMessage(role="system", content=full_system_prompt))

        if request.history:
            for msg in request.history:
                role = msg.get("role", "user")
                content = msg.get("content", "")
                if role and content:
                    messages.append(ChatMessage(role=role, content=content))

        messages.append(ChatMessage(role="user", content=request.input))

        return messages, knowledge_chunks

    async def _retrieve_knowledge(self, knowledge_base_id: str, query: str, tenant_id: str, strategy: str = "simple"):
        """检索知识库"""
        try:
            if strategy != "simple":
                provider = await self._get_langchain_provider(tenant_id)
                llm = provider.get_llm() if provider else None

                retriever = RAGRetrieverFactory.create_retriever(
                    knowledge_base_id=knowledge_base_id,
                    tenant_id=tenant_id,
                    llm=llm,
                    retrieval_strategy=strategy
                )

                docs = await retriever.agetrieve(query)

                context_parts = []
                for i, doc in enumerate(docs, 1):
                    score = doc.metadata.get("score", 0)
                    context_parts.append(f"[文档{i}] (相关度: {score:.2f})\n{doc.page_content}")

                chunks = []
                for doc in docs:
                    chunks.append({
                        "content": doc.page_content,
                        "score": doc.metadata.get("score", 0),
                        "metadata": doc.metadata
                    })

                return "\n\n---\n\n".join(context_parts), chunks

            chunks = await self.knowledge_retriever.retrieve(
                knowledge_base_id=knowledge_base_id,
                query=query,
                tenant_id=tenant_id
            )
            if not chunks:
                return "", []

            context_parts = []
            for i, chunk in enumerate(chunks, 1):
                context_parts.append(f"[片段{i}] (相关度: {chunk.get('score', 0):.2f})\n{chunk.get('content', '')}")

            return "\n\n---\n\n".join(context_parts), chunks
        except Exception as e:
            logger.error(f"Retrieve knowledge failed: {e}")
            return "", []

    async def _get_langchain_provider(self, tenant_id: str):
        """获取 LangChain provider"""
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


chat_service = ChatService()