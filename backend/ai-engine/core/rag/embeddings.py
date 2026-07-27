"""嵌入生成模块"""
import logging
from typing import List, Optional

from core.models.model_router import model_router
from core.models.providers import ProviderFactory

logger = logging.getLogger(__name__)


class EmbeddingGenerator:
    """嵌入向量生成器"""
    
    def __init__(self, model_id: Optional[str] = None):
        self.model_id = model_id
        self._provider = None
        self._model_name = None
        self._dimension = None
    
    def _ensure_provider(self):
        """确保 provider 已初始化"""
        if self._provider is not None:
            return
        
        emb_config = model_router.get_embedding_model(self.model_id)
        self._provider = ProviderFactory.get_provider(emb_config.provider)
        self._model_name = emb_config.model_name
        self._dimension = 1536  # 默认维度
        
        if self._provider is None:
            raise RuntimeError("Embedding provider not available")
    
    async def generate(self, texts: List[str], batch_size: int = 10) -> List[List[float]]:
        """生成嵌入向量"""
        self._ensure_provider()
        
        all_embeddings = []
        
        for i in range(0, len(texts), batch_size):
            batch = texts[i:i + batch_size]
            response = await self._provider.embed(batch, self._model_name)
            all_embeddings.extend(response.embeddings)
        
        return all_embeddings
    
    async def generate_single(self, text: str) -> List[float]:
        """生成单条文本的嵌入向量"""
        embeddings = await self.generate([text])
        return embeddings[0] if embeddings else []
    
    @property
    def dimension(self) -> int:
        """嵌入维度"""
        return self._dimension or 1536


# 全局实例
default_embedder = EmbeddingGenerator()
