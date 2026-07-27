-- ==========================================
-- Version: V009
-- Description: 修复 agents.knowledge_base_ids 类型不兼容
-- Date: 2026-07-13
-- 问题：uuid[] 类型在 Go GORM 中无法直接 Scan 到 []uuid.UUID
-- 修复：改为 jsonb 类型，Go 端用 []string + serializer:json
-- ==========================================

BEGIN;

-- 删除旧列（uuid[]类型无法被GORM扫描）
ALTER TABLE agents DROP COLUMN IF EXISTS knowledge_base_ids;

-- 重新添加为 jsonb 类型
ALTER TABLE agents ADD COLUMN knowledge_base_ids jsonb NOT NULL DEFAULT '[]'::jsonb;

COMMIT;
