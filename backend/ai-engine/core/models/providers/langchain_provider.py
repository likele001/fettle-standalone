"""LangChain 模型提供商封装"""
import time
import logging
from typing import List, AsyncIterator, Optional, Any

from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_anthropic import ChatAnthropic

from .base import BaseProvider, ChatMessage, ChatResponse, EmbeddingResponse

logger = logging.getLogger(__name__)

# 可选提供商（可能未安装）
try:
    from langchain_deepseek import ChatDeepSeek
    DEEPSEEK_AVAILABLE = True
except ImportError:
    ChatDeepSeek = None
    DEEPSEEK_AVAILABLE = False
    logger.info("langchain_deepseek not available")

try:
    from langchain_google_genai import ChatGoogleGenerativeAI, GoogleGenerativeAIEmbeddings
    GOOGLE_AVAILABLE = True
except ImportError:
    ChatGoogleGenerativeAI = None
    GoogleGenerativeAIEmbeddings = None
    GOOGLE_AVAILABLE = False

try:
    from langchain_ollama import ChatOllama, OllamaEmbeddings
    OLLAMA_AVAILABLE = True
except ImportError:
    ChatOllama = None
    OllamaEmbeddings = None
    OLLAMA_AVAILABLE = False

try:
    from langchain_mistralai import ChatMistralAI
    MISTRAL_AVAILABLE = True
except ImportError:
    ChatMistralAI = None
    MISTRAL_AVAILABLE = False

try:
    from langchain_huggingface import HuggingFaceEmbeddings
    HUGGINGFACE_AVAILABLE = True
except ImportError:
    HuggingFaceEmbeddings = None
    HUGGINGFACE_AVAILABLE = False


class LangChainProvider(BaseProvider):

    PROVIDER_MAP = {
        "openai": ChatOpenAI,
        "anthropic": ChatAnthropic,
        "deepseek": ChatDeepSeek if DEEPSEEK_AVAILABLE else None,
        "google": ChatGoogleGenerativeAI if GOOGLE_AVAILABLE else None,
        "ollama": ChatOllama if OLLAMA_AVAILABLE else None,
        "mistralai": ChatMistralAI if MISTRAL_AVAILABLE else None,
    }

    EMBEDDING_MAP = {
        "openai": OpenAIEmbeddings,
        "deepseek": None,
        "google": GoogleGenerativeAIEmbeddings if GOOGLE_AVAILABLE else None,
        "ollama": OllamaEmbeddings if OLLAMA_AVAILABLE else None,
        "huggingface": HuggingFaceEmbeddings if HUGGINGFACE_AVAILABLE else None,
    }

    def __init__(self, provider_type: str, api_key: str, **kwargs):
        super().__init__(api_key, **kwargs)
        self.provider_type = provider_type
        self._init_llm(provider_type, api_key, kwargs)
        self._init_embeddings(provider_type, api_key, kwargs)

    def _init_llm(self, provider_type: str, api_key: str, config: dict):
        llm_class = self.PROVIDER_MAP.get(provider_type)
        if not llm_class:
            logger.warning(f"No LangChain LLM class found for provider: {provider_type}")
            self.llm = None
            return

        llm_kwargs = {}
        if provider_type == "openai":
            llm_kwargs["api_key"] = api_key
            llm_kwargs["base_url"] = config.get("base_url")
        elif provider_type == "anthropic":
            llm_kwargs["api_key"] = api_key
        elif provider_type == "deepseek":
            llm_kwargs["api_key"] = api_key
        elif provider_type == "google":
            llm_kwargs["google_api_key"] = api_key
        elif provider_type == "ollama":
            llm_kwargs["base_url"] = config.get("base_url", "http://localhost:11434")
        elif provider_type == "mistralai":
            llm_kwargs["api_key"] = api_key

        try:
            self.llm = llm_class(**llm_kwargs)
        except Exception as e:
            logger.error(f"Failed to initialize {provider_type} LLM: {e}")
            self.llm = None

    def _init_embeddings(self, provider_type: str, api_key: str, config: dict):
        embedding_class = self.EMBEDDING_MAP.get(provider_type)
        if not embedding_class:
            logger.warning(f"No LangChain embedding class found for provider: {provider_type}")
            self.embeddings = None
            return

        embed_kwargs = {}
        if provider_type == "openai":
            embed_kwargs["api_key"] = api_key
            embed_kwargs["base_url"] = config.get("base_url")
        elif provider_type == "deepseek":
            embed_kwargs["api_key"] = api_key
        elif provider_type == "google":
            embed_kwargs["google_api_key"] = api_key
        elif provider_type == "ollama":
            embed_kwargs["base_url"] = config.get("base_url", "http://localhost:11434")
        elif provider_type == "huggingface":
            embed_kwargs["model_name"] = config.get("model_name", "all-MiniLM-L6-v2")

        try:
            self.embeddings = embedding_class(**embed_kwargs)
        except Exception as e:
            logger.error(f"Failed to initialize {provider_type} embeddings: {e}")
            self.embeddings = None

    def get_llm(self, model_name: str = None) -> Any:
        if self.llm is None:
            return None
        if model_name:
            self.llm.model_name = model_name
        return self.llm

    def get_embeddings(self) -> Any:
        return self.embeddings

    async def chat(
        self,
        messages: List[ChatMessage],
        model: str,
        max_tokens: int = 2000,
        temperature: float = 0.7,
        **kwargs
    ) -> ChatResponse:
        if self.llm is None:
            return ChatResponse(content="", model=model, tokens_used=0, latency_ms=0)

        start_time = time.time()

        langchain_messages = [
            {"role": msg.role, "content": msg.content}
            for msg in messages
        ]

        try:
            llm_with_model = self.llm.bind(model=model, max_tokens=max_tokens, temperature=temperature)
            response = await llm_with_model.ainvoke(langchain_messages)

            latency_ms = (time.time() - start_time) * 1000

            return ChatResponse(
                content=response.content if hasattr(response, "content") else str(response),
                model=model,
                tokens_used=0,
                latency_ms=latency_ms
            )
        except Exception as e:
            logger.error(f"LangChain chat error: {e}")
            return ChatResponse(content="", model=model, tokens_used=0, latency_ms=0)

    async def stream_chat(
        self,
        messages: List[ChatMessage],
        model: str,
        max_tokens: int = 2000,
        temperature: float = 0.7,
        **kwargs
    ) -> AsyncIterator[str]:
        if self.llm is None:
            return

        langchain_messages = [
            {"role": msg.role, "content": msg.content}
            for msg in messages
        ]

        try:
            llm_with_model = self.llm.bind(model=model, max_tokens=max_tokens, temperature=temperature)
            async for chunk in llm_with_model.astream(langchain_messages):
                if hasattr(chunk, "content") and chunk.content:
                    yield chunk.content
        except Exception as e:
            logger.error(f"LangChain stream error: {e}")

    async def embed(
        self,
        texts: List[str],
        model: str,
        **kwargs
    ) -> EmbeddingResponse:
        if self.embeddings is None:
            return EmbeddingResponse(
                embeddings=[[0.0] * 1536 for _ in texts],
                model=model,
                tokens_used=0
            )

        try:
            embeddings = await self.embeddings.aembed_documents(texts)
            return EmbeddingResponse(
                embeddings=embeddings,
                model=model,
                tokens_used=0
            )
        except Exception as e:
            logger.error(f"LangChain embed error: {e}")
            return EmbeddingResponse(
                embeddings=[[0.0] * 1536 for _ in texts],
                model=model,
                tokens_used=0
            )

    async def close(self):
        pass