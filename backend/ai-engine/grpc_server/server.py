"""gRPC 服务实现"""
import json
import logging
import time
from typing import Dict, Any

import grpc
from proto import ai_engine_pb2, ai_engine_pb2_grpc
from core.models.model_router import model_router, TaskType
from core.models.providers import ProviderFactory, ChatMessage
from core.rag.knowledge_retriever import KnowledgeRetriever

logger = logging.getLogger(__name__)


class AIEngineServicer(ai_engine_pb2_grpc.AIEngineServicer):

    def __init__(self):
        self.knowledge_retriever = KnowledgeRetriever()
    """AI 引擎 gRPC 服务"""
    
    async def RecognizeIntent(
        self,
        request: ai_engine_pb2.IntentRequest,
        context: grpc.ServicerContext
    ) -> ai_engine_pb2.IntentResponse:
        """意图识别"""
        try:
            logger.info(f"Intent recognition request: tenant={request.tenant_id}, input={request.input[:50]}")
            
            # 获取模型
            model_config = model_router.select_model(
                task_type=TaskType.INTENT_RECOGNITION,
                tenant_plan=request.context.get('tenant_plan')
            )
            
            # 获取提供商
            provider = ProviderFactory.get_provider(model_config.provider)
            if not provider:
                return ai_engine_pb2.IntentResponse(
                    intent="unknown",
                    confidence=0.0,
                    raw_response="Provider not available"
                )
            
            # 构建提示
            messages = [
                ChatMessage(role="system", content="""你是一个意图识别助手。分析用户输入，识别意图和实体。
返回JSON格式：{"intent": "意图", "confidence": 0.95, "entities": {"key": "value"}}
支持的意图：greeting, farewell, question, complaint, request, chitchat, unknown"""),
                ChatMessage(role="user", content=request.input)
            ]
            
            # 调用模型
            response = await provider.chat(
                messages=messages,
                model=model_config.model_name,
                max_tokens=500,
                temperature=0.3
            )
            
            # 解析响应（简化处理）
            try:
                result = json.loads(response.content)
                return ai_engine_pb2.IntentResponse(
                    intent=result.get('intent', 'unknown'),
                    confidence=result.get('confidence', 0.5),
                    entities=result.get('entities', {}),
                    raw_response=response.content
                )
            except (json.JSONDecodeError, KeyError, ValueError):
                return ai_engine_pb2.IntentResponse(
                    intent="unknown",
                    confidence=0.0,
                    raw_response=response.content
                )
        
        except Exception as e:
            logger.error(f"Intent recognition failed: {e}")
            return ai_engine_pb2.IntentResponse(
                intent="error",
                confidence=0.0,
                raw_response=str(e)
            )
    
    async def PlanTask(
        self,
        request: ai_engine_pb2.PlanRequest,
        context: grpc.ServicerContext
    ) -> ai_engine_pb2.PlanResponse:
        """任务规划"""
        try:
            logger.info(f"Task planning request: tenant={request.tenant_id}, goal={request.goal[:50]}")
            
            model_config = model_router.select_model(
                task_type=TaskType.TASK_PLANNING,
                tenant_plan=request.context.get('tenant_plan')
            )
            
            provider = ProviderFactory.get_provider(model_config.provider)
            if not provider:
                return ai_engine_pb2.PlanResponse(
                    steps=[],
                    reasoning="Provider not available"
                )
            
            # 构建提示
            tools_str = ", ".join(request.available_tools) if request.available_tools else "none"
            messages = [
                ChatMessage(role="system", content=f"""你是一个任务规划助手。将目标分解为可执行的步骤。
可用工具：{tools_str}
返回JSON格式：{{"steps": [{{"step_id": "1", "action": "动作", "tool": "工具名", "parameters": {{}}, "description": "描述"}}]}}"""),
                ChatMessage(role="user", content=request.goal)
            ]
            
            response = await provider.chat(
                messages=messages,
                model=model_config.model_name,
                max_tokens=1000,
                temperature=0.5
            )
            
            # 解析响应
            try:
                result = json.loads(response.content)
                steps = []
                for step_data in result.get('steps', []):
                    steps.append(ai_engine_pb2.TaskStep(
                        step_id=step_data.get('step_id', ''),
                        action=step_data.get('action', ''),
                        tool=step_data.get('tool', ''),
                        parameters=step_data.get('parameters', {}),
                        description=step_data.get('description', '')
                    ))
                
                return ai_engine_pb2.PlanResponse(
                    steps=steps,
                    reasoning=result.get('reasoning', '')
                )
            except (json.JSONDecodeError, KeyError, ValueError):
                return ai_engine_pb2.PlanResponse(
                    steps=[],
                    reasoning=response.content
                )
        
        except Exception as e:
            logger.error(f"Task planning failed: {e}")
            return ai_engine_pb2.PlanResponse(
                steps=[],
                reasoning=str(e)
            )
    
    async def GenerateReply(
        self,
        request: ai_engine_pb2.ReplyRequest,
        context: grpc.ServicerContext
    ) -> ai_engine_pb2.ReplyResponse:
        """回复生成"""
        try:
            start_time = time.time()
            logger.info(f"Reply generation request: tenant={request.tenant_id}, input={request.input[:50]}")
            
            # 选择模型
            model_config = model_router.select_model(
                task_type=TaskType.KNOWLEDGE_QA,
                tenant_plan=request.context.get('tenant_plan'),
                model_id=request.model if request.model else None
            )
            
            provider = ProviderFactory.get_provider(model_config.provider)
            if not provider:
                return ai_engine_pb2.ReplyResponse(
                    reply="服务暂时不可用",
                    model_used="",
                    tokens_used=0,
                    latency_ms=0
                )
            
            # 构建消息
            messages = []
            
            # 添加历史消息
            for msg in request.history:
                messages.append(ChatMessage(role=msg.role, content=msg.content))
            
            # 添加当前输入
            messages.append(ChatMessage(role="user", content=request.input))
            
            # 调用模型
            response = await provider.chat(
                messages=messages,
                model=model_config.model_name,
                max_tokens=model_config.max_tokens,
                temperature=model_config.temperature
            )
            
            latency_ms = (time.time() - start_time) * 1000
            
            return ai_engine_pb2.ReplyResponse(
                reply=response.content,
                model_used=response.model,
                tokens_used=response.tokens_used,
                latency_ms=latency_ms
            )
        
        except Exception as e:
            logger.error(f"Reply generation failed: {e}")
            return ai_engine_pb2.ReplyResponse(
                reply=f"生成回复失败: {str(e)}",
                model_used="",
                tokens_used=0,
                latency_ms=0
            )
    
    async def StreamReply(
        self,
        request: ai_engine_pb2.StreamRequest,
        context: grpc.ServicerContext
    ):
        """流式回复"""
        try:
            logger.info(f"Stream reply request: tenant={request.tenant_id}")
            
            model_config = model_router.select_model(
                task_type=TaskType.KNOWLEDGE_QA,
                tenant_plan=request.context.get('tenant_plan'),
                model_id=request.model if request.model else None
            )
            
            provider = ProviderFactory.get_provider(model_config.provider)
            if not provider:
                yield ai_engine_pb2.StreamResponse(
                    chunk="服务暂时不可用",
                    is_final=True,
                    model_used="",
                    tokens_used=0
                )
                return
            
            # 构建消息
            messages = []
            for msg in request.history:
                messages.append(ChatMessage(role=msg.role, content=msg.content))
            messages.append(ChatMessage(role="user", content=request.input))
            
            # 流式调用
            tokens_used = 0
            async for chunk in provider.stream_chat(
                messages=messages,
                model=model_config.model_name,
                max_tokens=model_config.max_tokens,
                temperature=model_config.temperature
            ):
                tokens_used += 1  # 简化计算
                yield ai_engine_pb2.StreamResponse(
                    chunk=chunk,
                    is_final=False,
                    model_used=model_config.model_name,
                    tokens_used=tokens_used
                )
            
            # 最终响应
            yield ai_engine_pb2.StreamResponse(
                chunk="",
                is_final=True,
                model_used=model_config.model_name,
                tokens_used=tokens_used
            )
        
        except Exception as e:
            logger.error(f"Stream reply failed: {e}")
            yield ai_engine_pb2.StreamResponse(
                chunk=f"生成回复失败: {str(e)}",
                is_final=True,
                model_used="",
                tokens_used=0
            )
    
    async def GenerateEmbedding(
        self,
        request: ai_engine_pb2.EmbeddingRequest,
        context: grpc.ServicerContext
    ) -> ai_engine_pb2.EmbeddingResponse:
        """生成嵌入"""
        try:
            logger.info(f"Embedding request: {len(request.texts)} texts")
            
            emb_config = model_router.get_embedding_model(
                request.model if request.model else None
            )
            
            provider = ProviderFactory.get_provider(emb_config.provider)
            if not provider:
                return ai_engine_pb2.EmbeddingResponse(
                    embeddings=[],
                    model_used="",
                    tokens_used=0
                )
            
            response = await provider.embed(
                texts=list(request.texts),
                model=emb_config.model_name
            )
            
            # 转换为 protobuf
            embeddings = []
            for emb in response.embeddings:
                embeddings.append(ai_engine_pb2.EmbeddingVector(
                    values=emb,
                    dimension=len(emb)
                ))
            
            return ai_engine_pb2.EmbeddingResponse(
                embeddings=embeddings,
                model_used=response.model,
                tokens_used=response.tokens_used
            )
        
        except Exception as e:
            logger.error(f"Embedding generation failed: {e}")
            return ai_engine_pb2.EmbeddingResponse(
                embeddings=[],
                model_used="",
                tokens_used=0
            )
    
    async def RetrieveKnowledge(
        self,
        request: ai_engine_pb2.KnowledgeRequest,
        context: grpc.ServicerContext
    ) -> ai_engine_pb2.KnowledgeResponse:
        """知识检索"""
        try:
            logger.info(f"Knowledge retrieval request: query={request.query[:50]}")
            
            if not request.knowledge_base_id:
                return ai_engine_pb2.KnowledgeResponse(chunks=[], total_found=0)
            
            results = await self.knowledge_retriever.retrieve(
                knowledge_base_id=request.knowledge_base_id,
                query=request.query,
                tenant_id=request.tenant_id,
                top_k=request.top_k or 5,
                score_threshold=request.score_threshold or 0.5
            )
            
            chunks = []
            for r in results:
                chunks.append(ai_engine_pb2.KnowledgeChunk(
                    content=r.get("content", ""),
                    score=r.get("score", 0.0),
                    metadata={k: str(v) for k, v in r.get("metadata", {}).items()},
                    document_id=r.get("document_id", "")
                ))
            
            return ai_engine_pb2.KnowledgeResponse(
                chunks=chunks,
                total_found=len(chunks)
            )
        
        except Exception as e:
            logger.error(f"Knowledge retrieval failed: {e}")
            return ai_engine_pb2.KnowledgeResponse(
                chunks=[],
                total_found=0
            )
