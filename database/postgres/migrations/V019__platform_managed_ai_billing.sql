-- V019: 平台托管AI - 余额、资源包、平台定价

-- 平台模型定价表（原价转售，不抽成）
CREATE TABLE IF NOT EXISTS platform_model_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES ai_models(id),
    input_price_per_1k NUMERIC(10,6) NOT NULL DEFAULT 0,   -- 输入价格 元/千token
    output_price_per_1k NUMERIC(10,6) NOT NULL DEFAULT 0,  -- 输出价格 元/千token
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(model_id)
);
COMMENT ON TABLE platform_model_pricing IS '平台模型定价（原价转售）';
COMMENT ON COLUMN platform_model_pricing.model_id IS '关联的AI模型ID';
COMMENT ON COLUMN platform_model_pricing.input_price_per_1k IS '输入价格（元/千token）';
COMMENT ON COLUMN platform_model_pricing.output_price_per_1k IS '输出价格（元/千token）';
COMMENT ON COLUMN platform_model_pricing.is_enabled IS '是否启用';

-- 租户余额表
CREATE TABLE IF NOT EXISTS tenant_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id),
    balance NUMERIC(12,2) NOT NULL DEFAULT 0,          -- 可用余额（元）
    total_recharged NUMERIC(12,2) NOT NULL DEFAULT 0,  -- 累计充值
    total_consumed NUMERIC(12,2) NOT NULL DEFAULT 0,   -- 累计消费
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE tenant_balances IS '租户余额表';
COMMENT ON COLUMN tenant_balances.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_balances.balance IS '可用余额（元）';
COMMENT ON COLUMN tenant_balances.total_recharged IS '累计充值（元）';
COMMENT ON COLUMN tenant_balances.total_consumed IS '累计消费（元）';

-- 资源包定义表
CREATE TABLE IF NOT EXISTS resource_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,          -- 资源包名称
    token_amount BIGINT NOT NULL,        -- 包含token数
    price NUMERIC(10,2) NOT NULL,        -- 价格（元）
    description TEXT,                    -- 描述
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE resource_packages IS '资源包定义表';
COMMENT ON COLUMN resource_packages.name IS '资源包名称';
COMMENT ON COLUMN resource_packages.token_amount IS '包含token数';
COMMENT ON COLUMN resource_packages.price IS '价格（元）';

-- 租户资源包购买记录
CREATE TABLE IF NOT EXISTS tenant_resource_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    package_id UUID NOT NULL REFERENCES resource_packages(id),
    token_amount BIGINT NOT NULL,        -- 包含token数
    token_used BIGINT NOT NULL DEFAULT 0, -- 已用token数
    token_remaining BIGINT NOT NULL DEFAULT 0, -- 剩余token数
    price NUMERIC(10,2) NOT NULL,        -- 购买价格
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active/expired/exhausted
    expires_at TIMESTAMP WITH TIME ZONE, -- 过期时间（NULL=永不过期）
    purchased_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE tenant_resource_packages IS '租户资源包购买记录';
COMMENT ON COLUMN tenant_resource_packages.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_resource_packages.package_id IS '资源包ID';
COMMENT ON COLUMN tenant_resource_packages.token_amount IS '包含token数';
COMMENT ON COLUMN tenant_resource_packages.token_used IS '已用token数';
COMMENT ON COLUMN tenant_resource_packages.token_remaining IS '剩余token数';
COMMENT ON COLUMN tenant_resource_packages.price IS '购买价格（元）';
COMMENT ON COLUMN tenant_resource_packages.status IS '状态: active/expired/exhausted';
COMMENT ON COLUMN tenant_resource_packages.expires_at IS '过期时间';

-- 租户AI调用计费明细
CREATE TABLE IF NOT EXISTS tenant_usage_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID,
    conversation_id UUID,
    message_id UUID,
    model_id UUID,
    provider_id UUID,
    input_tokens INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    total_tokens INT NOT NULL DEFAULT 0,
    cost NUMERIC(10,4) NOT NULL DEFAULT 0,       -- 本次费用（元）
    billing_mode VARCHAR(20) NOT NULL DEFAULT 'balance', -- balance/package
    package_id UUID,                              -- 如果用了资源包
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE tenant_usage_details IS '租户AI调用计费明细';
COMMENT ON COLUMN tenant_usage_details.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_usage_details.model_id IS '模型ID';
COMMENT ON COLUMN tenant_usage_details.input_tokens IS '输入token数';
COMMENT ON COLUMN tenant_usage_details.output_tokens IS '输出token数';
COMMENT ON COLUMN tenant_usage_details.total_tokens IS '总token数';
COMMENT ON COLUMN tenant_usage_details.cost IS '本次费用（元）';
COMMENT ON COLUMN tenant_usage_details.billing_mode IS '计费模式: balance-余额, package-资源包';
COMMENT ON COLUMN tenant_usage_details.package_id IS '使用的资源包ID';

-- 充值记录表
CREATE TABLE IF NOT EXISTS tenant_recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    amount NUMERIC(10,2) NOT NULL,           -- 充值金额
    payment_method VARCHAR(20) NOT NULL DEFAULT 'manual', -- manual/admin/online
    trade_order_id VARCHAR(200),             -- 支付平台订单号
    status VARCHAR(20) NOT NULL DEFAULT 'completed', -- completed/pending/failed
    remark TEXT,                             -- 备注
    operated_by UUID,                        -- 操作人（管理员充值时）
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE tenant_recharge_records IS '租户充值记录';
COMMENT ON COLUMN tenant_recharge_records.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_recharge_records.amount IS '充值金额（元）';
COMMENT ON COLUMN tenant_recharge_records.payment_method IS '支付方式: manual-手动/admin-管理员/online-在线';
COMMENT ON COLUMN tenant_recharge_records.status IS '状态: completed/pending/failed';

-- 初始化：从现有 ai_models 同步定价（用官方价格）
INSERT INTO platform_model_pricing (model_id, input_price_per_1k, output_price_per_1k)
SELECT id, input_price_per1k, output_price_per1k
FROM ai_models
WHERE status = 'active'
ON CONFLICT (model_id) DO NOTHING;

-- 初始化资源包
INSERT INTO resource_packages (name, token_amount, price, description, sort_order) VALUES
('体验包', 100000, 0.50, '10万token，适合体验', 0),
('基础包', 1000000, 5.00, '100万token，适合小规模使用', 1),
('标准包', 5000000, 25.00, '500万token，适合中等规模', 2),
('专业包', 20000000, 100.00, '2000万token，适合大规模使用', 3),
('企业包', 100000000, 500.00, '1亿token，适合企业级使用', 4)
ON CONFLICT DO NOTHING;

-- 记录迁移
INSERT INTO schema_migrations (version, description) VALUES ('V019', 'platform_managed_ai_billing') ON CONFLICT (version) DO NOTHING;
