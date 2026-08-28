"""对话服务 - 核心对话逻辑"""
import logging
import time
from typing import List, Dict, Any, Optional, AsyncIterator
from dataclasses import dataclass

from core.models.model_router import model_router, TaskType
from core.models.db_model_router import db_model_router
from core.models.providers import ProviderFactory, ChatMessage
from core.rag.knowledge_retriever import KnowledgeRetriever
from core.rag.advanced_retrievers import RAGRetrieverFactory
from core.workflow import Workflow, WorkflowInstance, workflow_engine

logger = logging.getLogger(__name__)

async def _select_provider_and_model(tenant_id: str, task_type, tenant_plan, model_id):
    """优先用 db_model_router (租户配置) 选 model；失败时 fallback 到 model_router；
    最终 fallback 扫描租户所有 api_key 找有 key 的厂商的默认 model。

    Returns: (provider_code, model_name, max_tokens, temperature, source)
             source = "db" | "yaml" | "tenant_key"
    """
    # 1. 优先 db_model_router（读 tenant_ai_configs + ai_models）
    db_cfg = None
    if tenant_id:
        try:
            db_cfg = await db_model_router.select_model(tenant_id=tenant_id, model_id=model_id)
            if db_cfg and db_cfg.provider_code:
                return db_cfg.provider_code, db_cfg.model_code, db_cfg.max_output_tokens or 2000, 0.7, "db"
        except Exception as e:
            logger.warning(f"db_model_router.select_model failed for tenant={tenant_id}: {e}")

    # 2. Fallback: 扫租户所有 tenant_api_keys，找第一个有 key 的厂商，再取该厂商的默认 model
    if tenant_id:
        try:
            from core.models.db_config import ai_config_db
            providers = await ai_config_db.list_providers()
            for prov in providers:
                tk = await ai_config_db.get_tenant_api_key(tenant_id, prov.id)
                if tk and tk.api_key_value:
                    # 找该 provider_code 下的默认 chat model
                    models = await ai_config_db.list_models(provider_id=prov.id, model_type="chat")
                    default_m = next((m for m in models if m.is_default), None) or (models[0] if models else None)
                    if default_m:
                        logger.info(f"Using tenant_key fallback: provider={prov.code}, model={default_m.model_code}")
                        return prov.code, default_m.model_code, default_m.max_output_tokens or 2000, 0.7, "tenant_key"
        except Exception as e:
            logger.warning(f"tenant_key fallback failed for tenant={tenant_id}: {e}")

    # 3. 最后 fallback: model_router (yaml)
    cfg = model_router.select_model(
        task_type=task_type, tenant_plan=tenant_plan, model_id=model_id
    )
    return cfg.provider, cfg.model_name, cfg.max_tokens, cfg.temperature, "yaml"



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
    image_paths: List[str] = None
    audio_paths: List[str] = None


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
        provider_code, model_name, max_tokens, temperature, _src = await _select_provider_and_model(
            tenant_id=request.tenant_id,
            task_type=TaskType.KNOWLEDGE_QA if request.knowledge_base_id else TaskType.SIMPLE_CHAT,
            tenant_plan=request.context.get("tenant_plan") if request.context else None,
            model_id=request.model if request.model else None,
        )
        logger.info(f"chat_service: provider={provider_code}, model={model_name}, source={_src}")

        provider = await ProviderFactory.get_provider_async(
            provider_code, tenant_id=request.tenant_id
        )
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
            model=model_name,
            max_tokens=max_tokens,
            temperature=temperature
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
        provider_code, model_name, max_tokens, temperature, _src = await _select_provider_and_model(
            tenant_id=request.tenant_id,
            task_type=TaskType.KNOWLEDGE_QA if request.knowledge_base_id else TaskType.SIMPLE_CHAT,
            tenant_plan=request.context.get("tenant_plan") if request.context else None,
            model_id=request.model if request.model else None,
        )
        logger.info(f"chat_service: provider={provider_code}, model={model_name}, source={_src}")

        provider = await ProviderFactory.get_provider_async(
            provider_code, tenant_id=request.tenant_id
        )
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
            model=model_name,
            max_tokens=max_tokens,
            temperature=temperature
        ):
            tokens_used += 1
            full_content.append(chunk)
            yield {
                "chunk": chunk,
                "is_final": False,
                "model_used": model_name,
                "tokens_used": tokens_used
            }

        yield {
            "chunk": "",
            "is_final": True,
            "model_used": model_name,
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

        media_text = await self._extract_media_text(request)
        if media_text:
            messages[-1] = ChatMessage(role="user", content=request.input + "\n\n" + media_text)

        return messages, knowledge_chunks

    async def _extract_media_text(self, request) -> str:
        """提取图片(OCR)/语音(转写)文本，注入用户消息（依赖缺失时优雅跳过）。"""
        parts = []
        try:
            if getattr(request, "image_paths", None):
                from core.multimodal.image_processor import image_processor
                for p in request.image_paths:
                    with open(p, "rb") as f:
                        data = f.read()
                    r = await image_processor.process_image(data, extract_text=True)
                    t = r.get("extracted_text", "")
                    if t.strip():
                        parts.append(f"[图片内容 {p}]:\n{t}")
        except Exception as e:
            logger.error(f"Image OCR failed: {e}")
        try:
            if getattr(request, "audio_paths", None):
                from core.multimodal.audio_processor import audio_processor
                for p in request.audio_paths:
                    with open(p, "rb") as f:
                        data = f.read()
                    r = await audio_processor.process_audio(data, filename=p, transcribe=True)
                    t = r.get("transcription", "")
                    if t.strip():
                        parts.append(f"[语音转写 {p}]:\n{t}")
        except Exception as e:
            logger.error(f"Audio transcribe failed: {e}")
        return "\n\n".join(parts)

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