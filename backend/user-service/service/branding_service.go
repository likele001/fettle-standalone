package service

import (
	"ai-platform/user-service/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BrandingService struct {
	db *gorm.DB
}

func NewBrandingService(db *gorm.DB) *BrandingService {
	return &BrandingService{db: db}
}

// UpdateBrandingRequest 更新品牌配置请求
type UpdateBrandingRequest struct {
	BrandName          *string `json:"brand_name"`
	BrandLogoURL       *string `json:"brand_logo_url"`
	BrandFaviconURL    *string `json:"brand_favicon_url"`
	BrandPrimaryColor  *string `json:"brand_primary_color"`
	BrandSecondaryColor *string `json:"brand_secondary_color"`
	BrandFooterText    *string `json:"brand_footer_text"`
	WhiteLabelEnabled  *bool   `json:"white_label_enabled"`
}

// GetBranding 获取租户品牌配置
func (s *BrandingService) GetBranding(tenantID string) (*models.BrandingConfig, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}

	var tenant models.Tenant
	if err := s.db.First(&tenant, "id = ?", tID).Error; err != nil {
		return nil, errors.New("tenant not found")
	}

	branding := tenant.GetBranding()
	return &branding, nil
}

// UpdateBranding 更新租户品牌配置
func (s *BrandingService) UpdateBranding(tenantID string, req *UpdateBrandingRequest) (*models.BrandingConfig, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}

	var tenant models.Tenant
	if err := s.db.First(&tenant, "id = ?", tID).Error; err != nil {
		return nil, errors.New("tenant not found")
	}

	// Check plan restrictions for white-label
	if req.WhiteLabelEnabled != nil && *req.WhiteLabelEnabled {
		if tenant.PlanType != "enterprise" {
			return nil, errors.New("白标功能仅限企业版套餐使用")
		}
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.BrandName != nil {
		updates["brand_name"] = *req.BrandName
	}
	if req.BrandLogoURL != nil {
		updates["brand_logo_url"] = *req.BrandLogoURL
	}
	if req.BrandFaviconURL != nil {
		updates["brand_favicon_url"] = *req.BrandFaviconURL
	}
	if req.BrandPrimaryColor != nil {
		updates["brand_primary_color"] = *req.BrandPrimaryColor
	}
	if req.BrandSecondaryColor != nil {
		updates["brand_secondary_color"] = *req.BrandSecondaryColor
	}
	if req.BrandFooterText != nil {
		updates["brand_footer_text"] = *req.BrandFooterText
	}
	if req.WhiteLabelEnabled != nil {
		updates["white_label_enabled"] = *req.WhiteLabelEnabled
	}

	if len(updates) > 0 {
		if err := s.db.Model(&tenant).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	// Reload and return
	if err := s.db.First(&tenant, "id = ?", tID).Error; err != nil {
		return nil, err
	}

	branding := tenant.GetBranding()
	return &branding, nil
}

// GetPublicBranding 获取公开的品牌配置（用于前端展示）
func (s *BrandingService) GetPublicBranding(tenantID string) (*models.BrandingConfig, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}

	var tenant models.Tenant
	if err := s.db.First(&tenant, "id = ?", tID).Error; err != nil {
		return nil, errors.New("tenant not found")
	}

	branding := tenant.GetBranding()
	return &branding, nil
}
