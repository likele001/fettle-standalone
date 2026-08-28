"""数据库驱动的 Provider 工厂 - 使用租户 API Key"""
import logging
from typing import Optional

from .base import BaseProvider
from .qwen import QwenProvider
from .deepseek import DeepSeekProvider
from .openai_provider import OpenAIProvider
from .minimax_provider import MiniMaxProvider
from .langchain_provider import LangChainProvider
from config.settings import settings
from ..db_config import TenantAPIKey

logger = logging.getLogger(__name__)


class DBProviderFactory:
    """数据库驱动的 Provider 工厂"""
    
    PROVIDER_MAP = {
        'aliyun': QwenProvider,
        'qwen': QwenProvider,
        'baidu': None,
        'tencent': None,
        'openai': OpenAIProvider,
        'anthropic': None,
        'google': None,
        'deepseek': DeepSeekProvider,
        'minimax': MiniMaxProvider,
        'moonshot': OpenAIProvider,
        'zhipu': OpenAIProvider,
    }

    LANGCHAIN_PROVIDER_MAP = {
        'openai': 'openai',
        'anthropic': 'anthropic',
        'deepseek': 'deepseek',
        'google': 'google',
        'ollama': 'ollama',
        'mistralai': 'mistralai',
        'huggingface': 'huggingface',
    }
    
    @classmethod
    async def get_provider(
        cls,
        provider_code: str,
        tenant_api_key: Optional[TenantAPIKey] = None,
        api_key: Optional[str] = None,
        use_langchain: bool = False,
    ) -> Optional[BaseProvider]:
        """
        获取 Provider 实例
        
        优先使用租户 API Key，其次使用传入的 api_key
        use_langchain: 是否使用 LangChain 封装
        """
        if use_langchain:
            langchain_type = cls.LANGCHAIN_PROVIDER_MAP.get(provider_code)
            if langchain_type:
                return await cls._get_langchain_provider(langchain_type, tenant_api_key, api_key)
            logger.warning(f"No LangChain provider found for: {provider_code}")

        provider_class = cls.PROVIDER_MAP.get(provider_code)
        if not provider_class:
            logger.warning(f"No provider class found for: {provider_code}")
            return None
        
        final_api_key = None
        base_url = None
        custom_headers = {}
        
        if tenant_api_key:
            final_api_key = tenant_api_key.api_key_value
            if tenant_api_key.custom_base_url:
                base_url = tenant_api_key.custom_base_url
            if tenant_api_key.custom_headers:
                custom_headers = tenant_api_key.custom_headers
        elif api_key:
            final_api_key = api_key

        # 全局 fallback：未配置租户 base_url 时，用 .env 的 {provider}_base_url（如 OPENAI_BASE_URL 指向本地网关）
        if not base_url:
            base_url = getattr(settings, f"{provider_code}_base_url", None)
            if base_url:
                logger.info(f"Using global .env base_url for provider={provider_code}: {base_url}")
        
        if not final_api_key:
            logger.warning(f"No API key available for provider: {provider_code}")
            return None
        
        try:
            kwargs = {}
            if base_url:
                kwargs['base_url'] = base_url
            if custom_headers:
                kwargs['custom_headers'] = custom_headers
            
            provider = provider_class(final_api_key, **kwargs)
            logger.info(f"Created provider: {provider_code}")
            return provider
        except Exception as e:
            logger.error(f"Failed to create provider {provider_code}: {e}")
            return None

    @classmethod
    async def _get_langchain_provider(
        cls,
        langchain_type: str,
        tenant_api_key: Optional[TenantAPIKey] = None,
        api_key: Optional[str] = None,
    ) -> Optional[LangChainProvider]:
        """获取 LangChain 封装的 Provider"""
        final_api_key = None
        base_url = None
        
        if tenant_api_key:
            final_api_key = tenant_api_key.api_key_value
            if tenant_api_key.custom_base_url:
                base_url = tenant_api_key.custom_base_url
        elif api_key:
            final_api_key = api_key

        # 全局 fallback：未配置租户 base_url 时，用 .env 的 {provider}_base_url（如 OPENAI_BASE_URL 指向本地网关）
        if not base_url:
            base_url = getattr(settings, f"{provider_code}_base_url", None)
            if base_url:
                logger.info(f"Using global .env base_url for provider={provider_code}: {base_url}")
        
        if not final_api_key:
            logger.warning(f"No API key available for LangChain provider: {langchain_type}")
            return None
        
        try:
            kwargs = {}
            if base_url:
                kwargs['base_url'] = base_url
            
            provider = LangChainProvider(langchain_type, final_api_key, **kwargs)
            logger.info(f"Created LangChain provider: {langchain_type}")
            return provider
        except Exception as e:
            logger.error(f"Failed to create LangChain provider {langchain_type}: {e}")
            return None
    
    @classmethod
    async def get_provider_by_tenant(
        cls,
        tenant_id: str,
        provider_id: str,
        db_config,
        use_langchain: bool = False,
    ) -> Optional[BaseProvider]:
        """通过租户 ID 和厂商 ID 获取 Provider"""
        tenant_api_key = await db_config.get_tenant_api_key(tenant_id, provider_id)
        if tenant_api_key:
            return await cls.get_provider(tenant_api_key.provider_code, tenant_api_key, use_langchain=use_langchain)
        
        logger.warning(f"No API key found for tenant {tenant_id} and provider {provider_id}")
        return None