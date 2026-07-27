"""知识库检索服务"""
import logging
from typing import List, Dict, Any, Optional

from core.rag.embeddings import EmbeddingGenerator
from config.settings import settings

logger = logging.getLogger(__name__)


class KnowledgeRetriever:
    """知识库检索器"""
    
    def __init__(self, embedder: Optional[EmbeddingGenerator] = None):
        self.embedder = embedder or EmbeddingGenerator()
    
    async def retrieve(
        self,
        knowledge_base_id: str,
        query: str,
        tenant_id: str,
        top_k: int = None,
        score_threshold: float = 0.5
    ) -> List[Dict[str, Any]]:
        """检索知识库"""
        top_k = top_k or settings.top_k
        
        try:
            # 1. 生成查询向量
            query_embedding = await self.embedder.generate_single(query)
            if not query_embedding:
                logger.error("Failed to generate query embedding")
                return []
            
            # 2. Milvus 向量检索
            vector_results = await self._vector_search(
                knowledge_base_id=knowledge_base_id,
                query_embedding=query_embedding,
                tenant_id=tenant_id,
                top_k=top_k,
                score_threshold=score_threshold
            )
            
            # 3. 合并结果
            return vector_results
        
        except Exception as e:
            logger.error(f"Knowledge retrieval failed: {e}")
            return []
    
    async def _vector_search(
        self,
        knowledge_base_id: str,
        query_embedding: List[float],
        tenant_id: str,
        top_k: int,
        score_threshold: float
    ) -> List[Dict[str, Any]]:
        """Milvus 向量检索"""
        try:
            from pymilvus import Collection, connections
            
            connections.connect(host=settings.milvus_host, port=settings.milvus_port)
            
            collection_name = f"kb_{knowledge_base_id.replace('-', '_')}"
            if not Collection.exists(collection_name):
                connections.disconnect("default")
                return []
            
            collection = Collection(collection_name)
            collection.load()
            
            search_params = {
                "metric_type": "COSINE",
                "params": {"nprobe": 10}
            }
            
            results = collection.search(
                data=[query_embedding],
                anns_field="embedding",
                param=search_params,
                limit=top_k,
                expr=f'tenant_id == "{tenant_id}"',
                output_fields=["content", "metadata", "doc_id"]
            )
            
            chunks = []
            for hits in results:
                for hit in hits:
                    if hit.score >= score_threshold:
                        import json
                        metadata = {}
                        try:
                            metadata = json.loads(hit.entity.get("metadata", "{}"))
                        except:
                            pass
                        
                        chunks.append({
                            "content": hit.entity.get("content", ""),
                            "score": hit.score,
                            "metadata": metadata,
                            "document_id": hit.entity.get("doc_id", ""),
                        })
            
            connections.disconnect("default")
            return chunks
        
        except Exception as e:
            logger.error(f"Vector search failed: {e}")
            return []
    
    async def retrieve_with_context(
        self,
        knowledge_base_id: str,
        query: str,
        tenant_id: str,
        top_k: int = None
    ) -> str:
        """检索并生成上下文文本（用于注入 LLM prompt）"""
        chunks = await self.retrieve(knowledge_base_id, query, tenant_id, top_k)
        
        if not chunks:
            return ""
        
        context_parts = []
        for i, chunk in enumerate(chunks, 1):
            context_parts.append(f"[片段{i}] (相关度: {chunk['score']:.2f})\n{chunk['content']}")
        
        return "\n\n---\n\n".join(context_parts)


# 全局实例
default_retriever = KnowledgeRetriever()
