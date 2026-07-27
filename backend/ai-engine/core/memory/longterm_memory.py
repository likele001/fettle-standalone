"""长期记忆 - 基于 Milvus 向量数据库"""
import json
import logging
from typing import List, Optional, Dict, Any

from pymilvus import connections, Collection, FieldSchema, CollectionSchema, DataType, utility

from config.settings import settings

logger = logging.getLogger(__name__)


class MemoryItem:
    """记忆项"""

    def __init__(
        self,
        id: str,
        tenant_id: str,
        agent_id: str,
        user_id: str,
        content: str,
        memory_type: str,
        embedding: Optional[List[float]] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ):
        self.id = id
        self.tenant_id = tenant_id
        self.agent_id = agent_id
        self.user_id = user_id
        self.content = content
        self.memory_type = memory_type
        self.embedding = embedding or []
        self.metadata = metadata or {}


class LongTermMemory:
    """长期记忆管理器（Milvus）"""

    COLLECTION_NAME = "long_term_memory"
    EMBEDDING_DIM = 1536

    def __init__(self):
        self.collection: Optional[Collection] = None
        self._connected = False

    async def _ensure_connected(self):
        """确保已连接 Milvus"""
        if self._connected:
            return

        connections.connect(
            "default",
            host=settings.milvus_host,
            port=settings.milvus_port,
        )
        self._connected = True
        logger.info(f"Connected to Milvus: {settings.milvus_host}:{settings.milvus_port}")

        # 创建集合（如果不存在）
        if not utility.has_collection(self.COLLECTION_NAME):
            fields = [
                FieldSchema(name="id", dtype=DataType.VARCHAR, is_primary=True, max_length=64),
                FieldSchema(name="tenant_id", dtype=DataType.VARCHAR, max_length=64),
                FieldSchema(name="agent_id", dtype=DataType.VARCHAR, max_length=64),
                FieldSchema(name="user_id", dtype=DataType.VARCHAR, max_length=64),
                FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=4096),
                FieldSchema(name="memory_type", dtype=DataType.VARCHAR, max_length=32),
                FieldSchema(name="metadata", dtype=DataType.VARCHAR, max_length=8192),
                FieldSchema(
                    name="embedding",
                    dtype=DataType.FLOAT_VECTOR,
                    dim=self.EMBEDDING_DIM,
                ),
            ]
            schema = CollectionSchema(fields, description="长期记忆存储")
            self.collection = Collection(self.COLLECTION_NAME, schema)

            # 创建索引
            index_params = {
                "index_type": "IVF_FLAT",
                "metric_type": "L2",
                "params": {"nlist": 128},
            }
            self.collection.create_index("embedding", index_params)
            logger.info(f"Created collection: {self.COLLECTION_NAME}")
        else:
            self.collection = Collection(self.COLLECTION_NAME)
            self.collection.load()

    async def save_memory(self, item: MemoryItem) -> None:
        """保存长期记忆"""
        await self._ensure_connected()

        data = [
            [item.id],
            [item.tenant_id],
            [item.agent_id],
            [item.user_id],
            [item.content],
            [item.memory_type],
            [json.dumps(item.metadata, ensure_ascii=False)],
            [item.embedding],
        ]
        self.collection.insert(data)
        logger.info(f"Saved long-term memory: {item.id}")

    async def retrieve(
        self,
        tenant_id: str,
        query_embedding: List[float],
        top_k: int = 5,
    ) -> List[MemoryItem]:
        """检索相关记忆"""
        await self._ensure_connected()

        search_params = {"metric_type": "L2", "params": {"nprobe": 16}}

        results = self.collection.search(
            data=[query_embedding],
            anns_field="embedding",
            param=search_params,
            limit=top_k,
            expr=f'tenant_id == "{tenant_id}"',
            output_fields=["id", "tenant_id", "agent_id", "user_id", "content", "memory_type", "metadata"],
        )

        memories = []
        for hits in results:
            for hit in hits:
                item = MemoryItem(
                    id=hit.entity.get("id"),
                    tenant_id=hit.entity.get("tenant_id"),
                    agent_id=hit.entity.get("agent_id"),
                    user_id=hit.entity.get("user_id"),
                    content=hit.entity.get("content"),
                    memory_type=hit.entity.get("memory_type"),
                    embedding=None,
                    metadata=json.loads(hit.entity.get("metadata", "{}")),
                )
                memories.append(item)

        return memories

    async def delete_memory(self, memory_id: str) -> None:
        """删除记忆"""
        await self._ensure_connected()
        self.collection.delete(expr=f'id == "{memory_id}"')
        logger.info(f"Deleted long-term memory: {memory_id}")

    async def summarize_conversation(self, messages: List[Dict[str, str]]) -> str:
        """生成会话摘要（简化版）"""
        if not messages:
            return ""
        # 取最后 3 条消息拼接
        recent = messages[-3:]
        summary = "\n".join(f"{m['role']}: {m['content']}" for m in recent)
        return summary


# 全局实例
longterm_memory = LongTermMemory()
