"""文档向量化触发 API"""
import logging
import os
import time
from typing import Optional
from urllib.request import urlopen, Request

from fastapi import APIRouter
from pydantic import BaseModel

from tasks.celery_app import celery_app

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/internal", tags=["documents"])

# 下载文件存放目录（相对 ai-engine 工作目录）
DOWNLOAD_DIR = os.getenv("AI_ENGINE_DOC_DIR", "data/documents")


class VectorizeRequest(BaseModel):
    doc_id: str
    file_url: str          # presigned HTTP URL（由 agent-service GetURL 生成）
    knowledge_base_id: str
    tenant_id: str
    file_name: Optional[str] = "document"


def _download(url: str, dest: str) -> str:
    """下载远程文件到本地，返回本地路径"""
    os.makedirs(os.path.dirname(dest), exist_ok=True)
    req = Request(url, headers={"User-Agent": "fettle-ai-engine/1.0"})
    with urlopen(req, timeout=300) as resp, open(dest, "wb") as f:
        f.write(resp.read())
    return dest


@router.post("/vectorize")
async def trigger_vectorize(req: VectorizeRequest):
    """将文档入队向量化：下载文件到本地 → enqueue celery 任务"""
    try:
        dest_dir = os.path.join(DOWNLOAD_DIR, req.tenant_id, req.knowledge_base_id)
        safe_name = os.path.basename(req.file_name or "document")
        local_path = os.path.join(dest_dir, f"{req.doc_id}_{int(time.time() * 1000)}_{safe_name}")

        try:
            _download(req.file_url, local_path)
        except Exception as e:
            logger.error("download failed: url=%s err=%s", req.file_url, e)
            return {"code": 5001, "message": f"download failed: {e}", "data": None}

        metadata = {
            "knowledge_base_id": req.knowledge_base_id,
            "tenant_id": req.tenant_id,
            "document_id": req.doc_id,
            "source": req.file_name,
        }
        celery_app.send_task(
            "tasks.vectorize_document",
            args=[local_path, metadata],
        )
        logger.info("vectorize queued: doc=%s path=%s", req.doc_id, local_path)
        return {"code": 0, "message": "queued", "data": {"doc_id": req.doc_id, "status": "processing"}}
    except Exception as e:
        logger.exception("trigger vectorize failed")
        return {"code": 5000, "message": str(e), "data": None}
