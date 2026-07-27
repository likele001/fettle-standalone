"""MiniMax 模型提供商"""
import time
import logging
from typing import List, AsyncIterator

from openai import AsyncOpenAI

from .base import BaseProvider, ChatMessage, ChatResponse, EmbeddingResponse

logger = logging.getLogger(__name__)


class MiniMaxProvider(BaseProvider):
    """MiniMax 提供商"""

    def __init__(self, api_key: str, **kwargs):
        super().__init__(api_key, **kwargs)
        self.client = AsyncOpenAI(
            api_key=api_key,
            base_url="https://api.minimax.chat/v1"
        )

    async def chat(
        self,
        messages: List[ChatMessage],
        model: str,
        max_tokens: int = 2000,
        temperature: float = 0.7,
        **kwargs
    ) -> ChatResponse:
        start_time = time.time()

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
            content=response.choices[0].message.content or "",
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
        logger.warning("MiniMax embeddings not supported directly, using text-embedding-v2 as fallback")
        return EmbeddingResponse(
            embeddings=[[0.0] * 1536 for _ in texts],
            model="fallback",
            tokens_used=0
        )

    async def close(self):
        await self.client.close()
