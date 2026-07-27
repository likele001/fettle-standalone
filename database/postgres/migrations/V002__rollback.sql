-- ==========================================
-- 回滚 V002：超级管理员后台数据库变更
-- 命令：
--   psql -U ai_platform -d ai_platform -h 127.0.0.1 \
--     -f database/postgres/migrations/V002__rollback.sql
-- ==========================================

BEGIN;

-- 1. 移除 tenants 新增字段
ALTER TABLE tenants DROP COLUMN IF EXISTS audit_status;
ALTER TABLE tenants DROP COLUMN IF EXISTS audit_remark;
ALTER TABLE tenants DROP COLUMN IF EXISTS audit_at;
ALTER TABLE tenants DROP COLUMN IF EXISTS audit_by;
ALTER TABLE tenants DROP COLUMN IF EXISTS banned_at;
ALTER TABLE tenants DROP COLUMN IF EXISTS banned_reason;
ALTER TABLE tenants DROP COLUMN IF EXISTS banned_by;

-- 2. 删除平台配置表
DROP TABLE IF EXISTS platform_config;

-- 3. 删除操作日志表
DROP TABLE IF EXISTS admin_audit_log;

-- 4. 移除 plans 新增字段
ALTER TABLE plans DROP COLUMN IF EXISTS is_published;
ALTER TABLE plans DROP COLUMN IF EXISTS published_at;

COMMIT;
