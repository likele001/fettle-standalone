"""数据库访问模块 - 从 PostgreSQL 读取 AI 配置"""
import asyncio
import logging
from typing import Optional, Dict, List, Any
from dataclasses import dataclass

import asyncpg

from config.settings import settings
from core.utils.crypto import aes_cipher

logger = logging.getLogger(__name__)


@dataclass
class DBProvider:
    """AI 厂商配置（从数据库读取）"""
    id: str
    code: str
    name: str
    api_base_url: str
    auth_type: str
    auth_config: Dict[str, Any]
    support_streaming: bool


@dataclass
class DBModel:
    """AI 模型配置（从数据库读取）"""
    id: str
    provider_id: str
    provider_code: str
    model_code: str
    model_name: str
    model_type: str
    max_input_tokens: int
    max_output_tokens: int
    input_price_per_1k: float
    output_price_per_1k: float
    capabilities: List[str]
    is_default: bool


@dataclass
class TenantAPIKey:
    """租户 API Key（从数据库读取）"""
    id: str
    tenant_id: str
    provider_id: str
    provider_code: str
    api_key_value: str
    custom_base_url: str
    custom_headers: Dict[str, Any]
    status: str


@dataclass
class TenantAIConfig:
    """租户 AI 配置（从数据库读取）"""
    tenant_id: str
    default_chat_model_id: Optional[str]
    default_embedding_model_id: Optional[str]
    default_provider_id: Optional[str]
    ai_enabled: bool
    streaming_enabled: bool


class AIConfigDB:
    """AI 配置数据库访问"""
    
    def __init__(self):
        self.pool = None
    
    async def init_pool(self):
        """初始化数据库连接池"""
        if self.pool:
            return
        
        try:
            self.pool = await asyncpg.create_pool(
                host=settings.db_host,
                port=settings.db_port,
                user=settings.db_user,
                password=settings.db_password,
                database=settings.db_name,
                min_size=1,
                max_size=10,
                command_timeout=30,
            )
            logger.info("AI config database pool initialized")
        except Exception as e:
            logger.error(f"Failed to initialize database pool: {e}")
            raise
    
    async def get_provider(self, provider_id: str = None, provider_code: str = None) -> Optional[DBProvider]:
        """获取厂商配置"""
        await self.init_pool()
        
        query = """
            SELECT id, code, name, api_base_url, auth_type, auth_config, support_streaming
            FROM ai_providers
            WHERE status = 'active'
        """
        params = []
        
        if provider_id:
            query += " AND id = $1"
            params.append(provider_id)
        elif provider_code:
            query += " AND code = $1"
            params.append(provider_code)
        
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, *params)
            if row:
                return DBProvider(
                    id=str(row['id']),
                    code=row['code'],
                    name=row['name'],
                    api_base_url=row['api_base_url'],
                    auth_type=row['auth_type'],
                    auth_config=row['auth_config'] or {},
                    support_streaming=row['support_streaming'],
                )
        return None
    
    async def list_providers(self) -> List[DBProvider]:
        """获取所有厂商列表"""
        await self.init_pool()
        
        query = """
            SELECT id, code, name, api_base_url, auth_type, auth_config, support_streaming
            FROM ai_providers
            WHERE status = 'active'
            ORDER BY is_domestic DESC, code ASC
        """
        
        async with self.pool.acquire() as conn:
            rows = await conn.fetch(query)
            return [
                DBProvider(
                    id=str(row['id']),
                    code=row['code'],
                    name=row['name'],
                    api_base_url=row['api_base_url'],
                    auth_type=row['auth_type'],
                    auth_config=row['auth_config'] or {},
                    support_streaming=row['support_streaming'],
                )
                for row in rows
            ]
    
    async def get_model(self, model_id: str = None, model_code: str = None, provider_id: str = None) -> Optional[DBModel]:
        """获取模型配置"""
        await self.init_pool()
        
        query = """
            SELECT m.id, m.provider_id, p.code as provider_code,
                   m.model_code, m.model_name, m.model_type,
                   m.max_input_tokens, m.max_output_tokens,
                   m.input_price_per_1k, m.output_price_per_1k,
                   m.capabilities, m.is_default
            FROM ai_models m
            JOIN ai_providers p ON m.provider_id = p.id
            WHERE m.status = 'active' AND p.status = 'active'
        """
        params = []
        
        if model_id:
            query += " AND m.id = $1"
            params.append(model_id)
        elif model_code and provider_id:
            query += " AND m.model_code = $1 AND m.provider_id = $2"
            params.append(model_code)
            params.append(provider_id)
        
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, *params)
            if row:
                return DBModel(
                    id=str(row['id']),
                    provider_id=str(row['provider_id']),
                    provider_code=row['provider_code'],
                    model_code=row['model_code'],
                    model_name=row['model_name'],
                    model_type=row['model_type'],
                    max_input_tokens=row['max_input_tokens'],
                    max_output_tokens=row['max_output_tokens'],
                    input_price_per_1k=row['input_price_per_1k'],
                    output_price_per_1k=row['output_price_per_1k'],
                    capabilities=row['capabilities'] or [],
                    is_default=row['is_default'],
                )
        return None
    
    async def list_models(self, provider_id: str = None, model_type: str = None) -> List[DBModel]:
        """获取模型列表"""
        await self.init_pool()
        
        query = """
            SELECT m.id, m.provider_id, p.code as provider_code,
                   m.model_code, m.model_name, m.model_type,
                   m.max_input_tokens, m.max_output_tokens,
                   m.input_price_per_1k, m.output_price_per_1k,
                   m.capabilities, m.is_default
            FROM ai_models m
            JOIN ai_providers p ON m.provider_id = p.id
            WHERE m.status = 'active' AND p.status = 'active'
        """
        params = []
        
        if provider_id:
            query += " AND m.provider_id = $1"
            params.append(provider_id)
        if model_type:
            query += " AND m.model_type = $2" if params else " AND m.model_type = $1"
            params.append(model_type)
        
        query += " ORDER BY m.priority ASC, m.created_at ASC"
        
        async with self.pool.acquire() as conn:
            rows = await conn.fetch(query, *params)
            return [
                DBModel(
                    id=str(row['id']),
                    provider_id=str(row['provider_id']),
                    provider_code=row['provider_code'],
                    model_code=row['model_code'],
                    model_name=row['model_name'],
                    model_type=row['model_type'],
                    max_input_tokens=row['max_input_tokens'],
                    max_output_tokens=row['max_output_tokens'],
                    input_price_per_1k=row['input_price_per_1k'],
                    output_price_per_1k=row['output_price_per_1k'],
                    capabilities=row['capabilities'] or [],
                    is_default=row['is_default'],
                )
                for row in rows
            ]
    
    async def get_default_model(self, provider_id: str, model_type: str) -> Optional[DBModel]:
        """获取厂商默认模型"""
        await self.init_pool()
        
        query = """
            SELECT m.id, m.provider_id, p.code as provider_code,
                   m.model_code, m.model_name, m.model_type,
                   m.max_input_tokens, m.max_output_tokens,
                   m.input_price_per_1k, m.output_price_per_1k,
                   m.capabilities, m.is_default
            FROM ai_models m
            JOIN ai_providers p ON m.provider_id = p.id
            WHERE m.provider_id = $1 AND m.model_type = $2 
              AND m.is_default = true AND m.status = 'active'
        """
        
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, provider_id)
            if row:
                return DBModel(
                    id=str(row['id']),
                    provider_id=str(row['provider_id']),
                    provider_code=row['provider_code'],
                    model_code=row['model_code'],
                    model_name=row['model_name'],
                    model_type=row['model_type'],
                    max_input_tokens=row['max_input_tokens'],
                    max_output_tokens=row['max_output_tokens'],
                    input_price_per_1k=row['input_price_per_1k'],
                    output_price_per_1k=row['output_price_per_1k'],
                    capabilities=row['capabilities'] or [],
                    is_default=row['is_default'],
                )
        return None
    
    async def get_tenant_api_key(self, tenant_id: str, provider_id: str) -> Optional[TenantAPIKey]:
        """获取租户 API Key（自动解密）"""
        await self.init_pool()
        
        query = """
            SELECT tk.id, tk.tenant_id, tk.provider_id, p.code as provider_code,
                   tk.api_key_value, tk.api_key_encrypted, tk.custom_base_url, tk.custom_headers, tk.status
            FROM tenant_api_keys tk
            JOIN ai_providers p ON tk.provider_id = p.id
            WHERE tk.tenant_id = $1 AND tk.provider_id = $2 AND tk.status = 'active'
        """
        
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, tenant_id, provider_id)
            if row:
                api_key_value = row['api_key_value']
                if row['api_key_encrypted'] and api_key_value:
                    try:
                        api_key_value = aes_cipher.decrypt(api_key_value)
                    except Exception as e:
                        logger.error(f"Failed to decrypt API key: {e}")
                
                return TenantAPIKey(
                    id=str(row['id']),
                    tenant_id=str(row['tenant_id']),
                    provider_id=str(row['provider_id']),
                    provider_code=row['provider_code'],
                    api_key_value=api_key_value,
                    custom_base_url=row['custom_base_url'],
                    custom_headers=row['custom_headers'] or {},
                    status=row['status'],
                )
        return None
    
    async def get_tenant_ai_config(self, tenant_id: str) -> Optional[TenantAIConfig]:
        """获取租户 AI 配置"""
        await self.init_pool()
        
        query = """
            SELECT tenant_id, default_chat_model_id, default_embedding_model_id,
                   default_provider_id, ai_enabled, streaming_enabled
            FROM tenant_ai_configs
            WHERE tenant_id = $1
        """
        
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, tenant_id)
            if row:
                return TenantAIConfig(
                    tenant_id=str(row['tenant_id']),
                    default_chat_model_id=str(row['default_chat_model_id']) if row['default_chat_model_id'] else None,
                    default_embedding_model_id=str(row['default_embedding_model_id']) if row['default_embedding_model_id'] else None,
                    default_provider_id=str(row['default_provider_id']) if row['default_provider_id'] else None,
                    ai_enabled=row['ai_enabled'],
                    streaming_enabled=row['streaming_enabled'],
                )
        return None
    
    async def create_usage_log(self, log_data: Dict[str, Any]):
        """创建使用日志"""
        await self.init_pool()
        
        query = """
            INSERT INTO ai_usage_logs (
                tenant_id, user_id, conversation_id, message_id,
                provider_id, model_id, input_tokens, output_tokens, total_tokens,
                input_cost, output_cost, total_cost, request_type, latency_ms, success, error_message
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
        """
        
        async with self.pool.acquire() as conn:
            await conn.execute(query,
                log_data.get('tenant_id'),
                log_data.get('user_id'),
                log_data.get('conversation_id'),
                log_data.get('message_id'),
                log_data.get('provider_id'),
                log_data.get('model_id'),
                log_data.get('input_tokens', 0),
                log_data.get('output_tokens', 0),
                log_data.get('total_tokens', 0),
                log_data.get('input_cost', 0),
                log_data.get('output_cost', 0),
                log_data.get('total_cost', 0),
                log_data.get('request_type', 'chat'),
                log_data.get('latency_ms', 0),
                log_data.get('success', True),
                log_data.get('error_message'),
            )
    
    async def get_usage_logs(self, tenant_id: str, page: int = 1, page_size: int = 50, 
                              start_date: Optional[str] = None, end_date: Optional[str] = None) -> tuple:
        """查询使用日志"""
        await self.init_pool()
        
        query = """
            SELECT id, tenant_id, user_id, conversation_id, message_id,
                   provider_id, model_id, input_tokens, output_tokens, total_tokens,
                   input_cost, output_cost, total_cost, request_type, latency_ms, 
                   success, error_message, created_at
            FROM ai_usage_logs
            WHERE tenant_id = $1
        """
        params = [tenant_id]
        
        if start_date:
            query += " AND created_at >= $" + str(len(params) + 1)
            params.append(start_date)
        if end_date:
            query += " AND created_at <= $" + str(len(params) + 1)
            params.append(end_date)
        
        query += " ORDER BY created_at DESC"
        
        async with self.pool.acquire() as conn:
            count_query = "SELECT COUNT(*) FROM ai_usage_logs WHERE tenant_id = $1"
            count_params = [tenant_id]
            if start_date:
                count_query += " AND created_at >= $" + str(len(count_params) + 1)
                count_params.append(start_date)
            if end_date:
                count_query += " AND created_at <= $" + str(len(count_params) + 1)
                count_params.append(end_date)
            
            total = await conn.fetchval(count_query, *count_params)
            
            offset = (page - 1) * page_size
            query += " OFFSET $" + str(len(params) + 1) + " LIMIT $" + str(len(params) + 2)
            params.append(offset)
            params.append(page_size)
            
            rows = await conn.fetch(query, *params)
            
            logs = []
            for row in rows:
                logs.append({
                    'id': str(row['id']),
                    'tenant_id': str(row['tenant_id']),
                    'user_id': str(row['user_id']) if row['user_id'] else None,
                    'conversation_id': str(row['conversation_id']) if row['conversation_id'] else None,
                    'message_id': str(row['message_id']) if row['message_id'] else None,
                    'provider_id': str(row['provider_id']) if row['provider_id'] else None,
                    'model_id': str(row['model_id']) if row['model_id'] else None,
                    'input_tokens': row['input_tokens'],
                    'output_tokens': row['output_tokens'],
                    'total_tokens': row['total_tokens'],
                    'input_cost': row['input_cost'],
                    'output_cost': row['output_cost'],
                    'total_cost': row['total_cost'],
                    'request_type': row['request_type'],
                    'latency_ms': row['latency_ms'],
                    'success': row['success'],
                    'error_message': row['error_message'],
                    'created_at': row['created_at'].isoformat() if row['created_at'] else None,
                })
            
            return logs, total
    
    async def get_usage_stats(self, tenant_id: str, start_date: str, end_date: str) -> Dict[str, Any]:
        """获取使用统计"""
        await self.init_pool()
        
        query = """
            SELECT 
                SUM(input_tokens) as total_input_tokens,
                SUM(output_tokens) as total_output_tokens,
                SUM(total_tokens) as total_tokens,
                SUM(input_cost) as total_input_cost,
                SUM(output_cost) as total_output_cost,
                SUM(total_cost) as total_cost,
                COUNT(*) as total_requests,
                SUM(CASE WHEN success THEN 1 ELSE 0 END) as success_requests,
                AVG(latency_ms) as avg_latency_ms
            FROM ai_usage_logs
            WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
        """
        
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, tenant_id, start_date, end_date)
            
            return {
                'total_input_tokens': row['total_input_tokens'] or 0,
                'total_output_tokens': row['total_output_tokens'] or 0,
                'total_tokens': row['total_tokens'] or 0,
                'total_input_cost': row['total_input_cost'] or 0,
                'total_output_cost': row['total_output_cost'] or 0,
                'total_cost': row['total_cost'] or 0,
                'total_requests': row['total_requests'] or 0,
                'success_requests': row['success_requests'] or 0,
                'avg_latency_ms': float(row['avg_latency_ms']) if row['avg_latency_ms'] else 0,
            }


# 全局数据库访问实例
ai_config_db = AIConfigDB()