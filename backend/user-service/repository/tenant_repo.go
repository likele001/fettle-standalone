package repository

import (
	"errors"

	"ai-platform/user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantRepository 租户数据访问层
type TenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository 创建租户数据访问层
func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// GetByID 根据ID获取租户
func (r *TenantRepository) GetByID(tenantID string) (*models.Tenant, error) {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}

	var tenant models.Tenant
	if err := r.db.Where("id = ?", id).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}
	return &tenant, nil
}

// Create 创建租户
func (r *TenantRepository) Create(tenant *models.Tenant) error {
	return r.db.Create(tenant).Error
}

// Update 更新租户
func (r *TenantRepository) Update(tenant *models.Tenant) error {
	return r.db.Save(tenant).Error
}

// Delete 删除租户（软删除）
func (r *TenantRepository) Delete(tenantID string) error {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return errors.New("invalid tenant id")
	}

	return r.db.Where("id = ?", id).Delete(&models.Tenant{}).Error
}

// List 获取租户列表
func (r *TenantRepository) List(page, pageSize int) ([]models.Tenant, int64, error) {
	var tenants []models.Tenant
	var total int64

	// Count
	if err := r.db.Model(&models.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Find
	if err := r.db.Model(&models.Tenant{}).Offset((page - 1) * pageSize).Limit(pageSize).Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

// GetByStatus 根据状态获取租户
func (r *TenantRepository) GetByStatus(status string) ([]models.Tenant, error) {
	var tenants []models.Tenant
	if err := r.db.Where("status = ?", status).Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

// UpdateConfig 更新租户配置
func (r *TenantRepository) UpdateConfig(tenantID string, config models.JSONMap) error {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return errors.New("invalid tenant id")
	}

	return r.db.Model(&models.Tenant{}).Where("id = ?", id).Update("config", config).Error
}

// UpdatePlanType 更新租户套餐
func (r *TenantRepository) UpdatePlanType(tenantID, planType string) error {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return errors.New("invalid tenant id")
	}

	return r.db.Model(&models.Tenant{}).Where("id = ?", id).Update("plan_type", planType).Error
}
