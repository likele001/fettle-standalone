"""模型提供商基类"""
from abc import ABC, abstractmethod
from typing import List, Dict, Any, Optional, AsyncIterator
from dataclasses import dataclass


@dataclass
class ChatMessage:
    """聊天消息"""
    role: str  # user, assistant, system
    content: str


@dataclass
class ChatResponse:
    """聊天响应"""
    content: str
    model: str
    tokens_used: int
    latency_ms: float


@dataclass
class EmbeddingResponse:
    """嵌入响应"""
    embeddings: List[List[float]]
    model: str
    tokens_used: int


class BaseProvider(ABC):
    """模型提供商基类"""
    
    def __init__(self, api_key: str, **kwargs):
        self.api_key = api_key
        self.config = kwargs
    
    @abstractmethod
    async def chat(
        self,
        messages: List[ChatMessage],
        model: str,
        max_tokens: int = 2000,
        temperature: float = 0.7,
        **kwargs
    ) -> ChatResponse:
        """聊天补全"""
        pass
    
    @abstractmethod
    async def stream_chat(
        self,
        messages: List[ChatMessage],
        model: str,
        max_tokens: int = 2000,
        temperature: float = 0.7,
        **kwargs
    ) -> AsyncIterator[str]:
        """流式聊天"""
        pass
    
    @abstractmethod
    async def embed(
        self,
        texts: List[str],
        model: str,
        **kwargs
    ) -> EmbeddingResponse:
        """生成嵌入"""
        pass
    
    async def close(self):
        """关闭连接"""
        pass
