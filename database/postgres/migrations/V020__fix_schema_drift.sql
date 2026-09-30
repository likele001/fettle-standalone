-- ============================================================
-- V020: 补齐 standalone 库缺失列（schema drift 修复）
-- 生成方式：从 ai_platform 库 information_schema 逐列对齐定义
-- 生成时间：2026-09-29
--
-- 背景：fettle-standalone 从未执行完整迁移链，导致 6 张表
--       缺少 11 个列，前端/后端访问这些字段时报错或静默失效。
--
-- 全部语句幂等（IF NOT EXISTS），可安全重复执行。
-- ============================================================

BEGIN;

-- ---------- billing_plans (2 列) ----------
ALTER TABLE billing_plans
    ADD COLUMN IF NOT EXISTS max_documents_per_kb BIGINT NOT NULL DEFAULT 10;
ALTER TABLE billing_plans
    ADD COLUMN IF NOT EXISTS price_yearly NUMERIC(10,2) NOT NULL DEFAULT 0;

-- ---------- channels (1 列) ----------
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS webhook_secret VARCHAR(200);

-- ---------- subscriptions (3 列) ----------
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS auto_renew BOOLEAN DEFAULT false;
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS messages_used BIGINT DEFAULT 0;

-- ---------- billing_records (3 列) ----------
-- cost / type 在主库为 NOT NULL 且无默认值：先给默认值补列，再放开约束
ALTER TABLE billing_records
    ADD COLUMN IF NOT EXISTS cost NUMERIC(10,4) NOT NULL DEFAULT 0;
ALTER TABLE billing_records
    ADD COLUMN IF NOT EXISTS record_date TIMESTAMPTZ;
ALTER TABLE billing_records
    ADD COLUMN IF NOT EXISTS type VARCHAR(20) NOT NULL DEFAULT 'usage';

-- 如果主库无默认值，这里去掉默认（保持与主库一致，仅当表已有数据时不报错）
ALTER TABLE billing_records ALTER COLUMN cost DROP DEFAULT;

-- ---------- platform_model_pricing (2 列) ----------
ALTER TABLE platform_model_pricing
    ADD COLUMN IF NOT EXISTS input_price_per_1k NUMERIC(10,6) NOT NULL DEFAULT 0;
ALTER TABLE platform_model_pricing
    ADD COLUMN IF NOT EXISTS output_price_per_1k NUMERIC(10,6) NOT NULL DEFAULT 0;

-- ---------- 补齐 billing_records.record_date 数据 ----------
UPDATE billing_records SET record_date = created_at WHERE record_date IS NULL;

-- ---------- 记录迁移 ----------
INSERT INTO schema_migrations (version, description)
VALUES ('V020', 'fix_schema_drift_missing_columns')
ON CONFLICT (version) DO NOTHING;

COMMIT;
