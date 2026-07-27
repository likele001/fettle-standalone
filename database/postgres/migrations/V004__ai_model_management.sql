-- ==========================================
-- Version: V004
-- Description: AI 模型管理系统数据库结构
-- Date: 2026-07-08
-- 功能说明：
--   1. 平台层面：配置支持的 AI 厂商和模型列表（超管后台管理）
--   2. 租户层面：每个租户独立配置 API Key 和默认模型（租户后台管理）
-- ==========================================

BEGIN;

-- ==========================================
-- 1. ai_providers 表 - AI 厂商配置（平台层面）
-- ==========================================
CREATE TABLE IF NOT EXISTS ai_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,              -- 厂商代码: openai, anthropic, aliyun, baidu, tencent, google
    name VARCHAR(100) NOT NULL,                    -- 厂商名称: OpenAI, Anthropic, 阿里云, 百度, 腾讯, Google
    name_en VARCHAR(100),                          -- 英文名称
    logo_url VARCHAR(500),                         -- Logo 图片 URL
    description TEXT,                              -- 厂商描述
    api_base_url VARCHAR(500),                     -- API 基础 URL（如 https://api.openai.com/v1）
    auth_type VARCHAR(20) NOT NULL DEFAULT 'api_key', -- 认证方式: api_key, oauth, custom
    auth_config JSONB DEFAULT '{}',                -- 认证配置（如 header 格式）
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- 状态: active, inactive
    is_domestic BOOLEAN NOT NULL DEFAULT false,    -- 是否国内厂商（用于合规判断）
    support_streaming BOOLEAN NOT NULL DEFAULT true, -- 是否支持流式输出
    support_vision BOOLEAN NOT NULL DEFAULT false, -- 是否支持图片输入
    support_function_call BOOLEAN NOT NULL DEFAULT false, -- 是否支持函数调用
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE ai_providers IS 'AI 厂商配置表（平台层面，超管后台管理）';
COMMENT ON COLUMN ai_providers.code IS '厂商代码，唯一标识：openai, anthropic, aliyun, baidu, tencent, google, minimax, moonshot, zhipu, deepseek';
COMMENT ON COLUMN ai_providers.is_domestic IS '是否国内厂商，用于合规判断和数据出境控制';
COMMENT ON COLUMN ai_providers.auth_config IS '认证配置，如 header 格式、签名方式等';

-- 插入默认 AI 厂商配置
INSERT INTO ai_providers (code, name, name_en, logo_url, description, api_base_url, is_domestic, support_streaming, support_vision, support_function_call) VALUES
-- 国内厂商
('aliyun', '阿里云', 'Alibaba Cloud', NULL, '阿里云通义千问大模型服务', 'https://dashscope.aliyuncs.com/api/v1', true, true, true, true),
('baidu', '百度智能云', 'Baidu Cloud', NULL, '百度文心一言大模型服务', 'https://aip.baidubce.com/rpc/2.0/ai_custom/v1', true, true, true, true),
('tencent', '腾讯云', 'Tencent Cloud', NULL, '腾讯混元大模型服务', 'https://api.hunyuanao.tencentcloudapi.com', true, true, false, true),
('minimax', 'MiniMax', 'MiniMax', NULL, 'MiniMax 大模型服务', 'https://api.minimax.chat/v1', true, true, true, true),
('moonshot', 'Moonshot AI', 'Moonshot AI', NULL, 'Moonshot Kimi 大模型服务', 'https://api.moonshot.cn/v1', true, true, true, true),
('zhipu', '智谱 AI', 'Zhipu AI', NULL, '智谱 GLM 大模型服务', 'https://open.bigmodel.cn/api/paas/v4', true, true, true, true),
('deepseek', 'DeepSeek', 'DeepSeek', NULL, 'DeepSeek 大模型服务', 'https://api.deepseek.com/v1', true, true, true, true),
-- 国外厂商
('openai', 'OpenAI', 'OpenAI', NULL, 'OpenAI GPT 系列大模型服务', 'https://api.openai.com/v1', false, true, true, true),
('anthropic', 'Anthropic', 'Anthropic', NULL, 'Anthropic Claude 系列大模型服务', 'https://api.anthropic.com/v1', false, true, true, true),
('google', 'Google AI', 'Google AI', NULL, 'Google Gemini 系列大模型服务', 'https://generativelanguage.googleapis.com/v1beta', false, true, true, true)
ON CONFLICT (code) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_ai_providers_code ON ai_providers(code);
CREATE INDEX IF NOT EXISTS idx_ai_providers_status ON ai_providers(status);

-- ==========================================
-- 2. ai_models 表 - AI 模型配置（平台层面）
-- ==========================================
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES ai_providers(id) ON DELETE CASCADE,
    model_code VARCHAR(100) NOT NULL,              -- 模型代码: qwen-max, gpt-4, claude-3-opus
    model_name VARCHAR(200) NOT NULL,              -- 模型显示名称: 通义千问-Max, GPT-4, Claude 3 Opus
    model_type VARCHAR(20) NOT NULL DEFAULT 'chat', -- 模型类型: chat, embedding, image, audio, video
    max_input_tokens INTEGER NOT NULL DEFAULT 4096, -- 最大输入 tokens
    max_output_tokens INTEGER NOT NULL DEFAULT 2048, -- 最大输出 tokens
    input_price_per_1k FLOAT NOT NULL DEFAULT 0.01, -- 输入价格（每千 tokens，美元）
    output_price_per_1k FLOAT NOT NULL DEFAULT 0.03, -- 输出价格（每千 tokens，美元）
    capabilities JSONB NOT NULL DEFAULT '[]',      -- 能力列表: ["chat", "vision", "function_call", "code"]
    description TEXT,                              -- 模型描述
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- 状态: active, deprecated, beta
    is_default BOOLEAN NOT NULL DEFAULT false,     -- 是否为该厂商默认模型
    priority INTEGER NOT NULL DEFAULT 0,           -- 排序优先级（数字越小越靠前）
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider_id, model_code)
);

COMMENT ON TABLE ai_models IS 'AI 模型配置表（平台层面，超管后台管理）';
COMMENT ON COLUMN ai_models.model_type IS '模型类型: chat(对话), embedding(向量), image(图片生成), audio(语音), video(视频)';
COMMENT ON COLUMN ai_models.capabilities IS '能力列表，JSON 数组格式，如 ["chat", "vision", "function_call", "code", "reasoning"]';
COMMENT ON COLUMN ai_models.is_default IS '是否为该厂商的默认推荐模型';

-- 插入默认 AI 模型配置
-- 阿里云通义千问
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='aliyun'), 'qwen-turbo', '通义千问-Turbo', 'chat', 4096, 2048, 0.002, 0.006, '["chat", "code"]', false, 1),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'qwen-plus', '通义千问-Plus', 'chat', 32768, 2048, 0.004, 0.012, '["chat", "code", "reasoning"]', false, 2),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'qwen-max', '通义千问-Max', 'chat', 32768, 2048, 0.02, 0.06, '["chat", "vision", "function_call", "code", "reasoning"]', true, 0),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'qwen-max-longcontext', '通义千问-Max-长文本', 'chat', 28000, 2048, 0.02, 0.06, '["chat", "long_context"]', false, 3),
((SELECT id FROM ai_providers WHERE code='aliyun'), 'text-embedding-v2', '通义文本向量-v2', 'embedding', 2048, 0, 0.0007, 0, '["embedding"]', true, 0)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- 百度文心一言
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='baidu'), 'ernie-bot-4', '文心一言 4.0', 'chat', 8192, 2048, 0.03, 0.06, '["chat", "reasoning"]', true, 0),
((SELECT id FROM ai_providers WHERE code='baidu'), 'ernie-bot-3.5', '文心一言 3.5', 'chat', 4096, 2048, 0.004, 0.008, '["chat"]', false, 1),
((SELECT id FROM ai_providers WHERE code='baidu'), 'ernie-bot-turbo', '文心一言 Turbo', 'chat', 4096, 2048, 0.001, 0.002, '["chat"]', false, 2)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- 腾讯混元
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='tencent'), 'hunyuan-lite', '混元 Lite', 'chat', 4096, 2048, 0.001, 0.002, '["chat"]', false, 2),
((SELECT id FROM ai_providers WHERE code='tencent'), 'hunyuan-standard', '混元 Standard', 'chat', 8192, 2048, 0.004, 0.008, '["chat", "function_call"]', true, 0),
((SELECT id FROM ai_providers WHERE code='tencent'), 'hunyuan-pro', '混元 Pro', 'chat', 8192, 2048, 0.01, 0.02, '["chat", "function_call", "reasoning"]', false, 1)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- OpenAI GPT
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='openai'), 'gpt-4o', 'GPT-4o', 'chat', 128000, 4096, 0.005, 0.015, '["chat", "vision", "function_call", "code"]', true, 0),
((SELECT id FROM ai_providers WHERE code='openai'), 'gpt-4o-mini', 'GPT-4o Mini', 'chat', 128000, 4096, 0.00015, 0.0006, '["chat", "vision", "function_call"]', false, 1),
((SELECT id FROM ai_providers WHERE code='openai'), 'gpt-4-turbo', 'GPT-4 Turbo', 'chat', 128000, 4096, 0.01, 0.03, '["chat", "vision", "function_call", "code"]', false, 2),
((SELECT id FROM ai_providers WHERE code='openai'), 'gpt-3.5-turbo', 'GPT-3.5 Turbo', 'chat', 16385, 4096, 0.0005, 0.0015, '["chat", "function_call"]', false, 3),
((SELECT id FROM ai_providers WHERE code='openai'), 'text-embedding-3-large', 'Embedding Large', 'embedding', 8191, 0, 0.00013, 0, '["embedding"]', true, 0),
((SELECT id FROM ai_providers WHERE code='openai'), 'text-embedding-3-small', 'Embedding Small', 'embedding', 8191, 0, 0.00002, 0, '["embedding"]', false, 1)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- Anthropic Claude
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='anthropic'), 'claude-3-opus', 'Claude 3 Opus', 'chat', 200000, 4096, 0.015, 0.075, '["chat", "vision", "code", "reasoning"]', false, 1),
((SELECT id FROM ai_providers WHERE code='anthropic'), 'claude-3-sonnet', 'Claude 3 Sonnet', 'chat', 200000, 4096, 0.003, 0.015, '["chat", "vision", "function_call", "code"]', true, 0),
((SELECT id FROM ai_providers WHERE code='anthropic'), 'claude-3-haiku', 'Claude 3 Haiku', 'chat', 200000, 4096, 0.00025, 0.00125, '["chat", "vision"]', false, 2),
((SELECT id FROM ai_providers WHERE code='anthropic'), 'claude-3.5-sonnet', 'Claude 3.5 Sonnet', 'chat', 200000, 8192, 0.003, 0.015, '["chat", "vision", "function_call", "code", "reasoning"]', false, 0)
ON CONFLICT (provider_id, model_code) DO NOTHING;

-- DeepSeek
INSERT INTO ai_models (provider_id, model_code, model_name, model_type, max_input_tokens, max_output_tokens, input_price_per_1k, output_price_per_1k, capabilities, is_default, priority) VALUES
((SELECT id FROM ai_providers WHERE code='deepseek'), 'deepseek-chat', 'DeepSeek Chat', 'chat', 32000, 4096, 0.00014, 0.00028, '["chat", "code"]', true, 0),
((SELECT id FROM ai_providers WHERE code='deepseek'), 'deepseek-coder', 'DeepSeek Coder', 'chat', 16000, 4096, 0.00014, 0.00028, '["chat", "code"]', false, 1)
ON CONFLICT (provider_id, model_code) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_ai_models_provider_id ON ai_models(provider_id);
CREATE INDEX IF NOT EXISTS idx_ai_models_model_code ON ai_models(model_code);
CREATE INDEX IF NOT EXISTS idx_ai_models_model_type ON ai_models(model_type);
CREATE INDEX IF NOT EXISTS idx_ai_models_status ON ai_models(status);

-- ==========================================
-- 3. tenant_ai_configs 表 - 租户 AI 配置（租户层面）
-- ==========================================
CREATE TABLE IF NOT EXISTS tenant_ai_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    -- 默认模型配置
    default_chat_model_id UUID REFERENCES ai_models(id),     -- 默认对话模型
    default_embedding_model_id UUID REFERENCES ai_models(id), -- 默认向量模型
    default_provider_id UUID REFERENCES ai_providers(id),    -- 默认厂商（备用）
    -- 全局开关
    ai_enabled BOOLEAN NOT NULL DEFAULT true,                -- 是否启用 AI 功能
    streaming_enabled BOOLEAN NOT NULL DEFAULT true,         -- 是否启用流式输出
    vision_enabled BOOLEAN NOT NULL DEFAULT false,           -- 是否启用图片理解
    function_call_enabled BOOLEAN NOT NULL DEFAULT false,    -- 是否启用函数调用
    -- 额度控制
    monthly_token_limit INTEGER NOT NULL DEFAULT 100000,     -- 每月 token 额度（0 表示无限制）
    monthly_token_used INTEGER NOT NULL DEFAULT 0,           -- 本月已使用 token
    token_limit_reset_at TIMESTAMPTZ,                        -- 额度重置时间
    -- 速率限制
    rate_limit_per_minute INTEGER NOT NULL DEFAULT 60,       -- 每分钟请求限制
    rate_limit_per_day INTEGER NOT NULL DEFAULT 1000,        -- 每天请求限制
    -- 其他配置
    config JSONB NOT NULL DEFAULT '{}',                      -- 扩展配置（JSON 格式）
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE tenant_ai_configs IS '租户 AI 配置表（租户层面，租户后台管理）';
COMMENT ON COLUMN tenant_ai_configs.default_chat_model_id IS '默认对话模型 ID，如未设置则使用平台默认';
COMMENT ON COLUMN tenant_ai_configs.monthly_token_limit IS '每月 token 额度，0 表示无限制（由套餐决定）';
COMMENT ON COLUMN tenant_ai_configs.config IS '扩展配置，JSON 格式，可存储自定义参数';

CREATE INDEX IF NOT EXISTS idx_tenant_ai_configs_tenant_id ON tenant_ai_configs(tenant_id);

-- ==========================================
-- 4. tenant_api_keys 表 - 租户 API Key 配置（租户层面）
-- ==========================================
CREATE TABLE IF NOT EXISTS tenant_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES ai_providers(id) ON DELETE CASCADE,
    api_key_name VARCHAR(100) NOT NULL,                     -- API Key 名称（便于管理）
    api_key_value TEXT NOT NULL,                            -- API Key 值（加密存储）
    api_key_encrypted BOOLEAN NOT NULL DEFAULT false,       -- 是否加密存储
    -- 配额限制（针对单个厂商）
    monthly_quota INTEGER NOT NULL DEFAULT 0,               -- 每月配额（0 表示使用租户全局配额）
    monthly_used INTEGER NOT NULL DEFAULT 0,                -- 本月已使用
    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'active',           -- 状态: active, disabled, expired
    expired_at TIMESTAMPTZ,                                 -- 过期时间
    -- 其他配置
    custom_base_url VARCHAR(500),                           -- 自定义 API 地址（代理或私有部署）
    custom_headers JSONB DEFAULT '{}',                      -- 自定义请求头
    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,                               -- 最后使用时间
    UNIQUE(tenant_id, provider_id, api_key_name)
);

COMMENT ON TABLE tenant_api_keys IS '租户 API Key 配置表（租户后台管理）';
COMMENT ON COLUMN tenant_api_keys.api_key_value IS 'API Key 值，建议加密存储';
COMMENT ON COLUMN tenant_api_keys.custom_base_url IS '自定义 API 地址，用于代理或私有部署场景';

CREATE INDEX IF NOT EXISTS idx_tenant_api_keys_tenant_id ON tenant_api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_api_keys_provider_id ON tenant_api_keys(provider_id);
CREATE INDEX IF NOT EXISTS idx_tenant_api_keys_status ON tenant_api_keys(status);

-- ==========================================
-- 5. ai_usage_logs 表 - AI 使用日志（用于计费和统计）
-- ==========================================
CREATE TABLE IF NOT EXISTS ai_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID,                                           -- 用户 ID
    conversation_id UUID,                                   -- 会话 ID
    message_id UUID,                                        -- 消息 ID
    provider_id UUID NOT NULL REFERENCES ai_providers(id),
    model_id UUID NOT NULL REFERENCES ai_models(id),
    -- Token 使用量
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    -- 费用（美元）
    input_cost FLOAT NOT NULL DEFAULT 0,
    output_cost FLOAT NOT NULL DEFAULT 0,
    total_cost FLOAT NOT NULL DEFAULT 0,
    -- 请求信息
    request_type VARCHAR(20) NOT NULL,                      -- 请求类型: chat, embedding, function_call
    latency_ms INTEGER NOT NULL DEFAULT 0,                  -- 响应延迟（毫秒）
    success BOOLEAN NOT NULL DEFAULT true,                  -- 是否成功
    error_message TEXT,                                     -- 错误信息
    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE ai_usage_logs IS 'AI 使用日志表（用于计费和统计）';
COMMENT ON COLUMN ai_usage_logs.total_cost IS '总费用（美元），用于计费';
COMMENT ON COLUMN ai_usage_logs.request_type IS '请求类型: chat(对话), embedding(向量), function_call(函数调用)';

CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_tenant_id ON ai_usage_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_created_at ON ai_usage_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_provider_id ON ai_usage_logs(provider_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_model_id ON ai_usage_logs(model_id);

-- ==========================================
-- 6. 为默认租户插入 AI 配置
-- ==========================================
INSERT INTO tenant_ai_configs (tenant_id, ai_enabled, streaming_enabled, monthly_token_limit) VALUES
((SELECT id FROM tenants WHERE id='00000000-0000-0000-0000-000000000001'), true, true, 0)
ON CONFLICT (tenant_id) DO NOTHING;

COMMIT;