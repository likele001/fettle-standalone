-- ==========================================
-- Version: V003
-- Description: 聊天服务数据库表 - conversations, messages, channels
-- Date: 2026-07-08
-- 迁移命令：
--   psql -U ai_platform -d ai_platform -h 127.0.0.1 \
--     -f database/postgres/migrations/V003__chat_service.sql
-- 回滚命令：
--   psql -U ai_platform -d ai_platform -h 127.0.0.1 \
--     -f database/postgres/migrations/V003__rollback.sql
-- ==========================================

BEGIN;

-- ==========================================
-- 1. conversations 表 - 会话表
-- ==========================================
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    agent_id UUID,
    channel VARCHAR(20) NOT NULL DEFAULT 'web',
    channel_id VARCHAR(100),
    customer_id VARCHAR(100),
    customer_name VARCHAR(100),
    customer_avatar VARCHAR(500),
    title VARCHAR(200),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    priority INTEGER NOT NULL DEFAULT 0,
    tags VARCHAR(500),
    assigned_to UUID,
    last_message_at TIMESTAMPTZ,
    message_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_conversations_tenant_id ON conversations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_agent_id ON conversations(agent_id);
CREATE INDEX IF NOT EXISTS idx_conversations_customer_id ON conversations(customer_id);
CREATE INDEX IF NOT EXISTS idx_conversations_status ON conversations(status);
CREATE INDEX IF NOT EXISTS idx_conversations_last_message_at ON conversations(last_message_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_conversations_deleted_at ON conversations(deleted_at);

COMMENT ON TABLE conversations IS '会话表';
COMMENT ON COLUMN conversations.tenant_id IS '租户 ID';
COMMENT ON COLUMN conversations.user_id IS '用户 ID';
COMMENT ON COLUMN conversations.agent_id IS '智能体 ID';
COMMENT ON COLUMN conversations.channel IS '渠道: web, wechat, feishu, dingtalk, wecom, douyin';
COMMENT ON COLUMN conversations.channel_id IS '渠道标识';
COMMENT ON COLUMN conversations.customer_id IS '客户 ID';
COMMENT ON COLUMN conversations.customer_name IS '客户名称';
COMMENT ON COLUMN conversations.customer_avatar IS '客户头像';
COMMENT ON COLUMN conversations.title IS '会话标题';
COMMENT ON COLUMN conversations.status IS '状态: active-进行中, closed-已关闭';
COMMENT ON COLUMN conversations.priority IS '优先级';
COMMENT ON COLUMN conversations.assigned_to IS '分配给的坐席 ID';
COMMENT ON COLUMN conversations.last_message_at IS '最后消息时间';
COMMENT ON COLUMN conversations.message_count IS '消息数量';

-- ==========================================
-- 2. messages 表 - 消息表
-- ==========================================
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    sender_id VARCHAR(100),
    sender_type VARCHAR(20) NOT NULL,
    sender_name VARCHAR(100),
    content_type VARCHAR(20) NOT NULL DEFAULT 'text',
    content TEXT NOT NULL,
    metadata TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'sent',
    model_used VARCHAR(100),
    tokens_used INTEGER NOT NULL DEFAULT 0,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_tenant_id ON messages(tenant_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_type ON messages(sender_type);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_content ON messages USING gin (to_tsvector('simple', content));

COMMENT ON TABLE messages IS '消息表';
COMMENT ON COLUMN messages.conversation_id IS '会话 ID';
COMMENT ON COLUMN messages.tenant_id IS '租户 ID';
COMMENT ON COLUMN messages.sender_id IS '发送者 ID';
COMMENT ON COLUMN messages.sender_type IS '发送者类型: user, agent, assistant, system';
COMMENT ON COLUMN messages.sender_name IS '发送者名称';
COMMENT ON COLUMN messages.content_type IS '内容类型: text, image, file, card';
COMMENT ON COLUMN messages.content IS '消息内容';
COMMENT ON COLUMN messages.metadata IS '元数据，JSON 格式';
COMMENT ON COLUMN messages.status IS '状态: sending-发送中, sent-已发送, failed-失败, read-已读';
COMMENT ON COLUMN messages.model_used IS '使用的 AI 模型';
COMMENT ON COLUMN messages.tokens_used IS '消耗的 token 数';
COMMENT ON COLUMN messages.latency_ms IS '响应延迟（毫秒）';

-- ==========================================
-- 3. channels 表 - 渠道配置表
-- ==========================================
CREATE TABLE IF NOT EXISTS channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    config TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    webhook_secret VARCHAR(200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_channels_tenant_id ON channels(tenant_id);
CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(type);
CREATE INDEX IF NOT EXISTS idx_channels_status ON channels(status);
CREATE INDEX IF NOT EXISTS idx_channels_deleted_at ON channels(deleted_at);

COMMENT ON TABLE channels IS '渠道配置表';
COMMENT ON COLUMN channels.tenant_id IS '租户 ID';
COMMENT ON COLUMN channels.name IS '渠道名称';
COMMENT ON COLUMN channels.type IS '渠道类型: wechat, feishu, dingtalk, wecom, douyin, web';
COMMENT ON COLUMN channels.config IS '渠道配置，JSON 格式';
COMMENT ON COLUMN channels.status IS '状态: active-启用, inactive-停用';
COMMENT ON COLUMN channels.webhook_secret IS 'Webhook 密钥';

COMMIT;
