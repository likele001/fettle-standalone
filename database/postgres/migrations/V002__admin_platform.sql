-- ==========================================
-- Version: V002
-- Description: 超级管理员后台数据库变更
-- Date: 2026-07-07
-- 迁移命令：
--   psql -U ai_platform -d ai_platform -h 127.0.0.1 \
--     -f database/postgres/migrations/V002__admin_platform.sql
-- 回滚命令：
--   psql -U ai_platform -d ai_platform -h 127.0.0.1 \
--     -f database/postgres/migrations/V002__rollback.sql
-- ==========================================

BEGIN;

-- ==========================================
-- 1. tenants 表 - 增加审核和封禁字段
-- ==========================================
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS audit_status VARCHAR(20) NOT NULL DEFAULT 'pending';
COMMENT ON COLUMN tenants.audit_status IS '审核状态: pending-待审核, approved-已通过, rejected-已驳回';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS audit_remark TEXT;
COMMENT ON COLUMN tenants.audit_remark IS '审核备注';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS audit_at TIMESTAMPTZ;
COMMENT ON COLUMN tenants.audit_at IS '审核时间';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS audit_by UUID;
COMMENT ON COLUMN tenants.audit_by IS '审核人 ID';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS banned_at TIMESTAMPTZ;
COMMENT ON COLUMN tenants.banned_at IS '封禁时间';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS banned_reason TEXT;
COMMENT ON COLUMN tenants.banned_reason IS '封禁原因';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS banned_by UUID;
COMMENT ON COLUMN tenants.banned_by IS '封禁操作人 ID';

-- ==========================================
-- 2. 平台配置表
-- ==========================================
CREATE TABLE IF NOT EXISTS platform_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value JSONB NOT NULL DEFAULT '{}',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID
);

COMMENT ON TABLE platform_config IS '平台全局配置';
COMMENT ON COLUMN platform_config.config_key IS '配置键，如: sms.provider, storage.qiniu, mail.smtp';
COMMENT ON COLUMN platform_config.config_value IS '配置值，JSON 格式';

-- 插入默认配置
INSERT INTO platform_config (config_key, config_value, description) VALUES
('sms.provider', '{"name":"aliyun","access_key":"","access_secret":"","sign_name":""}', '短信服务商配置（阿里云）'),
('storage.default', '{"provider":"local","base_url":"/uploads"}', '默认存储配置'),
('storage.qiniu', '{"bucket":"","access_key":"","secret_key":"","domain":""}', '七牛对象存储配置'),
('storage.tencent', '{"bucket":"","region":"","secret_id":"","secret_key":"","domain":""}', '腾讯云对象存储配置'),
('mail.smtp', '{"host":"","port":465,"user":"","password":"","from":"","from_name":""}', '邮件服务 SMTP 配置'),
('system.registration', '{"allow_register":true,"need_audit":false}', '注册开关配置'),
('system.session', '{"token_expire_hours":24,"refresh_expire_days":7}', '会话过期配置')
ON CONFLICT (config_key) DO NOTHING;

-- ==========================================
-- 3. 管理员操作日志表
-- ==========================================
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id UUID,
    detail JSONB DEFAULT '{}',
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE admin_audit_log IS '管理员操作审计日志';
COMMENT ON COLUMN admin_audit_log.action IS '操作类型: login, audit_tenant, ban_tenant, create_plan, update_plan, update_config, test_sms 等';
COMMENT ON COLUMN admin_audit_log.target_type IS '操作对象类型: tenant, plan, config, user, system';
COMMENT ON COLUMN admin_audit_log.target_id IS '操作对象 ID';

CREATE INDEX IF NOT EXISTS idx_admin_audit_log_admin_id ON admin_audit_log(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_created_at ON admin_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_action ON admin_audit_log(action);

-- ==========================================
-- 4. billings 表 - 增加套餐上下架字段
-- ==========================================
ALTER TABLE plans ADD COLUMN IF NOT EXISTS is_published BOOLEAN NOT NULL DEFAULT true;
COMMENT ON COLUMN plans.is_published IS '是否上架: true-上架, false-下架';

ALTER TABLE plans ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;
COMMENT ON COLUMN plans.published_at IS '上架时间';

COMMIT;
