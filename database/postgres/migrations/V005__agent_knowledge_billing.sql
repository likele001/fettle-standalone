-- ==========================================
-- Version: V005
-- Description: 智能体、知识库、技能、计费系统数据库表（增量迁移）
-- Date: 2026-07-08
-- ==========================================

BEGIN;

-- ==========================================
-- 1. agents 表 - 智能体配置表（可能已存在）
-- ==========================================
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    avatar_url VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'idle',
    agent_type VARCHAR(50) NOT NULL DEFAULT 'chat',
    personality_config JSONB DEFAULT '{}',
    workflow_config JSONB DEFAULT '{}',
    welcome_message VARCHAR(500),
    work_hours JSONB DEFAULT '{}',
    model_id VARCHAR(50) DEFAULT 'default',
    knowledge_base_ids UUID[] DEFAULT '{}',
    max_concurrent INTEGER NOT NULL DEFAULT 10,
    current_sessions INTEGER NOT NULL DEFAULT 0,
    total_conversations BIGINT NOT NULL DEFAULT 0,
    total_messages BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'agents' AND indexname = 'idx_agents_tenant_id') THEN
        CREATE INDEX idx_agents_tenant_id ON agents(tenant_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'agents' AND indexname = 'idx_agents_status') THEN
        CREATE INDEX idx_agents_status ON agents(status);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'agents' AND indexname = 'idx_agents_deleted_at') THEN
        CREATE INDEX idx_agents_deleted_at ON agents(deleted_at);
    END IF;
END $$;

-- ==========================================
-- 2. knowledge_bases 表 - 知识库表（可能已存在）
-- ==========================================
CREATE TABLE IF NOT EXISTS knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    doc_count INTEGER NOT NULL DEFAULT 0,
    chunk_count INTEGER NOT NULL DEFAULT 0,
    total_size BIGINT NOT NULL DEFAULT 0,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'knowledge_bases' AND indexname = 'idx_knowledge_bases_tenant_id') THEN
        CREATE INDEX idx_knowledge_bases_tenant_id ON knowledge_bases(tenant_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'knowledge_bases' AND indexname = 'idx_knowledge_bases_status') THEN
        CREATE INDEX idx_knowledge_bases_status ON knowledge_bases(status);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'knowledge_bases' AND indexname = 'idx_knowledge_bases_deleted_at') THEN
        CREATE INDEX idx_knowledge_bases_deleted_at ON knowledge_bases(deleted_at);
    END IF;
END $$;

-- ==========================================
-- 3. knowledge_documents 表 - 知识库文档表（可能已存在）
-- ==========================================
CREATE TABLE IF NOT EXISTS knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    file_size BIGINT NOT NULL,
    file_url VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    chunk_count INTEGER NOT NULL DEFAULT 0,
    error_message VARCHAR(500),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'knowledge_documents' AND indexname = 'idx_knowledge_documents_kb_id') THEN
        CREATE INDEX idx_knowledge_documents_kb_id ON knowledge_documents(knowledge_base_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'knowledge_documents' AND indexname = 'idx_knowledge_documents_tenant_id') THEN
        CREATE INDEX idx_knowledge_documents_tenant_id ON knowledge_documents(tenant_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'knowledge_documents' AND indexname = 'idx_knowledge_documents_status') THEN
        CREATE INDEX idx_knowledge_documents_status ON knowledge_documents(status);
    END IF;
END $$;

-- ==========================================
-- 4. skills 表 - 添加缺失列
-- ==========================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'skills' AND column_name = 'skill_type') THEN
        ALTER TABLE skills ADD COLUMN skill_type VARCHAR(20) NOT NULL DEFAULT 'api';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'skills' AND column_name = 'schema') THEN
        ALTER TABLE skills ADD COLUMN schema JSONB DEFAULT '{}';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'skills' AND column_name = 'provider') THEN
        ALTER TABLE skills ADD COLUMN provider VARCHAR(100);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'skills' AND column_name = 'status') THEN
        ALTER TABLE skills ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'skills' AND indexname = 'idx_skills_code') THEN
        CREATE INDEX idx_skills_code ON skills(code);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'skills' AND indexname = 'idx_skills_category') THEN
        CREATE INDEX idx_skills_category ON skills(category);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'skills' AND indexname = 'idx_skills_status') THEN
        CREATE INDEX idx_skills_status ON skills(status);
    END IF;
END $$;

-- ==========================================
-- 5. skill_installations 表 - 租户技能安装表
-- ==========================================
CREATE TABLE IF NOT EXISTS skill_installations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    config JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'installed',
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'skill_installations_tenant_id_skill_id_key') THEN
        ALTER TABLE skill_installations ADD CONSTRAINT skill_installations_tenant_id_skill_id_key UNIQUE(tenant_id, skill_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'skill_installations' AND indexname = 'idx_skill_installations_tenant_id') THEN
        CREATE INDEX idx_skill_installations_tenant_id ON skill_installations(tenant_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'skill_installations' AND indexname = 'idx_skill_installations_skill_id') THEN
        CREATE INDEX idx_skill_installations_skill_id ON skill_installations(skill_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'skill_installations' AND indexname = 'idx_skill_installations_status') THEN
        CREATE INDEX idx_skill_installations_status ON skill_installations(status);
    END IF;
END $$;

-- ==========================================
-- 6. billing_plans 表 - 计费套餐表
-- ==========================================
CREATE TABLE IF NOT EXISTS billing_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    plan_type VARCHAR(20) NOT NULL DEFAULT 'monthly',
    price DECIMAL(10,2) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    features JSONB DEFAULT '{}',
    token_limit INTEGER NOT NULL DEFAULT 100000,
    max_concurrent INTEGER NOT NULL DEFAULT 10,
    max_agents INTEGER NOT NULL DEFAULT 5,
    max_knowledge_bases INTEGER NOT NULL DEFAULT 10,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'billing_plans' AND indexname = 'idx_billing_plans_plan_type') THEN
        CREATE INDEX idx_billing_plans_plan_type ON billing_plans(plan_type);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'billing_plans' AND indexname = 'idx_billing_plans_is_active') THEN
        CREATE INDEX idx_billing_plans_is_active ON billing_plans(is_active);
    END IF;
END $$;

-- ==========================================
-- 7. subscriptions 表 - 租户订阅表
-- ==========================================
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES billing_plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    start_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_date TIMESTAMPTZ,
    next_billing_date TIMESTAMPTZ,
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMPTZ,
    total_tokens_used INTEGER NOT NULL DEFAULT 0,
    token_limit INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'subscriptions' AND column_name = 'token_limit') THEN
        ALTER TABLE subscriptions ADD COLUMN token_limit INTEGER NOT NULL DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'subscriptions' AND column_name = 'total_tokens_used') THEN
        ALTER TABLE subscriptions ADD COLUMN total_tokens_used INTEGER NOT NULL DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'subscriptions' AND column_name = 'current_period_end') THEN
        ALTER TABLE subscriptions ADD COLUMN current_period_end TIMESTAMPTZ;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'subscriptions_tenant_id_key') THEN
        ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_tenant_id_key UNIQUE(tenant_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'subscriptions' AND indexname = 'idx_subscriptions_plan_id') THEN
        CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'subscriptions' AND indexname = 'idx_subscriptions_status') THEN
        CREATE INDEX idx_subscriptions_status ON subscriptions(status);
    END IF;
END $$;

-- ==========================================
-- 8. billing_records 表 - 计费记录表
-- ==========================================
CREATE TABLE IF NOT EXISTS billing_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id),
    record_type VARCHAR(20) NOT NULL DEFAULT 'usage',
    amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    description VARCHAR(500),
    tokens_used INTEGER NOT NULL DEFAULT 0,
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    model_id UUID REFERENCES ai_models(id),
    provider_id UUID REFERENCES ai_providers(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    transaction_id VARCHAR(200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'billing_records' AND column_name = 'record_type') THEN
        ALTER TABLE billing_records ADD COLUMN record_type VARCHAR(20) NOT NULL DEFAULT 'usage';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'billing_records' AND column_name = 'input_tokens') THEN
        ALTER TABLE billing_records ADD COLUMN input_tokens INTEGER NOT NULL DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'billing_records' AND column_name = 'output_tokens') THEN
        ALTER TABLE billing_records ADD COLUMN output_tokens INTEGER NOT NULL DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'billing_records' AND indexname = 'idx_billing_records_tenant_id') THEN
        CREATE INDEX idx_billing_records_tenant_id ON billing_records(tenant_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'billing_records' AND indexname = 'idx_billing_records_created_at') THEN
        CREATE INDEX idx_billing_records_created_at ON billing_records(created_at DESC);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'billing_records' AND indexname = 'idx_billing_records_record_type') THEN
        CREATE INDEX idx_billing_records_record_type ON billing_records(record_type);
    END IF;
END $$;

-- ==========================================
-- 9. 插入默认技能数据（如果不存在）
-- ==========================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM skills WHERE code = 'weather') THEN
        INSERT INTO skills (name, description, category, icon, code, skill_type, config, schema, provider, version, is_public) VALUES
        ('天气查询', '查询指定城市的天气信息', 'life', '🌤️', 'weather', 'api', '{"api_url": "https://api.openweathermap.org/data/2.5/weather"}', '{"input": {"city": "string"}, "output": {"temperature": "number", "description": "string"}}', 'openweathermap', '1.0.0', true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM skills WHERE code = 'stock') THEN
        INSERT INTO skills (name, description, category, icon, code, skill_type, config, schema, provider, version, is_public) VALUES
        ('股票查询', '查询股票实时行情', 'finance', '📈', 'stock', 'api', '{"api_url": "https://api.financeapi.io/v3/quote"}', '{"input": {"symbol": "string"}, "output": {"price": "number", "change": "number"}}', 'financeapi', '1.0.0', true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM skills WHERE code = 'translate') THEN
        INSERT INTO skills (name, description, category, icon, code, skill_type, config, schema, provider, version, is_public) VALUES
        ('翻译', '多语言翻译', 'utility', '🌍', 'translate', 'api', '{"api_url": "https://api.translate.com/v2"}', '{"input": {"text": "string", "target_lang": "string"}, "output": {"translated_text": "string"}}', 'translate', '1.0.0', true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM skills WHERE code = 'calculator') THEN
        INSERT INTO skills (name, description, category, icon, code, skill_type, config, schema, provider, version, is_public) VALUES
        ('计算器', '数学计算', 'utility', '🧮', 'calculator', 'function', '{}', '{"input": {"expression": "string"}, "output": {"result": "number"}}', 'builtin', '1.0.0', true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM skills WHERE code = 'date_convert') THEN
        INSERT INTO skills (name, description, category, icon, code, skill_type, config, schema, provider, version, is_public) VALUES
        ('日期转换', '日期格式转换', 'utility', '📅', 'date_convert', 'function', '{}', '{"input": {"date": "string", "format": "string"}, "output": {"result": "string"}}', 'builtin', '1.0.0', true);
    END IF;
END $$;

-- ==========================================
-- 10. 插入默认计费套餐（如果不存在）
-- ==========================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM billing_plans WHERE name = '免费版') THEN
        INSERT INTO billing_plans (name, description, plan_type, price, currency, features, token_limit, max_concurrent, max_agents, max_knowledge_bases, is_default, priority) VALUES
        ('免费版', '适合个人开发者和小型团队', 'free', 0, 'CNY', '["基础对话", "5个智能体", "10个知识库"]', 100000, 5, 5, 10, true, 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM billing_plans WHERE name = '标准版') THEN
        INSERT INTO billing_plans (name, description, plan_type, price, currency, features, token_limit, max_concurrent, max_agents, max_knowledge_bases, is_default, priority) VALUES
        ('标准版', '适合中小型企业', 'monthly', 299, 'CNY', '["高级对话", "20个智能体", "50个知识库", "自定义模型", "API Key管理"]', 500000, 20, 20, 50, false, 1);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM billing_plans WHERE name = '专业版') THEN
        INSERT INTO billing_plans (name, description, plan_type, price, currency, features, token_limit, max_concurrent, max_agents, max_knowledge_bases, is_default, priority) VALUES
        ('专业版', '适合大型企业', 'monthly', 999, 'CNY', '["全部功能", "无限智能体", "无限知识库", "私有部署", "技术支持"]', 2000000, 100, 100, 100, false, 2);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM billing_plans WHERE name = '企业版') THEN
        INSERT INTO billing_plans (name, description, plan_type, price, currency, features, token_limit, max_concurrent, max_agents, max_knowledge_bases, is_default, priority) VALUES
        ('企业版', '定制化服务', 'monthly', 2999, 'CNY', '["定制开发", "专属客服", "SLA保障", "私有化部署"]', 10000000, 500, 500, 500, false, 3);
    END IF;
END $$;

-- ==========================================
-- 11. 为默认租户创建订阅（如果不存在）
-- ==========================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM subscriptions WHERE tenant_id = '00000000-0000-0000-0000-000000000001') THEN
        INSERT INTO subscriptions (tenant_id, plan_id, status, token_limit, current_period_end) VALUES
        ('00000000-0000-0000-0000-000000000001', (SELECT id FROM billing_plans WHERE name = '免费版'), 'active', 100000, NOW() + INTERVAL '30 days');
    END IF;
END $$;

COMMIT;