package service

import (
	"encoding/json"
	"errors"
	"time"

	"ai-platform/user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantService 租户服务
type TenantService struct {
	db *gorm.DB
}

// NewTenantService 创建租户服务
func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{db: db}
}

// GetByID 根据ID获取租户
func (s *TenantService) GetByID(tenantID string) (*models.Tenant, error) {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}

	var tenant models.Tenant
	if err := s.db.Where("id = ?", id).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name   string  `json:"name"`
	Code   *string `json:"code"`
	Config *string `json:"config"`
}

// Update 更新租户
func (s *TenantService) Update(tenantID string, req *UpdateTenantRequest) (*models.Tenant, error) {
	tenant, err := s.GetByID(tenantID)
	if err != nil {
		return nil, err
	}

	if req.Code != nil {
		tenant.Code = req.Code
	}
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Config != nil {
		var configMap models.JSONMap
		if err := json.Unmarshal([]byte(*req.Config), &configMap); err == nil {
			tenant.Config = configMap
		}
	}

	if err := s.db.Save(tenant).Error; err != nil {
		return nil, err
	}

	return tenant, nil
}

// ListTenantsRequest 租户列表请求
type ListTenantsRequest struct {
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
	Search      string `json:"search"`
	Status      string `json:"status"`
	AuditStatus string `json:"audit_status"`
}

// ListTenantsResponse 租户列表响应
type ListTenantsResponse struct {
	Items []models.Tenant `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
}

// ListTenants 管理员：获取所有租户列表
func (s *TenantService) ListTenants(req *ListTenantsRequest) (*ListTenantsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.Tenant{})

	if req.Search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.AuditStatus != "" {
		query = query.Where("audit_status = ?", req.AuditStatus)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var tenants []models.Tenant
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&tenants).Error; err != nil {
		return nil, err
	}

	return &ListTenantsResponse{
		Items: tenants,
		Total: total,
		Page:  req.Page,
	}, nil
}

// AuditTenantRequest 审核请求
type AuditTenantRequest struct {
	Status string `json:"status"` // approved / rejected
	Remark string `json:"remark"`
}

// AuditTenant 管理员：审核租户
func (s *TenantService) AuditTenant(tenantID string, adminID string, req *AuditTenantRequest) (*models.Tenant, error) {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}

	var tenant models.Tenant
	if err := s.db.Where("id = ?", id).First(&tenant).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	tenant.AuditStatus = &req.Status
	tenant.AuditAt = &now
	if req.Remark != "" {
		tenant.AuditRemark = &req.Remark
	}
	adminUUID, _ := uuid.Parse(adminID)
	tenant.AuditBy = &adminUUID

	if req.Status == "approved" {
		tenant.Status = "active"
	}

	if err := s.db.Save(&tenant).Error; err != nil {
		return nil, err
	}

	return &tenant, nil
}

// BanTenantRequest 封禁请求
type BanTenantRequest struct {
	Banned bool   `json:"banned"`
	Reason string `json:"reason"`
}

// BanTenant 管理员：封禁/解封租户
func (s *TenantService) BanTenant(tenantID string, adminID string, req *BanTenantRequest) (*models.Tenant, error) {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}

	var tenant models.Tenant
	if err := s.db.Where("id = ?", id).First(&tenant).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	adminUUID, _ := uuid.Parse(adminID)

	if req.Banned {
		tenant.Status = "banned"
		tenant.BannedAt = &now
		tenant.BannedBy = &adminUUID
		if req.Reason != "" {
			tenant.BannedReason = &req.Reason
		}
	} else {
		tenant.Status = "active"
		tenant.BannedAt = nil
		tenant.BannedReason = nil
		tenant.BannedBy = nil
	}

	if err := s.db.Save(&tenant).Error; err != nil {
		return nil, err
	}

	return &tenant, nil
}
