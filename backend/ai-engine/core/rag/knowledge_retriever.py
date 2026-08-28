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
        score_threshold: float = 0.5,
        rerank: bool = None,
    ) -> List[Dict[str, Any]]:
        """检索知识库（可选重排）"""
        top_k = top_k or settings.top_k

        try:
            # 1. 生成查询向量
            query_embedding = await self.embedder.generate_single(query)
            if not query_embedding:
                logger.error("Failed to generate query embedding")
                return []

            # 2. 决定是否重排，并据此拉取更多候选、降低阈值
            do_rerank = settings.rerank_enabled if rerank is None else rerank
            fetch_k = top_k
            fetch_threshold = score_threshold
            if do_rerank:
                fetch_k = max(top_k, settings.rerank_candidate_k)
                fetch_threshold = min(score_threshold, settings.rerank_min_score)

            vector_results = await self._vector_search(
                knowledge_base_id=knowledge_base_id,
                query_embedding=query_embedding,
                tenant_id=tenant_id,
                top_k=fetch_k,
                score_threshold=fetch_threshold,
            )

            if not vector_results:
                return []

            # 3. 重排（失败自动保留向量序）
            if do_rerank:
                try:
                    from core.rag.reranker import Reranker
                    reranker = Reranker(mode=settings.rerank_mode, neural_model=settings.rerank_model)
                    vector_results = reranker.rerank(query, vector_results, top_n=top_k)
                    logger.info(f"Reranked to {len(vector_results)} chunks (mode={settings.rerank_mode})")
                except Exception as e:
                    logger.error(f"Rerank failed, keep vector order: {e}")

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
            from pymilvus import Collection, connections, utility

            connections.connect(host=settings.milvus_host, port=settings.milvus_port)

            collection_name = f"kb_{knowledge_base_id.replace('-', '_')}"
            if not utility.has_collection(collection_name):
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
            score = chunk.get("rerank_score", chunk.get("score", 0))
            context_parts.append(f"[片段{i}] (相关度: {score:.2f})\n{chunk['content']}")

        return "\n\n---\n\n".join(context_parts)


# 全局实例
default_retriever = KnowledgeRetriever()
