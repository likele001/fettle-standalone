-- ==========================================
-- Version: V004 回滚脚本
-- Description: 回滚 AI 模型管理系统数据库结构
-- ==========================================

BEGIN;

-- 删除 AI 使用日志表
DROP TABLE IF EXISTS ai_usage_logs;

-- 删除租户 API Key 表
DROP TABLE IF EXISTS tenant_api_keys;

-- 删除租户 AI 配置表
DROP TABLE IF EXISTS tenant_ai_configs;

-- 删除 AI 模型表
DROP TABLE IF EXISTS ai_models;

-- 删除 AI 厂商表
DROP TABLE IF EXISTS ai_providers;

COMMIT;