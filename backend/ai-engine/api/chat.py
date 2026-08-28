"""对话 API - 使用数据库配置"""
import json
import logging
from typing import List, Dict, Any, Optional
from pydantic import BaseModel, Field

from fastapi import APIRouter, Query
from fastapi.responses import StreamingResponse

from core.db_chat_service import db_chat_service, ChatRequest
from core.models.db_config import ai_config_db

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/v1/chat")


class ChatMessageSchema(BaseModel):
    role: str
    content: str


class GenerateRequest(BaseModel):
    tenant_id: str
    user_id: str
    input: str
    agent_id: str = ""
    conversation_id: str = ""
    message_id: str = ""
    history: List[ChatMessageSchema] = Field(default_factory=list)
    system_prompt: str = ""
    model: str = ""
    knowledge_base_id: str = ""
    context: Dict[str, str] = Field(default_factory=dict)
    image_paths: List[str] = Field(default_factory=list)
    audio_paths: List[str] = Field(default_factory=list)


class GenerateResponse(BaseModel):
    reply: str
    model_used: str
    model_id: str
    provider_id: str
    provider_code: str
    tokens_used: int
    input_tokens: int
    output_tokens: int
    latency_ms: float
    knowledge_chunks: List[Dict[str, Any]] = Field(default_factory=list)


@router.post("/generate", response_model=GenerateResponse)
async def generate_chat(request: GenerateRequest):
    """生成对话回复"""
    # 技能桥接
    try:
        from core.tools.skill_bridge import ensure_tenant_skills
        await ensure_tenant_skills(request.tenant_id)
    except Exception:
        pass
    chat_req = ChatRequest(
        tenant_id=request.tenant_id,
        user_id=request.user_id,
        input=request.input,
        agent_id=request.agent_id,
        conversation_id=request.conversation_id,
        message_id=request.message_id,
        history=[{"role": m.role, "content": m.content} for m in request.history],
        system_prompt=request.system_prompt,
        model=request.model,
        knowledge_base_id=request.knowledge_base_id,
        context=request.context,
        image_paths=request.image_paths,
        audio_paths=request.audio_paths
    )

    result = await db_chat_service.generate_reply(chat_req)

    return GenerateResponse(
        reply=result.reply,
        model_used=result.model_used,
        model_id=result.model_id,
        provider_id=result.provider_id,
        provider_code=result.provider_code,
        tokens_used=result.tokens_used,
        input_tokens=result.input_tokens,
        output_tokens=result.output_tokens,
        latency_ms=result.latency_ms,
        knowledge_chunks=result.knowledge_chunks or []
    )


@router.post("/stream")
async def stream_chat(request: GenerateRequest):
    """流式生成对话回复（SSE）"""
    # 技能桥接
    try:
        from core.tools.skill_bridge import ensure_tenant_skills
        await ensure_tenant_skills(request.tenant_id)
    except Exception:
        pass
    chat_req = ChatRequest(
        tenant_id=request.tenant_id,
        user_id=request.user_id,
        input=request.input,
        agent_id=request.agent_id,
        conversation_id=request.conversation_id,
        message_id=request.message_id,
        history=[{"role": m.role, "content": m.content} for m in request.history],
        system_prompt=request.system_prompt,
        model=request.model,
        knowledge_base_id=request.knowledge_base_id,
        context=request.context,
        image_paths=request.image_paths,
        audio_paths=request.audio_paths
    )

    async def event_generator():
        async for chunk in db_chat_service.stream_reply(chat_req):
            data = json.dumps(chunk, ensure_ascii=False)
            yield f"data: {data}\n\n"

    return StreamingResponse(
        event_generator(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        }
    )


class UsageLogRequest(BaseModel):
    tenant_id: str
    page: int = 1
    page_size: int = 50
    start_date: Optional[str] = None
    end_date: Optional[str] = None


@router.get("/usage/logs")
async def get_usage_logs(
    tenant_id: str = Query(...),
    page: int = Query(1, ge=1),
    page_size: int = Query(50, ge=1, le=200),
    start_date: Optional[str] = Query(None),
    end_date: Optional[str] = Query(None),
):
    """获取使用日志"""
    logs, total = await ai_config_db.get_usage_logs(tenant_id, page, page_size, start_date, end_date)
    return {
        "code": 0,
        "message": "success",
        "data": {
            "items": logs,
            "total": total,
            "page": page,
            "page_size": page_size,
        },
    }


@router.get("/usage/stats")
async def get_usage_stats(
    tenant_id: str = Query(...),
    start_date: str = Query(...),
    end_date: str = Query(...),
):
    """获取使用统计"""
    stats = await ai_config_db.get_usage_stats(tenant_id, start_date, end_date)
    return {
        "code": 0,
        "message": "success",
        "data": stats,
    }