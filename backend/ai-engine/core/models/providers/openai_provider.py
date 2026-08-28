"""OpenAI 模型提供商"""
import time
import logging
from typing import List, AsyncIterator

from openai import AsyncOpenAI

from .base import BaseProvider, ChatMessage, ChatResponse, EmbeddingResponse

logger = logging.getLogger(__name__)


class OpenAIProvider(BaseProvider):
    """OpenAI 提供商"""

    def __init__(self, api_key: str, **kwargs):
        super().__init__(api_key, **kwargs)
        self.client = AsyncOpenAI(
            api_key=api_key,
            base_url=kwargs.get("base_url", "https://api.openai.com/v1")
        )

    async def chat(
        self,
        messages: List[ChatMessage],
        model: str,
        max_tokens: int = 2000,
        temperature: float = 0.7,
        usage_holder: dict = None,
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
        # usage_holder 是内部记账参数，不能传给 OpenAI SDK
        usage_holder = kwargs.pop('usage_holder', None)
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
            stream_options={"include_usage": True},
            **kwargs
        )

        input_tokens = 0

        output_tokens = 0

        async for chunk in stream:

            if getattr(chunk, "usage", None):

                input_tokens = chunk.usage.prompt_tokens or 0

                output_tokens = chunk.usage.completion_tokens or 0

            if chunk.choices and chunk.choices[0].delta.content:

                yield chunk.choices[0].delta.content

        if usage_holder is not None:

            usage_holder["input"] = input_tokens

            usage_holder["output"] = output_tokens

    async def embed(
        self,
        texts: List[str],
        model: str,
        **kwargs
    ) -> EmbeddingResponse:
        start_time = time.time()

        response = await self.client.embeddings.create(
            model=model,
            input=texts,
            **kwargs
        )

        latency_ms = (time.time() - start_time) * 1000

        embeddings = [item.embedding for item in response.data]

        return EmbeddingResponse(
            embeddings=embeddings,
            model=model,
            tokens_used=response.usage.total_tokens if response.usage else 0
        )

    async def close(self):
        await self.client.close()
