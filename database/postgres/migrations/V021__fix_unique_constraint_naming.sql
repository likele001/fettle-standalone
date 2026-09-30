-- ============================================================
-- V021: 修复 standalone 库唯一约束命名（GORM 兼容）
--
-- 背景：GORM AutoMigrate 期望 uni_<table>_<cols>，
--       PostgreSQL 默认生成 <table>_<cols>_key。
--       命名不一致导致每次启动都报
--       'constraint "uni_xxx" does not exist' 并中断迁移。
--
-- 全部语句幂等，可重复执行。
-- ============================================================

BEGIN;

-- schema_migrations_version_key -> uni_schema_migrations_version
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint
               WHERE conname = 'schema_migrations_version_key') THEN
        ALTER TABLE schema_migrations
            RENAME CONSTRAINT schema_migrations_version_key
            TO uni_schema_migrations_version;
    END IF;
END $$;

-- skill_installations_tenant_id_skill_id_key -> uni_skill_installations_tenant_id_skill_id
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint
               WHERE conname = 'skill_installations_tenant_id_skill_id_key') THEN
        ALTER TABLE skill_installations
            RENAME CONSTRAINT skill_installations_tenant_id_skill_id_key
            TO uni_skill_installations_tenant_id_skill_id;
    END IF;
END $$;

-- subscriptions_tenant_id_key -> uni_subscriptions_tenant_id
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint
               WHERE conname = 'subscriptions_tenant_id_key') THEN
        ALTER TABLE subscriptions
            RENAME CONSTRAINT subscriptions_tenant_id_key
            TO uni_subscriptions_tenant_id;
    END IF;
END $$;

INSERT INTO schema_migrations (version, description)
VALUES ('V021', 'fix_unique_constraint_naming_for_gorm')
ON CONFLICT (version) DO NOTHING;

COMMIT;
