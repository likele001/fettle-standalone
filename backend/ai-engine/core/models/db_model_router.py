"""数据库驱动的模型路由器 - 根据租户配置和任务类型选择模型"""
import logging
from typing import Optional, Dict, Any, List
from dataclasses import dataclass

from .db_config import ai_config_db, DBModel, TenantAPIKey

logger = logging.getLogger(__name__)


@dataclass
class ModelConfig:
    """模型配置"""
    model_id: str
    provider_code: str
    provider_id: str
    model_name: str
    model_code: str
    max_input_tokens: int
    max_output_tokens: int
    input_price_per_1k: float
    output_price_per_1k: float
    capabilities: List[str]
    is_default: bool


class DBModelRouter:
    """数据库驱动的模型路由器"""
    
    def __init__(self):
        self._cache = {
            'models': {},
            'providers': {},
            'tenant_configs': {},
        }
    
    async def _load_models(self) -> Dict[str, ModelConfig]:
        """从数据库加载所有模型"""
        models = await ai_config_db.list_models()
        result = {}
        for m in models:
            result[m.id] = ModelConfig(
                model_id=m.id,
                provider_code=m.provider_code,
                provider_id=m.provider_id,
                model_name=m.model_name,
                model_code=m.model_code,
                max_input_tokens=m.max_input_tokens,
                max_output_tokens=m.max_output_tokens,
                input_price_per_1k=m.input_price_per_1k,
                output_price_per_1k=m.output_price_per_1k,
                capabilities=m.capabilities,
                is_default=m.is_default,
            )
        self._cache['models'] = result
        logger.info(f"Loaded {len(result)} models from database")
        return result
    
    async def select_model(
        self,
        tenant_id: str,
        model_id: Optional[str] = None,
        provider_id: Optional[str] = None,
    ) -> ModelConfig:
        """
        选择模型
        
        优先级：
        1. 指定模型 ID
        2. 租户默认模型
        3. 租户默认厂商的默认模型
        4. 平台默认模型（第一个可用模型）
        """
        # 确保缓存已加载
        if not self._cache['models']:
            await self._load_models()
        
        models = self._cache['models']
        
        # 1. 指定模型 ID
        if model_id and model_id in models:
            return models[model_id]
        
        # 2. 获取租户配置
        tenant_config = await ai_config_db.get_tenant_ai_config(tenant_id)
        
        # 3. 租户默认模型
        if tenant_config and tenant_config.default_chat_model_id:
            if tenant_config.default_chat_model_id in models:
                return models[tenant_config.default_chat_model_id]
        
        # 4. 租户默认厂商的默认模型
        if tenant_config and tenant_config.default_provider_id:
            for _, model in models.items():
                if model.provider_id == tenant_config.default_provider_id and model.is_default:
                    return model
        
        # 5. 平台默认（第一个可用模型）
        if models:
            return next(iter(models.values()))
        
        raise ValueError("No available models in database")
    
    async def select_embedding_model(
        self,
        tenant_id: str,
        model_id: Optional[str] = None,
    ) -> ModelConfig:
        """选择嵌入模型"""
        if not self._cache['models']:
            await self._load_models()
        
        models = self._cache['models']
        
        # 指定模型 ID
        if model_id and model_id in models:
            model = models[model_id]
            if model.model_type == 'embedding':
                return model
        
        # 获取租户配置
        tenant_config = await ai_config_db.get_tenant_ai_config(tenant_id)
        
        # 租户默认嵌入模型
        if tenant_config and tenant_config.default_embedding_model_id:
            if tenant_config.default_embedding_model_id in models:
                model = models[tenant_config.default_embedding_model_id]
                if model.model_type == 'embedding':
                    return model
        
        # 返回第一个可用嵌入模型
        for _, model in models.items():
            if model.model_type == 'embedding':
                return model
        
        raise ValueError("No available embedding models in database")
    
    async def get_tenant_api_key(self, tenant_id: str, provider_id: str) -> Optional[TenantAPIKey]:
        """获取租户 API Key"""
        return await ai_config_db.get_tenant_api_key(tenant_id, provider_id)
    
    async def get_tenant_config(self, tenant_id: str):
        """获取租户配置"""
        return await ai_config_db.get_tenant_ai_config(tenant_id)
    
    async def list_models(self) -> List[Dict[str, Any]]:
        """列出所有模型"""
        if not self._cache['models']:
            await self._load_models()
        
        return [
            {
                'id': cfg.model_id,
                'provider_code': cfg.provider_code,
                'provider_id': cfg.provider_id,
                'model_code': cfg.model_code,
                'model_name': cfg.model_name,
                'model_type': 'chat' if 'chat' in cfg.capabilities else cfg.capabilities[0] if cfg.capabilities else 'chat',
                'capabilities': cfg.capabilities,
                'is_default': cfg.is_default,
                'max_input_tokens': cfg.max_input_tokens,
                'max_output_tokens': cfg.max_output_tokens,
            }
            for cfg in self._cache['models'].values()
        ]
    
    async def list_providers(self) -> List[Dict[str, Any]]:
        """列出所有厂商"""
        providers = await ai_config_db.list_providers()
        return [
            {
                'id': p.id,
                'code': p.code,
                'name': p.name,
                'api_base_url': p.api_base_url,
                'support_streaming': p.support_streaming,
            }
            for p in providers
        ]
    
    def clear_cache(self):
        """清除缓存"""
        self._cache = {
            'models': {},
            'providers': {},
            'tenant_configs': {},
        }


# 全局数据库模型路由器实例
db_model_router = DBModelRouter()