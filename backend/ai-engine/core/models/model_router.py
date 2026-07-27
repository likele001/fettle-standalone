"""模型路由器 - 根据任务类型和租户套餐选择最优模型"""
import logging
from typing import Optional, Dict, Any, List
from dataclasses import dataclass
from enum import Enum

import yaml
from config.settings import settings

logger = logging.getLogger(__name__)


class TaskType(str, Enum):
    """任务类型"""
    SIMPLE_CHAT = "simple_chat"
    COMPLEX_REASONING = "complex_reasoning"
    CODE_GENERATION = "code_generation"
    KNOWLEDGE_QA = "knowledge_qa"
    INTENT_RECOGNITION = "intent_recognition"
    TASK_PLANNING = "task_planning"


@dataclass
class ModelConfig:
    """模型配置"""
    provider: str
    model_name: str
    max_tokens: int
    temperature: float
    cost_per_1k_tokens: float
    capabilities: List[str]


class ModelRouter:
    """模型路由器"""
    
    def __init__(self, config_path: str = "config/models.yaml"):
        self.config_path = config_path
        self.models: Dict[str, ModelConfig] = {}
        self.embeddings: Dict[str, ModelConfig] = {}
        self.routing_config: Dict[str, Any] = {}
        self._load_config()
    
    def _load_config(self):
        """加载模型配置"""
        try:
            with open(self.config_path, 'r', encoding='utf-8') as f:
                config = yaml.safe_load(f)
            
            # 加载模型配置
            for model_id, model_cfg in config.get('models', {}).items():
                self.models[model_id] = ModelConfig(
                    provider=model_cfg['provider'],
                    model_name=model_cfg['model_name'],
                    max_tokens=model_cfg.get('max_tokens', 2000),
                    temperature=model_cfg.get('temperature', 0.7),
                    cost_per_1k_tokens=model_cfg.get('cost_per_1k_tokens', 0.01),
                    capabilities=model_cfg.get('capabilities', ['chat'])
                )
            
            # 加载嵌入模型配置
            for emb_id, emb_cfg in config.get('embeddings', {}).items():
                self.embeddings[emb_id] = ModelConfig(
                    provider=emb_cfg['provider'],
                    model_name=emb_cfg['model_name'],
                    max_tokens=0,
                    temperature=0,
                    cost_per_1k_tokens=emb_cfg.get('cost_per_1k_tokens', 0.001),
                    capabilities=['embedding']
                )
            
            # 加载路由配置
            self.routing_config = config.get('routing', {})
            
            logger.info(f"Loaded {len(self.models)} models, {len(self.embeddings)} embeddings")
        except Exception as e:
            logger.error(f"Failed to load model config: {e}")
            raise
    
    def select_model(
        self,
        task_type: Optional[TaskType] = None,
        tenant_plan: Optional[str] = None,
        required_capabilities: Optional[List[str]] = None,
        model_id: Optional[str] = None
    ) -> ModelConfig:
        """
        选择最优模型
        
        优先级：
        1. 指定模型ID
        2. 租户套餐路由
        3. 任务类型路由
        4. 默认模型
        """
        # 1. 指定模型
        if model_id and model_id in self.models:
            return self.models[model_id]
        
        # 2. 租户套餐路由
        if tenant_plan:
            tenant_routing = self.routing_config.get('tenant_routing', {})
            model_id = tenant_routing.get(tenant_plan)
            if model_id and model_id in self.models:
                model = self.models[model_id]
                # 检查能力要求
                if required_capabilities:
                    if all(cap in model.capabilities for cap in required_capabilities):
                        return model
                else:
                    return model
        
        # 3. 任务类型路由
        if task_type:
            task_routing = self.routing_config.get('task_routing', {})
            model_id = task_routing.get(task_type.value)
            if model_id and model_id in self.models:
                return self.models[model_id]
        
        # 4. 默认模型
        default_model = settings.default_model
        if default_model in self.models:
            return self.models[default_model]
        
        # 5. 返回第一个可用模型
        if self.models:
            return next(iter(self.models.values()))
        
        raise ValueError("No available models")
    
    def get_embedding_model(self, model_id: Optional[str] = None) -> ModelConfig:
        """获取嵌入模型"""
        if model_id and model_id in self.embeddings:
            return self.embeddings[model_id]
        
        # 使用配置的默认嵌入模型
        default_emb = settings.embedding_model
        if default_emb in self.embeddings:
            return self.embeddings[default_emb]
        
        # 返回第一个可用嵌入模型
        if self.embeddings:
            return next(iter(self.embeddings.values()))
        
        raise ValueError("No available embedding models")
    
    def list_models(self) -> List[Dict[str, Any]]:
        """列出所有可用模型"""
        return [
            {
                'id': model_id,
                'provider': cfg.provider,
                'model_name': cfg.model_name,
                'capabilities': cfg.capabilities,
                'cost_per_1k_tokens': cfg.cost_per_1k_tokens
            }
            for model_id, cfg in self.models.items()
        ]
    
    def list_embedding_models(self) -> List[Dict[str, Any]]:
        """列出所有嵌入模型"""
        return [
            {
                'id': emb_id,
                'provider': cfg.provider,
                'model_name': cfg.model_name,
                'dimension': settings.embedding_dimension,
                'cost_per_1k_tokens': cfg.cost_per_1k_tokens
            }
            for emb_id, cfg in self.embeddings.items()
        ]


# 全局模型路由器实例
model_router = ModelRouter()
