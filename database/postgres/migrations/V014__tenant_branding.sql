-- V014: Tenant branding / white-label support
-- Adds branding fields to tenants table for custom logo, name, colors

-- ============================================
-- 1. Add branding columns to tenants
-- ============================================
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS brand_name VARCHAR(100);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS brand_logo_url VARCHAR(500);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS brand_favicon_url VARCHAR(500);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS brand_primary_color VARCHAR(20);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS brand_secondary_color VARCHAR(20);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS brand_footer_text VARCHAR(200);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS white_label_enabled BOOLEAN NOT NULL DEFAULT false;

-- ============================================
-- 2. Index for branding lookups
-- ============================================
CREATE INDEX IF NOT EXISTS idx_tenants_white_label ON tenants(white_label_enabled);
