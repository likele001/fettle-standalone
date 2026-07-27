-- 中小企业 AI 智能体平台 - 初始数据

-- ============================================
-- 1. 平台租户（超管所属的特殊租户）
-- ============================================
INSERT INTO tenants (id, name, plan_type, status, config)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '平台管理',
    'enterprise',
    'active',
    '{"is_platform": true}'
) ON CONFLICT (id) DO NOTHING;

-- ============================================
-- 2. 超级管理员账号
--    用户名: superadmin
--    密码:   Admin@2026
--    (bcrypt hash of 'Admin@2026')
-- ============================================
INSERT INTO users (id, tenant_id, phone, email, password_hash, name, role, status)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000001',
    '13800000000',
    'admin@fettle.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7wVjB5T8h1Kd3O6E2XpF7Wm',
    '超级管理员',
    'super_admin',
    'active'
) ON CONFLICT (id) DO NOTHING;

-- ============================================
-- 3. 预置权限
-- ============================================
INSERT INTO permissions (code, name, module) VALUES
    ('user:read', '查看用户', 'user'),
    ('user:write', '管理用户', 'user'),
    ('agent:read', '查看智能体', 'agent'),
    ('agent:write', '管理智能体', 'agent'),
    ('chat:read', '查看会话', 'chat'),
    ('chat:write', '管理会话', 'chat'),
    ('skill:read', '查看技能', 'skill'),
    ('skill:write', '管理技能', 'skill'),
    ('billing:read', '查看账单', 'billing'),
    ('billing:write', '管理账单', 'billing'),
    ('tenant:read', '查看租户', 'tenant'),
    ('tenant:write', '管理租户', 'tenant'),
    ('plan:read', '查看套餐', 'plan'),
    ('plan:write', '管理套餐', 'plan'),
    ('system:read', '查看系统配置', 'system'),
    ('system:write', '管理系统配置', 'system'),
    ('admin:read', '查看管理员', 'admin'),
    ('admin:write', '管理管理员', 'admin')
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- 4. 预置套餐信息
--    free: 免费版（100次/天API调用，1个智能体，基础技能）
--    pro: 专业版（10000次/天API调用，10个智能体，全部技能，GPT-4o可用）
--    enterprise: 企业版（无限API调用，无限智能体，私有化部署，专属支持）
-- ============================================
