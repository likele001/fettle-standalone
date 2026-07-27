"""通义千问模型提供商"""
import time
import logging
from typing import List, AsyncIterator

from openai import AsyncOpenAI

from .base import BaseProvider, ChatMessage, ChatResponse, EmbeddingResponse

logger = logging.getLogger(__name__)


class QwenProvider(BaseProvider):
    """通义千问提供商"""
    
    def __init__(self, api_key: str, **kwargs):
        super().__init__(api_key, **kwargs)
        self.client = AsyncOpenAI(
            api_key=api_key,
            base_url="https://dashscope.aliyuncs.com/compatible-mode/v1"
        )
    
    async def chat(
        self,
        messages: List[ChatMessage],
        model: str,
        max_tokens: int = 2000,
        temperature: float = 0.7,
        **kwargs
    ) -> ChatResponse:
        """聊天补全"""
        start_time = time.time()
        
        # 转换为 OpenAI 格式
        openai_messages = [
            {"role": msg.role, "content": msg.content}
            for msg in messages
        ]
        
        response = await self.client.chat.completions.create(
            model=model,
            messages=openai_messages,
            max_tokens=max_tokens,
            temperature=temperature,
            **kwargs
        )
        
        latency_ms = (time.time() - start_time) * 1000
        
        return ChatResponse(
            content=response.choices[0].message.content,
            model=model,
            tokens_used=response.usage.total_tokens if response.usage else 0,
            latency_ms=latency_ms
        )
    
    async def stream_chat(
        self,
        messages: List[ChatMessage],
        model: str,
        max_tokens: int = 2000,
        temperature: float = 0.7,
        **kwargs
    ) -> AsyncIterator[str]:
        """流式聊天"""
        openai_messages = [
            {"role": msg.role, "content": msg.content}
            for msg in messages
        ]
        
        stream = await self.client.chat.completions.create(
            model=model,
            messages=openai_messages,
            max_tokens=max_tokens,
            temperature=temperature,
            stream=True,
            **kwargs
        )
        
        async for chunk in stream:
            if chunk.choices and chunk.choices[0].delta.content:
                yield chunk.choices[0].delta.content
    
    async def embed(
        self,
        texts: List[str],
        model: str,
        **kwargs
    ) -> EmbeddingResponse:
        """生成嵌入"""
        response = await self.client.embeddings.create(
            model=model,
            input=texts,
            **kwargs
        )
        
        embeddings = [item.embedding for item in response.data]
        tokens_used = response.usage.total_tokens if response.usage else 0
        
        return EmbeddingResponse(
            embeddings=embeddings,
            model=model,
            tokens_used=tokens_used
        )
    
    async def close(self):
        """关闭连接"""
        await self.client.close()
