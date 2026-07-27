-- 中小企业 AI 智能体平台 - PostgreSQL 初始化脚本
-- 在容器首次启动时由 docker-entrypoint-initdb.d 执行

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 确保默认时区为 UTC（业务层负责时区转换）
SET TIME ZONE 'UTC';

-- 创建基础 schema（GORM AutoMigrate 会创建具体表）
CREATE SCHEMA IF NOT EXISTS public;

-- 初始化系统租户（可选，正式环境建议通过管理后台创建）
-- INSERT INTO tenants (id, name, plan_type, status, created_at, updated_at)
-- VALUES (gen_random_uuid(), '系统默认租户', 'free', 'active', NOW(), NOW())
-- ON CONFLICT DO NOTHING;

-- 记录初始化完成
SELECT 'PostgreSQL initialization completed at ' || NOW() AS message;
