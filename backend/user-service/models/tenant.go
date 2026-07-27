package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tenant 租户模型
type Tenant struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Code          *string        `gorm:"size:50;uniqueIndex" json:"code"`
	PlanType      string         `gorm:"size:20;not null;default:'free'" json:"plan_type"`
	PlanExpiresAt *time.Time     `json:"plan_expires_at"`
	Config        JSONMap        `gorm:"type:jsonb;default:'{}'" json:"config"`
	Status        string         `gorm:"size:20;default:'active'" json:"status"`
	// 审核相关
	AuditStatus *string    `gorm:"size:20;default:'pending'" json:"audit_status"`
	AuditRemark *string    `gorm:"type:text" json:"audit_remark"`
	AuditAt     *time.Time `json:"audit_at"`
	AuditBy     *uuid.UUID `gorm:"type:uuid" json:"audit_by"`
	// 封禁相关
	BannedAt     *time.Time `json:"banned_at"`
	BannedReason *string    `gorm:"type:text" json:"banned_reason"`
	BannedBy     *uuid.UUID `gorm:"type:uuid" json:"banned_by"`
	// 品牌定制 / 白标
	BrandName          *string `gorm:"size:100" json:"brand_name"`
	BrandLogoURL       *string `gorm:"size:500" json:"brand_logo_url"`
	BrandFaviconURL    *string `gorm:"size:500" json:"brand_favicon_url"`
	BrandPrimaryColor  *string `gorm:"size:20" json:"brand_primary_color"`
	BrandSecondaryColor *string `gorm:"size:20" json:"brand_secondary_color"`
	BrandFooterText    *string `gorm:"size:200" json:"brand_footer_text"`
	WhiteLabelEnabled  bool    `gorm:"default:false" json:"white_label_enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Tenant) TableName() string {
	return "tenants"
}

// BrandingConfig 品牌配置（用于 API 响应）
type BrandingConfig struct {
	BrandName          string `json:"brand_name"`
	BrandLogoURL       string `json:"brand_logo_url"`
	BrandFaviconURL    string `json:"brand_favicon_url"`
	BrandPrimaryColor  string `json:"brand_primary_color"`
	BrandSecondaryColor string `json:"brand_secondary_color"`
	BrandFooterText    string `json:"brand_footer_text"`
	WhiteLabelEnabled  bool   `json:"white_label_enabled"`
}

// GetBranding 获取租户品牌配置
func (t *Tenant) GetBranding() BrandingConfig {
	b := BrandingConfig{
		WhiteLabelEnabled: t.WhiteLabelEnabled,
	}
	if t.BrandName != nil {
		b.BrandName = *t.BrandName
	}
	if t.BrandLogoURL != nil {
		b.BrandLogoURL = *t.BrandLogoURL
	}
	if t.BrandFaviconURL != nil {
		b.BrandFaviconURL = *t.BrandFaviconURL
	}
	if t.BrandPrimaryColor != nil {
		b.BrandPrimaryColor = *t.BrandPrimaryColor
	}
	if t.BrandSecondaryColor != nil {
		b.BrandSecondaryColor = *t.BrandSecondaryColor
	}
	if t.BrandFooterText != nil {
		b.BrandFooterText = *t.BrandFooterText
	}
	return b
}
