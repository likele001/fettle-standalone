package repository

import (
	"ai-platform/agent-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BundleRepository struct {
	db *gorm.DB
}

func NewBundleRepository(db *gorm.DB) *BundleRepository {
	return &BundleRepository{db: db}
}

func (r *BundleRepository) ListActive() ([]models.IndustryBundle, error) {
	var bundles []models.IndustryBundle
	err := r.db.Where("is_active = ?", true).Order("created_at ASC").Find(&bundles).Error
	return bundles, err
}

func (r *BundleRepository) GetByIndustry(industry string) (*models.IndustryBundle, error) {
	var bundle models.IndustryBundle
	err := r.db.Where("industry = ? AND is_active = ?", industry, true).First(&bundle).Error
	return &bundle, err
}

func (r *BundleRepository) GetByID(id string) (*models.IndustryBundle, error) {
	bID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	var bundle models.IndustryBundle
	err = r.db.Where("id = ?", bID).First(&bundle).Error
	return &bundle, err
}

// CreateAgent creates a new agent for a tenant (used by bundle apply)
func (r *BundleRepository) CreateAgent(agent *models.Agent) error {
	return r.db.Create(agent).Error
}

// CreateKB creates a new knowledge base for a tenant (used by bundle apply)
func (r *BundleRepository) CreateKB(kb *models.KnowledgeBase) error {
	return r.db.Create(kb).Error
}

// CountTenantAgents counts how many agents a tenant has
func (r *BundleRepository) CountTenantAgents(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Agent{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// GetTenantMaxAgents returns the max agents allowed for the tenant's plan
func (r *BundleRepository) GetTenantMaxAgents(tenantID uuid.UUID) int {
	var planType string
	err := r.db.Table("tenants").Select("plan_type").Where("id = ?", tenantID).Scan(&planType).Error
	if err != nil || planType == "" {
		planType = "free"
	}

	var maxAgents int
	err = r.db.Table("plans").Select("max_agents").Where("type = ?", planType).Scan(&maxAgents).Error
	if err != nil || maxAgents <= 0 {
		return 1 // default
	}
	return maxAgents
}
