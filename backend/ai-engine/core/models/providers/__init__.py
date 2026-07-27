"""提供商工厂"""
import logging
from typing import Optional

from .base import BaseProvider, ChatMessage, ChatResponse, EmbeddingResponse
from .qwen import QwenProvider
from .deepseek import DeepSeekProvider
from .openai_provider import OpenAIProvider
from .minimax_provider import MiniMaxProvider
from config.settings import settings

logger = logging.getLogger(__name__)


class ProviderFactory:
    """模型提供商工厂"""
    
    _providers = {
        'qwen': QwenProvider,
        'deepseek': DeepSeekProvider,
        'openai': OpenAIProvider,
        'minimax': MiniMaxProvider,
    }
    
    _instances = {}
    
    @classmethod
    def get_provider(cls, provider_name: str) -> Optional[BaseProvider]:
        """获取提供商实例"""
        if provider_name in cls._instances:
            return cls._instances[provider_name]
        
        provider_class = cls._providers.get(provider_name)
        if not provider_class:
            logger.warning(f"Unknown provider: {provider_name}")
            return None
        
        # 获取 API 密钥
        api_key = getattr(settings, f"{provider_name}_api_key", None)
        if not api_key:
            logger.warning(f"No API key for provider: {provider_name}")
            return None
        
        instance = provider_class(api_key=api_key)
        cls._instances[provider_name] = instance
        
        return instance
    
    @classmethod
    async def close_all(cls):
        """关闭所有提供商"""
        for instance in cls._instances.values():
            await instance.close()
        cls._instances.clear()
