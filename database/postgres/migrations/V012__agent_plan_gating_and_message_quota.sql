-- V012: Agent plan gating + message quota enforcement
-- Adds plan-based visibility for agents and sets message quotas per plan tier

-- ============================================
-- 1. Agent cross-tenant sharing (is_public)
-- ============================================
ALTER TABLE agents ADD COLUMN IF NOT EXISTS is_public boolean NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_agents_is_public ON agents(is_public);

-- ============================================
-- 2. Agent required plan tier
-- ============================================
ALTER TABLE agents ADD COLUMN IF NOT EXISTS required_plan varchar(20) NOT NULL DEFAULT 'free';
CREATE INDEX IF NOT EXISTS idx_agents_required_plan ON agents(required_plan);

-- ============================================
-- 3. Assign agents to plan tiers
-- ============================================
-- Free tier: basic customer service and copywriting
UPDATE agents SET required_plan = 'free'
WHERE name IN ('智能客服小助', '文案写手');

-- Pro tier: specialized business agents
UPDATE agents SET required_plan = 'pro'
WHERE name IN (
    '代码助手', '数据分析师', '翻译专家', '会议纪要助手',
    '法律顾问', '财务顾问',
    '营销专家', '销售助手'
);

-- Enterprise tier: advanced operations agents
UPDATE agents SET required_plan = 'enterprise'
WHERE name IN (
    '人事助手', '项目经理',
    '运营管家', '培训导师', '招聘顾问',
    '行政助理', '产品顾问', '售后客服'
);

-- ============================================
-- 4. Plan type normalization (plans table uses 'type' column)
-- ============================================
UPDATE plans SET type = 'free' WHERE name = '免费版';
UPDATE plans SET type = 'pro' WHERE name = '专业版';
UPDATE plans SET type = 'enterprise' WHERE name = '企业版';

-- ============================================
-- 5. Message quota per plan (monthly limits)
-- ============================================
UPDATE plans SET max_messages_per_month = 100 WHERE type = 'free';
UPDATE plans SET max_messages_per_month = 5000 WHERE type = 'pro';
UPDATE plans SET max_messages_per_month = 50000 WHERE type = 'enterprise';
