"""对话服务 - 使用数据库配置"""
import logging
import time
from typing import List, Dict, Any, Optional, AsyncIterator
from dataclasses import dataclass

from core.models.db_model_router import db_model_router, ModelConfig
from core.models.providers.db_factory import DBProviderFactory
from core.models.db_config import ai_config_db
from core.models.providers import ChatMessage
from core.rag.knowledge_retriever import KnowledgeRetriever

logger = logging.getLogger(__name__)


@dataclass
class ChatRequest:
    """对话请求"""
    tenant_id: str
    user_id: str
    input: str
    agent_id: str = ""
    conversation_id: str = ""
    message_id: str = ""
    history: List[Dict[str, str]] = None
    system_prompt: str = ""
    model: str = ""
    knowledge_base_id: str = ""
    context: Dict[str, str] = None
    image_paths: List[str] = None
    audio_paths: List[str] = None


@dataclass
class ChatResponse:
    """对话响应"""
    reply: str
    model_used: str
    model_id: str
    provider_id: str
    provider_code: str
    tokens_used: int
    input_tokens: int
    output_tokens: int
    latency_ms: float
    knowledge_chunks: List[Dict[str, Any]] = None


class DBChatService:
    """数据库驱动的对话服务"""

    def __init__(self):
        self.knowledge_retriever = KnowledgeRetriever()

    async def generate_reply(self, request: ChatRequest) -> ChatResponse:
        """生成回复"""
        start_time = time.time()
        logger.info(f"Generate reply: tenant={request.tenant_id}, input={request.input[:50]}")

        try:
            # 1. 获取租户配置
            tenant_config = await db_model_router.get_tenant_config(request.tenant_id)
            if not tenant_config or not tenant_config.ai_enabled:
                logger.warning(f"AI not enabled for tenant: {request.tenant_id}")
                await self._log_usage(request, 0, 0, False, "AI not enabled")
                return ChatResponse(
                    reply="AI 服务未启用，请联系管理员配置。",
                    model_used="",
                    model_id="",
                    provider_id="",
                    provider_code="",
                    tokens_used=0,
                    input_tokens=0,
                    output_tokens=0,
                    latency_ms=(time.time() - start_time) * 1000
                )

            # 2. 选择模型
            model_config = await db_model_router.select_model(
                tenant_id=request.tenant_id,
                model_id=request.model if request.model else None,
            )

            # 3. 获取租户 API Key
            tenant_api_key = await db_model_router.get_tenant_api_key(
                request.tenant_id,
                model_config.provider_id,
            )

            # 4. 创建 Provider
            provider = await DBProviderFactory.get_provider(
                model_config.provider_code,
                tenant_api_key=tenant_api_key,
            )
            if not provider:
                await self._log_usage(request, 0, 0, False, "No provider available")
                return ChatResponse(
                    reply="AI 服务暂时不可用，请检查 API Key 配置。",
                    model_used="",
                    model_id="",
                    provider_id="",
                    provider_code="",
                    tokens_used=0,
                    input_tokens=0,
                    output_tokens=0,
                    latency_ms=(time.time() - start_time) * 1000
                )

            # 5. 构建消息
            messages, knowledge_chunks = await self._build_messages(request)

            # 6. 调用模型
            response = await provider.chat(
                messages=messages,
                model=model_config.model_name,
                max_tokens=model_config.max_output_tokens,
                temperature=0.7,
            )

            # 7. 计算费用
            input_tokens = response.tokens_used // 2
            output_tokens = response.tokens_used - input_tokens
            input_cost = (input_tokens / 1000) * model_config.input_price_per_1k
            output_cost = (output_tokens / 1000) * model_config.output_price_per_1k

            latency_ms = (time.time() - start_time) * 1000

            # 8. 记录使用日志
            await self._log_usage(
                request,
                input_tokens,
                output_tokens,
                True,
                None,
                model_config=model_config,
                latency_ms=latency_ms,
                input_cost=input_cost,
                output_cost=output_cost,
            )

            return ChatResponse(
                reply=response.content,
                model_used=model_config.model_name,
                model_id=model_config.model_id,
                provider_id=model_config.provider_id,
                provider_code=model_config.provider_code,
                tokens_used=response.tokens_used,
                input_tokens=input_tokens,
                output_tokens=output_tokens,
                latency_ms=latency_ms,
                knowledge_chunks=knowledge_chunks
            )

        except Exception as e:
            logger.error(f"Generate reply failed: {e}")
            await self._log_usage(request, 0, 0, False, str(e))
            return ChatResponse(
                reply=f"生成回复失败：{str(e)}",
                model_used="",
                model_id="",
                provider_id="",
                provider_code="",
                tokens_used=0,
                input_tokens=0,
                output_tokens=0,
                latency_ms=(time.time() - start_time) * 1000
            )

    async def stream_reply(self, request: ChatRequest) -> AsyncIterator[Dict[str, Any]]:
        """流式生成回复"""
        start_time = time.time()
        logger.info(f"Stream reply: tenant={request.tenant_id}, input={request.input[:50]}")

        try:
            # 1. 获取租户配置
            tenant_config = await db_model_router.get_tenant_config(request.tenant_id)
            if not tenant_config or not tenant_config.ai_enabled:
                yield {
                    "chunk": "AI 服务未启用，请联系管理员配置。",
                    "is_final": True,
                    "model_used": "",
                    "tokens_used": 0
                }
                return

            # 2. 选择模型
            model_config = await db_model_router.select_model(
                tenant_id=request.tenant_id,
                model_id=request.model if request.model else None,
            )

            # 3. 获取租户 API Key
            tenant_api_key = await db_model_router.get_tenant_api_key(
                request.tenant_id,
                model_config.provider_id,
            )

            # 4. 创建 Provider
            provider = await DBProviderFactory.get_provider(
                model_config.provider_code,
                tenant_api_key=tenant_api_key,
            )
            if not provider:
                yield {
                    "chunk": "AI 服务暂时不可用，请检查 API Key 配置。",
                    "is_final": True,
                    "model_used": "",
                    "tokens_used": 0
                }
                return

            # 5. 构建消息
            messages, knowledge_chunks = await self._build_messages(request)

            # 6. 流式调用
            tokens_used = 0
            full_content = []

            async for chunk in provider.stream_chat(
                messages=messages,
                model=model_config.model_name,
                max_tokens=model_config.max_output_tokens,
                temperature=0.7,
            ):
                tokens_used += 1
                full_content.append(chunk)
                yield {
                    "chunk": chunk,
                    "is_final": False,
                    "model_used": model_config.model_name,
                    "model_id": model_config.model_id,
                    "provider_code": model_config.provider_code,
                    "tokens_used": tokens_used
                }

            # 7. 计算费用
            input_tokens = tokens_used // 3
            output_tokens = tokens_used - input_tokens
            input_cost = (input_tokens / 1000) * model_config.input_price_per_1k
            output_cost = (output_tokens / 1000) * model_config.output_price_per_1k

            latency_ms = (time.time() - start_time) * 1000

            # 8. 记录使用日志
            await self._log_usage(
                request,
                input_tokens,
                output_tokens,
                True,
                None,
                model_config=model_config,
                latency_ms=latency_ms,
                input_cost=input_cost,
                output_cost=output_cost,
            )

            yield {
                "chunk": "",
                "is_final": True,
                "model_used": model_config.model_name,
                "model_id": model_config.model_id,
                "provider_code": model_config.provider_code,
                "tokens_used": tokens_used,
                "knowledge_chunks": knowledge_chunks
            }

        except Exception as e:
            logger.error(f"Stream reply failed: {e}")
            await self._log_usage(request, 0, 0, False, str(e))
            yield {
                "chunk": f"生成回复失败：{str(e)}",
                "is_final": True,
                "model_used": "",
                "tokens_used": 0
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
                tenant_id=request.tenant_id
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

        user_content = request.input
        media_text = await self._extract_media_text(request)
        if media_text:
            user_content = request.input + "\n\n" + media_text
        messages.append(ChatMessage(role="user", content=user_content))

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

    async def _retrieve_knowledge(self, knowledge_base_id: str, query: str, tenant_id: str):
        """检索知识库"""
        try:
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

    async def _log_usage(
        self,
        request: ChatRequest,
        input_tokens: int,
        output_tokens: int,
        success: bool,
        error_message: Optional[str],
        model_config: Optional[ModelConfig] = None,
        latency_ms: float = 0,
        input_cost: float = 0,
        output_cost: float = 0,
    ):
        """记录使用日志"""
        try:
            log_data = {
                'tenant_id': request.tenant_id,
                'user_id': request.user_id,
                'conversation_id': request.conversation_id,
                'message_id': request.message_id,
                'provider_id': model_config.provider_id if model_config else None,
                'model_id': model_config.model_id if model_config else None,
                'input_tokens': input_tokens,
                'output_tokens': output_tokens,
                'total_tokens': input_tokens + output_tokens,
                'input_cost': input_cost,
                'output_cost': output_cost,
                'total_cost': input_cost + output_cost,
                'request_type': 'chat',
                'latency_ms': int(latency_ms),
                'success': success,
                'error_message': error_message,
            }
            await ai_config_db.create_usage_log(log_data)
        except Exception as e:
            logger.error(f"Failed to log usage: {e}")


# 全局数据库对话服务实例
db_chat_service = DBChatService()