-- Add allowed_channels to plans table for channel gating by plan tier
ALTER TABLE plans ADD COLUMN IF NOT EXISTS allowed_channels JSONB DEFAULT '[]'::jsonb;

-- Set allowed channels per plan tier
UPDATE plans SET allowed_channels = '[]'::jsonb WHERE type = 'free';
UPDATE plans SET allowed_channels = '["wechat", "wecom", "feishu"]'::jsonb WHERE type = 'pro';
UPDATE plans SET allowed_channels = '["wechat", "wecom", "feishu", "dingtalk", "douyin"]'::jsonb WHERE type = 'enterprise';

-- Create schema_migrations tracking table if not exists
CREATE TABLE IF NOT EXISTS schema_migrations (
    id SERIAL PRIMARY KEY,
    version VARCHAR(14) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    installed_on TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO schema_migrations (version, description) VALUES ('V015', 'plan_allowed_channels') ON CONFLICT (version) DO NOTHING;
