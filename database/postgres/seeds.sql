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
--    ⚠️ 口令不在此处明文定义，本段已默认关闭。
--    首次启动时由 user-service 的 AutoMigrate 播种：
--      · 读过环境变量 SUPER_ADMIN_PASSWORD 则用它；
--      · 未设置则生成随机口令，并在日志中只打印一次。
--    如需用本 SQL 手工初始化，请先自行生成 bcrypt 哈希再放开下面的语句。
-- ============================================
-- INSERT INTO users (id, tenant_id, phone, email, password_hash, name, role, status)
-- VALUES (
--     '00000000-0000-0000-0000-000000000010',
--     '00000000-0000-0000-0000-000000000001',
--     '13800000000',
--     'admin@fettle.com',
--     '<在此填入自建的 bcrypt 哈希>',
--     '超级管理员',
--     'super_admin',
--     'active'
-- ) ON CONFLICT (id) DO NOTHING;

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
