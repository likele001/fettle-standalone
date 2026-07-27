-- ==========================================
-- Version: V005 Rollback
-- Description: 回滚智能体、知识库、技能、计费系统数据库表
-- Date: 2026-07-08
-- ==========================================

BEGIN;

DROP TABLE IF EXISTS billing_records;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS billing_plans;
DROP TABLE IF EXISTS skill_installations;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS knowledge_documents;
DROP TABLE IF EXISTS knowledge_bases;
DROP TABLE IF EXISTS agents;

COMMIT;