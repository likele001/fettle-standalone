"""RAG 检索器"""
import logging
from typing import List, Dict, Any, Optional
from dataclasses import dataclass

from config.settings import settings

logger = logging.getLogger(__name__)


@dataclass
class RetrievalResult:
    """检索结果"""
    content: str
    score: float
    metadata: Dict[str, Any]
    document_id: str


class MilvusRetriever:
    """Milvus 向量检索器"""
    
    def __init__(
        self,
        collection_name: str = "knowledge_base",
        top_k: int = None,
        score_threshold: float = 0.5
    ):
        self.collection_name = collection_name
        self.top_k = top_k or settings.top_k
        self.score_threshold = score_threshold
        self.client = None
        self.collection = None
    
    async def connect(self):
        """连接 Milvus"""
        try:
            from pymilvus import connections, Collection
            
            connections.connect(
                host=settings.milvus_host,
                port=settings.milvus_port
            )
            
            self.collection = Collection(self.collection_name)
            self.collection.load()
            
            logger.info(f"Connected to Milvus collection: {self.collection_name}")
        
        except Exception as e:
            logger.error(f"Failed to connect to Milvus: {e}")
            raise
    
    async def search(
        self,
        query_embedding: List[float],
        top_k: int = None,
        score_threshold: float = None
    ) -> List[RetrievalResult]:
        """向量检索"""
        try:
            if not self.collection:
                await self.connect()
            
            top_k = top_k or self.top_k
            score_threshold = score_threshold or self.score_threshold
            
            # 搜索参数
            search_params = {
                "metric_type": "COSINE",
                "params": {"nprobe": 10}
            }
            
            # 执行搜索
            results = self.collection.search(
                data=[query_embedding],
                anns_field="embedding",
                param=search_params,
                limit=top_k,
                output_fields=["content", "metadata", "document_id"]
            )
            
            # 处理结果
            retrieval_results = []
            for hits in results:
                for hit in hits:
                    if hit.score >= score_threshold:
                        retrieval_results.append(RetrievalResult(
                            content=hit.entity.get("content"),
                            score=hit.score,
                            metadata=hit.entity.get("metadata", {}),
                            document_id=hit.entity.get("document_id", "")
                        ))
            
            logger.info(f"Retrieved {len(retrieval_results)} results")
            return retrieval_results
        
        except Exception as e:
            logger.error(f"Search failed: {e}")
            return []
    
    async def close(self):
        """关闭连接"""
        try:
            from pymilvus import connections
            connections.disconnect("default")
            logger.info("Disconnected from Milvus")
        except Exception as e:
            logger.error(f"Failed to disconnect from Milvus: {e}")
