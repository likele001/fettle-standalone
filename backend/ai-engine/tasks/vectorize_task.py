"""文档向量化任务"""
import logging
from typing import List, Dict, Any, Optional

from tasks.celery_app import celery_app
from core.rag import DocumentLoader, TextSplitter, TextChunk
from core.models.model_router import model_router
from core.models.providers import ProviderFactory

logger = logging.getLogger(__name__)


@celery_app.task(bind=True, name='tasks.vectorize_document', max_retries=3)
def vectorize_document(self, doc_id: str, file_path: str, knowledge_base_id: str, tenant_id: str):
    """文档向量化任务"""
    try:
        logger.info(f"Starting vectorization: doc={doc_id}, file={file_path}")
        
        # 1. 加载文档
        documents = DocumentLoader.load(file_path)
        if not documents:
            logger.error(f"Failed to load document: {file_path}")
            return {"status": "error", "message": "Failed to load document"}
        
        # 2. 分割文本
        splitter = TextSplitter(chunk_size=500, chunk_overlap=50)
        chunks = splitter.split_documents(documents)
        
        logger.info(f"Split into {len(chunks)} chunks")
        
        # 3. 生成嵌入向量
        texts = [chunk.content for chunk in chunks]
        embeddings = generate_embeddings_batch(texts)
        
        if not embeddings:
            logger.error("Failed to generate embeddings")
            return {"status": "error", "message": "Failed to generate embeddings"}
        
        # 4. 存储到 Milvus
        stored_count = store_to_milvus(
            chunks=chunks,
            embeddings=embeddings,
            knowledge_base_id=knowledge_base_id,
            tenant_id=tenant_id,
            doc_id=doc_id
        )
        
        logger.info(f"Stored {stored_count} chunks to Milvus")

        # 5. 更新文档状态（向量化成功）
        try:
            import asyncio
            import asyncpg
            from config.settings import settings as _s

            async def _update_doc_status():
                conn = await asyncpg.connect(
                    host=_s.db_host, port=_s.db_port,
                    user=_s.db_user, password=_s.db_password,
                    database=_s.db_name,
                )
                await conn.execute(
                    "UPDATE knowledge_documents SET status='processed', chunk_count=$1 WHERE id=$2",
                    len(chunks), doc_id,
                )
                await conn.close()

            asyncio.run(_update_doc_status())
            logger.info(f"Updated doc status to processed: {doc_id} chunks={len(chunks)}")
        except Exception as _e:
            logger.error(f"Failed to update doc status: {_e}")

        return {
            "status": "success",
            "doc_id": doc_id,
            "chunks": len(chunks),
            "stored": stored_count
        }
    
    except Exception as e:
        logger.error(f"Vectorization failed: {e}")
        # 重试
        raise self.retry(exc=e, countdown=60)


def generate_embeddings_batch(texts: List[str]) -> List[List[float]]:
    """批量生成嵌入向量"""
    import asyncio
    
    async def _generate():
        emb_config = model_router.get_embedding_model()
        provider = ProviderFactory.get_provider(emb_config.provider)
        if not provider:
            raise Exception("Embedding provider not available")
        
        # 分批处理，每批最多 10 条
        all_embeddings = []
        batch_size = 10
        
        for i in range(0, len(texts), batch_size):
            batch = texts[i:i + batch_size]
            response = await provider.embed(batch, emb_config.model_name)
            all_embeddings.extend(response.embeddings)
        
        return all_embeddings
    
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    try:
        return loop.run_until_complete(_generate())
    finally:
        loop.close()


def store_to_milvus(
    chunks: List[TextChunk],
    embeddings: List[List[float]],
    knowledge_base_id: str,
    tenant_id: str,
    doc_id: str
) -> int:
    """存储到 Milvus"""
    try:
        from pymilvus import Collection, connections, utility
        
        # 连接 Milvus
        connections.connect(host="localhost", port="19530")
        
        # 获取或创建 collection
        collection_name = f"kb_{knowledge_base_id.replace('-', '_')}"
        
        if not utility.has_collection(collection_name):
            from pymilvus import FieldSchema, CollectionSchema, DataType
            
            fields = [
                FieldSchema(name="id", dtype=DataType.VARCHAR, is_primary=True, max_length=64),
                FieldSchema(name="tenant_id", dtype=DataType.VARCHAR, max_length=64),
                FieldSchema(name="doc_id", dtype=DataType.VARCHAR, max_length=64),
                FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=4000),
                FieldSchema(name="metadata", dtype=DataType.JSON),
                FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=1536),
            ]
            schema = CollectionSchema(fields=fields)
            collection = Collection(name=collection_name, schema=schema)
            
            # 创建索引
            index_params = {
                "metric_type": "COSINE",
                "index_type": "IVF_FLAT",
                "params": {"nlist": 128}
            }
            collection.create_index(field_name="embedding", index_params=index_params)
        else:
            collection = Collection(collection_name)
        
        # 准备数据
        import json
        ids = [f"{doc_id}_{i}" for i in range(len(chunks))]
        tenant_ids = [tenant_id] * len(chunks)
        doc_ids = [doc_id] * len(chunks)
        contents = [chunk.content[:4000] for chunk in chunks]  # 截断
        metadatas = [json.dumps(chunk.metadata) for chunk in chunks]
        
        # 插入数据
        data = [ids, tenant_ids, doc_ids, contents, metadatas, embeddings]
        collection.insert(data)
        collection.flush()
        
        connections.disconnect("default")
        return len(chunks)
    
    except Exception as e:
        logger.error(f"Failed to store to Milvus: {e}")
        return 0


@celery_app.task(bind=True, name='tasks.delete_document_vectors')
def delete_document_vectors(self, knowledge_base_id: str, doc_id: str):
    """删除文档向量"""
    try:
        from pymilvus import Collection, connections, utility
        
        connections.connect(host="localhost", port="19530")
        
        collection_name = f"kb_{knowledge_base_id.replace('-', '_')}"
        if utility.has_collection(collection_name):
            collection = Collection(collection_name)
            expr = f'doc_id == "{doc_id}"'
            collection.delete(expr)
            collection.flush()
        
        connections.disconnect("default")
        return {"status": "success", "doc_id": doc_id}
    
    except Exception as e:
        logger.error(f"Failed to delete vectors: {e}")
        return {"status": "error", "message": str(e)}
