"""AI Engine 主入口"""
import logging
import os
import threading

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

from config.settings import settings
from api.health import router as health_router
from api.workflow import router as workflow_router
from api.webhook import router as webhook_router
from api.chat import router as chat_router
from api.documents import router as documents_router
from api.mcp_servers import router as mcp_servers_router
from core.security import internal_auth_middleware

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


app = FastAPI(
    title="AI Engine",
    description="AI 智能体引擎服务",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=[os.getenv("ALLOWED_ORIGINS", "*")],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
app.middleware("http")(internal_auth_middleware)

app.include_router(health_router, tags=["health"])
app.include_router(workflow_router, tags=["workflows"])
app.include_router(webhook_router, tags=["webhook"])
app.include_router(chat_router, tags=["chat"])
app.include_router(documents_router, tags=["documents"])
app.include_router(mcp_servers_router, tags=["mcp-servers"])


# gRPC 服务（后台启动） - 改用 grpc.aio 以正确 await async coroutine
async def _start_grpc_server():
    import grpc
    from grpc import aio as grpc_aio
    from proto import ai_engine_pb2_grpc
    from grpc_server.server import AIEngineServicer

    grpc_server = grpc_aio.server()
    ai_engine_pb2_grpc.add_AIEngineServicer_to_server(AIEngineServicer(), grpc_server)
    grpc_server.add_insecure_port(f'127.0.0.1:{settings.app_grpc_port}')
    logger.info(f'gRPC (aio) server starting on port {settings.app_grpc_port}')
    await grpc_server.start()
    await grpc_server.wait_for_termination()


@app.on_event("startup")
async def startup():
    try:
        from core.workflow.scheduler import workflow_scheduler
        await workflow_scheduler.start()
    except ImportError as e:
        logger.warning(f"Failed to start workflow scheduler: {e}")

    # 启动 gRPC 服务（async 方式）
    try:
        import asyncio
        asyncio.create_task(_start_grpc_server())
        logger.info(f"gRPC (aio) server task created on port {settings.app_grpc_port}")
    except Exception as e:
        logger.warning(f"Failed to start gRPC server: {e}")


@app.on_event("shutdown")
async def shutdown():
    try:
        from core.workflow.scheduler import workflow_scheduler
        await workflow_scheduler.stop()
    except ImportError:
        pass


@app.get("/")
async def root():
    return {
        "service": "ai-engine",
        "version": "1.0.0",
        "status": "running"
    }


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=os.getenv("AI_ENGINE_BIND_HOST", "127.0.0.1"),
        port=settings.app_http_port,
        reload=settings.debug,
        log_level="info"
    )
