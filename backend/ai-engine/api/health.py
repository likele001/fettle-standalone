"""健康检查 API"""
import asyncio
import logging

from fastapi import APIRouter
from datetime import datetime

from config.settings import settings

logger = logging.getLogger(__name__)

router = APIRouter()


@router.get("/health")
async def health_check():
    """健康检查"""
    return {
        "status": "ok",
        "service": "ai-engine",
        "timestamp": datetime.now().timestamp()
    }


async def check_database() -> str:
    """检查 PostgreSQL 连接"""
    try:
        import asyncpg
        conn = await asyncpg.connect(
            host=settings.db_host,
            port=settings.db_port,
            user=settings.db_user,
            password=settings.db_password,
            database=settings.db_name,
            timeout=2
        )
        await conn.close()
        return "ok"
    except Exception as e:
        logger.warning(f"Database health check failed: {e}")
        return "unhealthy"


async def check_redis() -> str:
    """检查 Redis 连接"""
    try:
        import redis.asyncio as aioredis
        r = aioredis.Redis(
            host=settings.redis_host,
            port=settings.redis_port,
            db=settings.redis_db,
            password=settings.redis_password,
            socket_timeout=2
        )
        await r.ping()
        await r.aclose()
        return "ok"
    except Exception as e:
        logger.warning(f"Redis health check failed: {e}")
        return "unhealthy"


async def check_milvus() -> str:
    """检查 Milvus 连接"""
    try:
        from pymilvus import connections
        connections.connect(host=settings.milvus_host, port=settings.milvus_port, timeout=2)
        connections.disconnect("default")
        return "ok"
    except Exception as e:
        logger.warning(f"Milvus health check failed: {e}")
        return "unhealthy"


async def check_nats() -> str:
    """检查 NATS 连接"""
    try:
        import nats
        nc = await nats.connect(settings.nats_url, timeout=2)
        await nc.close()
        return "ok"
    except Exception as e:
        logger.warning(f"NATS health check failed: {e}")
        return "unhealthy"


@router.get("/ready")
async def readiness_check():
    """就绪检查 - 真实检查所有依赖服务"""
    db_status, redis_status, milvus_status, nats_status = await asyncio.gather(
        check_database(),
        check_redis(),
        check_milvus(),
        check_nats(),
        return_exceptions=True
    )

    checks = {
        "database": db_status if not isinstance(db_status, Exception) else "unhealthy",
        "redis": redis_status if not isinstance(redis_status, Exception) else "unhealthy",
        "milvus": milvus_status if not isinstance(milvus_status, Exception) else "unhealthy",
        "nats": nats_status if not isinstance(nats_status, Exception) else "unhealthy",
    }

    all_healthy = all(v == "ok" for v in checks.values())

    return {
        "status": "ready" if all_healthy else "degraded",
        "service": "ai-engine",
        "checks": checks
    }
