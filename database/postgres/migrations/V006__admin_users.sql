-- ==========================================
-- Version: V006
-- Description: 独立超级管理员用户系统
-- Date: 2026-07-08
-- 功能说明：
--   1. 创建 admin_users 表，独立于租户用户（users 表）
--   2. 超级管理员账号存储在 admin_users 表
--   3. 租户用户在 users 表，两者完全分离
-- ==========================================

BEGIN;

-- ==========================================
-- 1. admin_users 表 - 超级管理员用户
-- ==========================================
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(100),
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(50),
    avatar_url TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'super_admin',
    status VARCHAR(10) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

COMMENT ON TABLE admin_users IS '超级管理员用户表（独立于租户用户）';
COMMENT ON COLUMN admin_users.role IS '角色：super_admin';

CREATE INDEX IF NOT EXISTS idx_admin_users_phone ON admin_users(phone);
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
CREATE INDEX IF NOT EXISTS idx_admin_users_status ON admin_users(status);

-- ==========================================
-- 2. 迁移现有超级管理员到 admin_users 表
-- ==========================================
-- 如果 admin_users 表为空，从 users 表迁移超级管理员
INSERT INTO admin_users (phone, email, password_hash, name, role, status)
SELECT phone, email, password_hash, name, role, status
FROM users
WHERE role = 'super_admin'
AND NOT EXISTS (SELECT 1 FROM admin_users WHERE phone = users.phone);

-- ==========================================
-- 3. 清理 users 表中的超级管理员（可选）
-- ==========================================
-- 注意：这里不删除 users 表中的超级管理员，保持数据完整性
-- 如果需要清理，可以手动执行：
-- DELETE FROM users WHERE role = 'super_admin' AND tenant_id = '00000000-0000-0000-0000-000000000001';

COMMIT;
