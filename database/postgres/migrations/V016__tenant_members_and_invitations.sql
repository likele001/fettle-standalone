-- V016: 租户成员与邀请码表

-- 租户成员表（一个用户只能属于一个租户）
CREATE TABLE IF NOT EXISTS tenant_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',  -- owner / admin / member
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active / resigned
    invited_by UUID REFERENCES users(id),
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resigned_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_members_tenant_id ON tenant_members(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_user_id ON tenant_members(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_status ON tenant_members(status);

-- 邀请码表
CREATE TABLE IF NOT EXISTS tenant_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL UNIQUE,
    invitee_phone VARCHAR(20),
    invitee_email VARCHAR(100),
    role VARCHAR(20) NOT NULL DEFAULT 'member',  -- admin / member
    invited_by UUID NOT NULL REFERENCES users(id),
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_invitations_code ON tenant_invitations(code);
CREATE INDEX IF NOT EXISTS idx_tenant_invitations_tenant_id ON tenant_invitations(tenant_id);

-- 迁移现有数据：每个用户自动成为其租户的 owner
INSERT INTO tenant_members (tenant_id, user_id, role, status, joined_at)
SELECT tenant_id, id, 'owner', 'active', created_at
FROM users
WHERE deleted_at IS NULL
ON CONFLICT (user_id) DO NOTHING;

-- 记录迁移
INSERT INTO schema_migrations (version, description) VALUES ('V016', 'tenant_members_and_invitations') ON CONFLICT (version) DO NOTHING;
