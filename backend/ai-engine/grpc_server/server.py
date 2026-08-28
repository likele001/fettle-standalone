"""gRPC 服务实现"""
import json
import logging
import time
from typing import Dict, Any

import grpc
from proto import ai_engine_pb2, ai_engine_pb2_grpc
from core.models.model_router import model_router, TaskType
from core.models.db_model_router import db_model_router
from core.models.providers import ProviderFactory, ChatMessage
from core.rag.knowledge_retriever import KnowledgeRetriever
from core.memory.customer_memory import customer_memory

logger = logging.getLogger(__name__)

async def _select_provider_and_model(tenant_id: str, task_type, tenant_plan, model_id):
    """优先用 db_model_router (租户配置) 选 model；失败时 fallback 到 model_router。

    Returns: (provider_code, model_name, source)  source = "db" | "yaml"
    """
    # 1. 优先 db_model_router（读 tenant_ai_configs + ai_models）
    if tenant_id:
        try:
            cfg = await db_model_router.select_model(tenant_id=tenant_id, model_id=model_id)
            return cfg.provider_code, cfg.model_code, "db"
        except Exception as e:
            logger.warning(f"db_model_router.select_model failed for tenant={tenant_id}: {e}")

    # 2. Fallback model_router (yaml)
    cfg = model_router.select_model(
        task_type=task_type, tenant_plan=tenant_plan, model_id=model_id
    )
    return cfg.provider, cfg.model_name, "yaml"



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
            provider = await ProviderFactory.get_provider_async(
                model_config.provider, tenant_id=request.tenant_id
            )
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
            
            provider = await ProviderFactory.get_provider_async(
                model_config.provider, tenant_id=request.tenant_id
            )
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
            provider_code, model_name, _src = await _select_provider_and_model(
                tenant_id=request.tenant_id,
                task_type=TaskType.KNOWLEDGE_QA,
                tenant_plan=request.context.get('tenant_plan'),
                model_id=request.model if request.model else None,
            )
            logger.info(f"grpc GenerateReply: provider={provider_code}, model={model_name}, source={_src}")

            provider = await ProviderFactory.get_provider_async(
                provider_code, tenant_id=request.tenant_id
            )
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

            # 人设约束：从 context 取 system_prompt，作为第一条 system 消息
            try:
                sys_prompt = (request.context or {}).get('system_prompt', '')
                if sys_prompt:
                    messages.insert(0, ChatMessage(role="system", content=sys_prompt))
                    logger.info(f"Persona injected: {sys_prompt[:50]}")
            except Exception as e:
                logger.error(f"Persona injection failed: {e}")

            # 客户记忆注入：跨会话记住客户
            try:
                mem = await customer_memory.get_memory(request.tenant_id, request.user_id)
                if mem:
                    messages.insert(0, ChatMessage(
                        role="system",
                        content=f"以下是该客户的历史记忆（跨会话），请据此提供更贴心、更有连贯性的服务：\n{mem}"
                    ))
                    logger.info(f"Customer memory injected: {len(mem)} chars")
                else:
                    mem = ""
            except Exception as e:
                logger.error(f"Customer memory injection failed: {e}")
                mem = ""

            # 知识库检索（RAG）：从 context 取 knowledge_base_id，检索结果以 system 消息注入
            try:
                req_ctx = request.context or {}
                kb_id = req_ctx.get('knowledge_base_id', '')
                if kb_id:
                    chunks = await self.knowledge_retriever.retrieve(
                        knowledge_base_id=kb_id,
                        query=request.input,
                        tenant_id=request.tenant_id,
                    )
                    if chunks:
                        context_parts = [
                            f"[片段{i}] (相关度: {c.get('score', 0):.2f})\n{c.get('content', '')}"
                            for i, c in enumerate(chunks, 1)
                        ]
                        context_text = "\n\n---\n\n".join(context_parts)
                        messages.insert(0, ChatMessage(
                            role="system",
                            content=f"以下是知识库检索到的相关资料，请优先基于这些资料回答用户问题，不要编造知识库没有的内容：\n\n{context_text}"
                        ))
                        logger.info(f"RAG: injected {len(chunks)} chunks from kb={kb_id}")
                    else:
                        logger.warning(f"RAG: no chunks retrieved for kb={kb_id}")
            except Exception as e:
                logger.error(f"RAG injection failed: {e}")
            
            # 调用模型
            response = await provider.chat(
                messages=messages,
                model=model_name,
                max_tokens=1000,
                temperature=0.7
            )
            
            latency_ms = (time.time() - start_time) * 1000

            # 异步更新客户摘要记忆（不阻塞回复）
            try:
                import asyncio
                mem_msgs = [{"role": m.role, "content": m.content} for m in request.history]
                mem_msgs.append({"role": "user", "content": request.input})
                mem_msgs.append({"role": "assistant", "content": response.content})
                asyncio.create_task(customer_memory.update_memory(
                    request.tenant_id, request.user_id, mem_msgs, mem or ""))
            except Exception as e:
                logger.error(f"customer memory update schedule failed: {e}")

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
            
            provider_code, model_name, _src = await _select_provider_and_model(
                tenant_id=request.tenant_id,
                task_type=TaskType.KNOWLEDGE_QA,
                tenant_plan=request.context.get('tenant_plan'),
                model_id=request.model if request.model else None,
            )
            logger.info(f"grpc GenerateReply: provider={provider_code}, model={model_name}, source={_src}")

            provider = await ProviderFactory.get_provider_async(
                provider_code, tenant_id=request.tenant_id
            )
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

            # 人设约束（StreamReply 同步补上）
            try:
                sys_prompt = (request.context or {}).get('system_prompt', '')
                if sys_prompt:
                    messages.insert(0, ChatMessage(role="system", content=sys_prompt))
            except Exception as e:
                logger.error(f"Stream persona injection failed: {e}")

            # 客户记忆注入
            try:
                mem = await customer_memory.get_memory(request.tenant_id, request.user_id)
                if mem:
                    messages.insert(0, ChatMessage(
                        role="system",
                        content=f"以下是该客户的历史记忆（跨会话），请据此提供更贴心、更有连贯性的服务：\n{mem}"
                    ))
                    logger.info(f"Stream customer memory injected: {len(mem)} chars")
                else:
                    mem = ""
            except Exception as e:
                logger.error(f"Stream customer memory injection failed: {e}")
                mem = ""

            # 流式调用
            tokens_used = 0
            full_reply = ""
            async for chunk in provider.stream_chat(
                messages=messages,
                model=model_name,
                max_tokens=1000,
                temperature=0.7
            ):
                tokens_used += 1  # 简化计算
                full_reply += chunk
                yield ai_engine_pb2.StreamResponse(
                    chunk=chunk,
                    is_final=False,
                    model_used=model_name.model_name,
                    tokens_used=tokens_used
                )
            
            # 异步更新客户摘要记忆
            try:
                import asyncio
                mem_msgs = [{"role": m.role, "content": m.content} for m in request.history]
                mem_msgs.append({"role": "user", "content": request.input})
                mem_msgs.append({"role": "assistant", "content": full_reply})
                asyncio.create_task(customer_memory.update_memory(
                    request.tenant_id, request.user_id, mem_msgs, mem or ""))
            except Exception as e:
                logger.error(f"stream customer memory update schedule failed: {e}")

            # 最终响应
            yield ai_engine_pb2.StreamResponse(
                chunk="",
                is_final=True,
                model_used=model_name.model_name,
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
            
            provider = await ProviderFactory.get_provider_async(
                emb_config.provider, tenant_id=request.tenant_id
            )
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
