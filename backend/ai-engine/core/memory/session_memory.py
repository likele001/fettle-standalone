"""短期记忆 - 基于 Redis 的会话记忆"""
import json
import logging
from typing import Optional, Dict, Any
from datetime import datetime

import redis.asyncio as aioredis

from config.settings import settings

logger = logging.getLogger(__name__)


class SessionData:
    """会话数据"""

    def __init__(
        self,
        conversation_id: str,
        tenant_id: str,
        user_id: str,
        agent_id: str,
        summary: str = "",
        preferences: Optional[Dict[str, Any]] = None,
    ):
        self.conversation_id = conversation_id
        self.tenant_id = tenant_id
        self.user_id = user_id
        self.agent_id = agent_id
        self.summary = summary
        self.preferences = preferences or {}
        self.last_active_at = datetime.now()
        self.created_at = datetime.now()

    def to_dict(self) -> Dict[str, Any]:
        return {
            "conversation_id": self.conversation_id,
            "tenant_id": self.tenant_id,
            "user_id": self.user_id,
            "agent_id": self.agent_id,
            "summary": self.summary,
            "preferences": self.preferences,
            "last_active_at": self.last_active_at.isoformat(),
            "created_at": self.created_at.isoformat(),
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SessionData":
        obj = cls(
            conversation_id=data["conversation_id"],
            tenant_id=data["tenant_id"],
            user_id=data["user_id"],
            agent_id=data["agent_id"],
            summary=data.get("summary", ""),
            preferences=data.get("preferences", {}),
        )
        obj.last_active_at = datetime.fromisoformat(data["last_active_at"])
        obj.created_at = datetime.fromisoformat(data["created_at"])
        return obj


class SessionMemory:
    """短期记忆管理器（Redis）"""

    def __init__(self, redis_client: Optional[aioredis.Redis] = None):
        self.redis = redis_client
        self.ttl = 86400  # 24 小时

    async def _get_redis(self) -> aioredis.Redis:
        if self.redis is None:
            self.redis = aioredis.from_url(
                settings.redis_url,
                decode_responses=True,
            )
        return self.redis

    async def save_session(self, data: SessionData) -> None:
        """保存会话记忆"""
        r = await self._get_redis()
        key = f"session:{data.conversation_id}"
        await r.setex(key, self.ttl, json.dumps(data.to_dict(), ensure_ascii=False))
        logger.info(f"Saved session memory: {data.conversation_id}")

    async def get_session(self, conversation_id: str) -> Optional[SessionData]:
        """获取会话记忆"""
        r = await self._get_redis()
        key = f"session:{conversation_id}"
        result = await r.get(key)
        if not result:
            return None
        return SessionData.from_dict(json.loads(result))

    async def update_summary(self, conversation_id: str, summary: str) -> None:
        """更新会话摘要"""
        data = await self.get_session(conversation_id)
        if not data:
            raise ValueError(f"Session not found: {conversation_id}")
        data.summary = summary
        data.last_active_at = datetime.now()
        await self.save_session(data)

    async def update_preferences(
        self, conversation_id: str, preferences: Dict[str, Any]
    ) -> None:
        """更新用户偏好"""
        data = await self.get_session(conversation_id)
        if not data:
            raise ValueError(f"Session not found: {conversation_id}")
        data.preferences.update(preferences)
        data.last_active_at = datetime.now()
        await self.save_session(data)

    async def delete_session(self, conversation_id: str) -> None:
        """删除会话记忆"""
        r = await self._get_redis()
        key = f"session:{conversation_id}"
        await r.delete(key)
        logger.info(f"Deleted session memory: {conversation_id}")

    async def extend_ttl(self, conversation_id: str) -> None:
        """延长过期时间"""
        r = await self._get_redis()
        key = f"session:{conversation_id}"
        await r.expire(key, self.ttl)


# 全局实例
session_memory = SessionMemory()
