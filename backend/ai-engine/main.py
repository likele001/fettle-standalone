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
from core.workflow.scheduler import workflow_scheduler

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

app.include_router(health_router, tags=["health"])
app.include_router(workflow_router, tags=["workflows"])
app.include_router(webhook_router, tags=["webhook"])


# gRPC 服务（后台启动）
def _start_grpc_server():
    from concurrent import futures
    import grpc
    from proto import ai_engine_pb2_grpc
    from grpc_server.server import AIEngineServicer

    grpc_server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    ai_engine_pb2_grpc.add_AIEngineServicer_to_server(AIEngineServicer(), grpc_server)
    grpc_server.add_insecure_port(f'[::]:{settings.app_grpc_port}')
    logger.info(f'gRPC server starting on port {settings.app_grpc_port}')
    grpc_server.start()
    grpc_server.wait_for_termination()


@app.on_event("startup")
async def startup():
    await workflow_scheduler.start()

    # 启动 gRPC 服务
    try:
        t = threading.Thread(target=_start_grpc_server, daemon=True)
        t.start()
        logger.info(f"gRPC server started on port {settings.app_grpc_port}")
    except Exception as e:
        logger.warning(f"Failed to start gRPC server: {e}")


@app.on_event("shutdown")
async def shutdown():
    await workflow_scheduler.stop()


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
        host="0.0.0.0",
        port=settings.app_http_port,
        reload=settings.debug,
        log_level="info"
    )
