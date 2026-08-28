# -*- coding: utf-8 -*-
"""客户级摘要记忆：Redis 存取 + LLM 滚动更新

设计：
- key: customer_memory:{tenant_id}:{user_id}，TTL 90 天
- 每次对话回复完成后，用 LLM 结合旧摘要 + 最新对话更新客户档案
- 客户再次对话时（任意会话），把档案注入 system prompt，实现跨会话记忆
- 所有异常吞掉，绝不影响主对话
"""
import json
import logging
from typing import List, Dict, Any, Optional

import redis.asyncio as aioredis

from config.settings import settings
from core.models.providers import ProviderFactory, ChatMessage

logger = logging.getLogger(__name__)

CUSTOMER_MEMORY_TTL = 90 * 24 * 3600  # 90 天


class CustomerMemory:
    """客户摘要记忆管理器"""

    def __init__(self):
        self._redis: Optional[aioredis.Redis] = None

    async def _get_redis(self) -> aioredis.Redis:
        if self._redis is None:
            self._redis = aioredis.from_url(settings.redis_url, decode_responses=True)
        return self._redis

    @staticmethod
    def _key(tenant_id: str, user_id: str) -> str:
        return f"customer_memory:{tenant_id}:{user_id}"

    async def get_memory(self, tenant_id: str, user_id: str) -> str:
        """读取客户摘要（供注入），失败静默返回空"""
        if not user_id or user_id in ("test", "system", "anonymous"):
            return ""
        try:
            r = await self._get_redis()
            val = await r.get(self._key(tenant_id, user_id))
            return val or ""
        except Exception as e:
            logger.warning(f"get customer memory failed: {e}")
            return ""

    async def delete_memory(self, tenant_id: str, user_id: str) -> None:
        """删除客户记忆（隐私需求）"""
        try:
            r = await self._get_redis()
            await r.delete(self._key(tenant_id, user_id))
            logger.info(f"customer memory deleted: {tenant_id[:8]}:{user_id[:8]}")
        except Exception as e:
            logger.warning(f"delete customer memory failed: {e}")

    async def update_memory(
        self,
        tenant_id: str,
        user_id: str,
        new_messages: List[Dict[str, str]],
        old_summary: str = "",
    ) -> None:
        """用 LLM 更新客户摘要（失败只记日志，不抛出）"""
        if not user_id or user_id in ("test", "system", "anonymous"):
            return
        if not new_messages:
            return
        try:
            # 未提供旧摘要时，从 Redis 读现有摘要（滚动更新）
            if not old_summary:
                old_summary = await self.get_memory(tenant_id, user_id)
            provider_code, model_name = await self._select_summarizer(tenant_id)
            provider = await ProviderFactory.get_provider_async(provider_code, tenant_id=tenant_id)
            if not provider:
                logger.warning("no provider for summarization")
                return

            # 只取最近 6 条消息做增量摘要（控制成本）
            recent = new_messages[-6:]
            convo_text = "\n".join(
                f"{m.get('role', 'user')}: {str(m.get('content', ''))[:500]}" for m in recent
            )

            prompt = (
                "你是一位客户档案助手。请根据【旧摘要】和【最新对话】，生成/更新一份客户档案摘要。\n"
                "要求：\n"
                "1. 包含：客户身份线索、关心的话题、需求与偏好、达成的约定/承诺\n"
                "2. 语言简洁（120-200 字中文），只输出摘要本身，不要任何解释\n"
                "3. 旧摘要中有价值的信息要保留，过时信息可更新\n\n"
                f"【旧摘要】\n{old_summary if old_summary else '（无）'}\n\n"
                f"【最新对话】\n{convo_text}"
            )

            resp = await provider.chat(
                messages=[ChatMessage(role="user", content=prompt)],
                model=model_name,
                max_tokens=300,
                temperature=0.3,
            )
            summary = (resp.content or "").strip()
            if not summary:
                return

            r = await self._get_redis()
            await r.setex(self._key(tenant_id, user_id), CUSTOMER_MEMORY_TTL, summary)
            logger.info(
                f"customer memory updated: tenant={tenant_id[:8]} user={user_id[:8]} len={len(summary)}"
            )
        except Exception as e:
            logger.error(f"update customer memory failed: {e}")

    @staticmethod
    async def _select_summarizer(tenant_id: str):
        """选摘要模型：优先租户配置，fallback 到 yaml 默认"""
        try:
            from core.models.db_model_router import db_model_router
            cfg = await db_model_router.select_model(tenant_id=tenant_id)
            return cfg.provider_code, cfg.model_code
        except Exception:
            from core.models.model_router import model_router, TaskType
            cfg = model_router.select_model(task_type=TaskType.KNOWLEDGE_QA, tenant_plan=None)
            return cfg.provider, cfg.model_name


# 全局实例
customer_memory = CustomerMemory()
