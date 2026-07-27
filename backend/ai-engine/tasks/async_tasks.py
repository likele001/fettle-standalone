"""异步任务定义"""
import asyncio
import logging
from typing import List, Dict, Any

from .celery_app import celery_app
from core.rag import DocumentLoader, TextSplitter
from core.rag.embeddings import EmbeddingGenerator
from core.models import model_router
from core.models.providers import ProviderFactory

logger = logging.getLogger(__name__)


async def _generate_and_store(file_path: str, metadata: Dict[str, Any]):
    """异步执行文档向量化并存储到 Milvus"""
    documents = DocumentLoader.load(file_path)
    if not documents:
        raise Exception("Failed to load document")

    splitter = TextSplitter()
    chunks = splitter.split_documents(documents)

    logger.info(f"Split document into {len(chunks)} chunks")

    embedder = EmbeddingGenerator()
    chunk_texts = [chunk.content for chunk in chunks]
    embeddings = await embedder.generate(chunk_texts)

    kb_id = metadata.get("knowledge_base_id", "").replace("-", "_")
    tenant_id = metadata.get("tenant_id", "")
    collection_name = f"kb_{kb_id}"

    from pymilvus import Collection, CollectionSchema, FieldSchema, DataType, connections

    connections.connect(host="localhost", port="19530")

    if not Collection.exists(collection_name):
        fields = [
            FieldSchema(name="id", dtype=DataType.INT64, is_primary=True, auto_id=True),
            FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=1536),
            FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=65535),
            FieldSchema(name="metadata", dtype=DataType.VARCHAR, max_length=65535),
            FieldSchema(name="doc_id", dtype=DataType.VARCHAR, max_length=255),
            FieldSchema(name="tenant_id", dtype=DataType.VARCHAR, max_length=255),
        ]
        schema = CollectionSchema(fields, description=f"Knowledge base {kb_id}")
        collection = Collection(name=collection_name, schema=schema)

        index_params = {
            "metric_type": "COSINE",
            "index_type": "IVF_FLAT",
            "params": {"nlist": 1024}
        }
        collection.create_index(field_name="embedding", index_params=index_params)
    else:
        collection = Collection(collection_name)

    entities = []
    for i, (chunk, embedding) in enumerate(zip(chunks, embeddings)):
        import json
        entities.append({
            "embedding": embedding,
            "content": chunk.content,
            "metadata": json.dumps({"source": file_path, "chunk_index": i}),
            "doc_id": metadata.get("document_id", ""),
            "tenant_id": tenant_id,
        })

    collection.insert(entities)
    collection.flush()

    connections.disconnect("default")

    logger.info(f"Stored {len(entities)} chunks to Milvus collection {collection_name}")
    return len(entities)


@celery_app.task(bind=True, name='tasks.vectorize_document')
def vectorize_document(self, file_path: str, metadata: Dict[str, Any]):
    """文档向量化任务"""
    try:
        logger.info(f"Starting document vectorization: {file_path}")

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            chunk_count = loop.run_until_complete(_generate_and_store(file_path, metadata))
        finally:
            loop.close()

        return {
            "status": "success",
            "chunks": chunk_count,
            "file_path": file_path
        }

    except Exception as e:
        logger.error(f"Document vectorization failed: {e}")
        return {"status": "error", "message": str(e)}


@celery_app.task(bind=True, name='tasks.batch_vectorize')
def batch_vectorize(self, file_paths: List[str], metadata: Dict[str, Any]):
    """批量文档向量化"""
    results = []
    
    for i, file_path in enumerate(file_paths):
        try:
            result = vectorize_document(file_path, metadata)
            results.append(result)
            
            # 更新进度
            self.update_state(
                state='PROGRESS',
                meta={
                    'current': i + 1,
                    'total': len(file_paths),
                    'file': file_path
                }
            )
        except Exception as e:
            logger.error(f"Failed to vectorize {file_path}: {e}")
            results.append({"status": "error", "file": file_path, "message": str(e)})
    
    return {
        "status": "completed",
        "results": results,
        "total": len(file_paths)
    }


@celery_app.task(bind=True, name='tasks.cleanup_expired_memories')
def cleanup_expired_memories(self, max_age_hours: int = 24):
    """清理过期工作记忆"""
    try:
        from core.memory import working_memory_manager
        working_memory_manager.cleanup_expired(max_age_hours)
        return {"status": "success"}
    except Exception as e:
        logger.error(f"Memory cleanup failed: {e}")
        return {"status": "error", "message": str(e)}


@celery_app.task(bind=True, name='tasks.generate_embeddings_batch')
def generate_embeddings_batch(self, texts: List[str], model: str = None):
    """批量生成嵌入向量"""
    try:
        import asyncio
        
        async def _generate():
            emb_config = model_router.get_embedding_model(model)
            provider = ProviderFactory.get_provider(emb_config.provider)
            if not provider:
                raise Exception("Provider not available")
            
            response = await provider.embed(texts, emb_config.model_name)
            return response
        
        # 运行异步任务
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            response = loop.run_until_complete(_generate())
            return {
                "status": "success",
                "embeddings": response.embeddings,
                "model": response.model,
                "tokens_used": response.tokens_used
            }
        finally:
            loop.close()
    
    except Exception as e:
        logger.error(f"Embedding generation failed: {e}")
        return {"status": "error", "message": str(e)}
