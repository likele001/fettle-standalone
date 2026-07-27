"""工作记忆 - 当前对话上下文"""
import logging
from typing import Dict, Any, List, Optional
from dataclasses import dataclass, field
from datetime import datetime

logger = logging.getLogger(__name__)


@dataclass
class WorkingMemory:
    """工作记忆"""
    conversation_id: str
    tenant_id: str
    user_id: str
    agent_id: str
    messages: List[Dict[str, Any]] = field(default_factory=list)
    context: Dict[str, Any] = field(default_factory=dict)
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)
    
    def add_message(self, role: str, content: str, metadata: Optional[Dict[str, Any]] = None):
        """添加消息"""
        message = {
            "role": role,
            "content": content,
            "timestamp": datetime.now().isoformat(),
            "metadata": metadata or {}
        }
        self.messages.append(message)
        self.updated_at = datetime.now()
    
    def get_recent_messages(self, limit: int = 10) -> List[Dict[str, Any]]:
        """获取最近的消息"""
        return self.messages[-limit:] if len(self.messages) > limit else self.messages
    
    def set_context(self, key: str, value: Any):
        """设置上下文"""
        self.context[key] = value
        self.updated_at = datetime.now()
    
    def get_context(self, key: str, default: Any = None) -> Any:
        """获取上下文"""
        return self.context.get(key, default)
    
    def clear(self):
        """清空记忆"""
        self.messages.clear()
        self.context.clear()
        self.updated_at = datetime.now()


class WorkingMemoryManager:
    """工作记忆管理器"""
    
    def __init__(self):
        self.memories: Dict[str, WorkingMemory] = {}
    
    def get_or_create(
        self,
        conversation_id: str,
        tenant_id: str,
        user_id: str,
        agent_id: str
    ) -> WorkingMemory:
        """获取或创建工作记忆"""
        if conversation_id not in self.memories:
            self.memories[conversation_id] = WorkingMemory(
                conversation_id=conversation_id,
                tenant_id=tenant_id,
                user_id=user_id,
                agent_id=agent_id
            )
            logger.info(f"Created working memory for conversation: {conversation_id}")
        
        return self.memories[conversation_id]
    
    def get(self, conversation_id: str) -> Optional[WorkingMemory]:
        """获取工作记忆"""
        return self.memories.get(conversation_id)
    
    def delete(self, conversation_id: str):
        """删除工作记忆"""
        if conversation_id in self.memories:
            del self.memories[conversation_id]
            logger.info(f"Deleted working memory for conversation: {conversation_id}")
    
    def cleanup_expired(self, max_age_hours: int = 24):
        """清理过期记忆"""
        now = datetime.now()
        expired = [
            conv_id for conv_id, memory in self.memories.items()
            if (now - memory.updated_at).total_seconds() > max_age_hours * 3600
        ]
        
        for conv_id in expired:
            del self.memories[conv_id]
        
        if expired:
            logger.info(f"Cleaned up {len(expired)} expired working memories")


# 全局实例
working_memory_manager = WorkingMemoryManager()
