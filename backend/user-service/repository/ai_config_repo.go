package repository

import (
	"ai-platform/user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AIProviderRepository AI 厂商仓储
type AIProviderRepository struct {
	db *gorm.DB
}

func NewAIProviderRepository(db *gorm.DB) *AIProviderRepository {
	return &AIProviderRepository{db: db}
}

func (r *AIProviderRepository) List(status string) ([]models.AIProvider, error) {
	var providers []models.AIProvider
	query := r.db.Order("is_domestic DESC, code ASC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&providers).Error
	return providers, err
}

func (r *AIProviderRepository) GetByID(id uuid.UUID) (*models.AIProvider, error) {
	var provider models.AIProvider
	err := r.db.First(&provider, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *AIProviderRepository) GetByCode(code string) (*models.AIProvider, error) {
	var provider models.AIProvider
	err := r.db.First(&provider, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *AIProviderRepository) Create(provider *models.AIProvider) error {
	return r.db.Create(provider).Error
}

func (r *AIProviderRepository) Update(provider *models.AIProvider) error {
	return r.db.Save(provider).Error
}

func (r *AIProviderRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.AIProvider{}, "id = ?", id).Error
}

// AIModelRepository AI 模型仓储
type AIModelRepository struct {
	db *gorm.DB
}

func NewAIModelRepository(db *gorm.DB) *AIModelRepository {
	return &AIModelRepository{db: db}
}

func (r *AIModelRepository) List(providerID *uuid.UUID, modelType string, status string) ([]models.AIModel, error) {
	var modelsList []models.AIModel
	query := r.db.Preload("Provider").Order("priority ASC, created_at ASC")
	
	if providerID != nil {
		query = query.Where("provider_id = ?", providerID)
	}
	if modelType != "" {
		query = query.Where("model_type = ?", modelType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Find(&modelsList).Error
	return modelsList, err
}

func (r *AIModelRepository) GetByID(id uuid.UUID) (*models.AIModel, error) {
	var model models.AIModel
	err := r.db.Preload("Provider").First(&model, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *AIModelRepository) GetByCode(providerID uuid.UUID, modelCode string) (*models.AIModel, error) {
	var model models.AIModel
	err := r.db.First(&model, "provider_id = ? AND model_code = ?", providerID, modelCode).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *AIModelRepository) GetDefaultModel(providerID uuid.UUID, modelType string) (*models.AIModel, error) {
	var model models.AIModel
	err := r.db.Where("provider_id = ? AND model_type = ? AND is_default = true AND status = 'active'", providerID, modelType).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *AIModelRepository) Create(model *models.AIModel) error {
	return r.db.Create(model).Error
}

func (r *AIModelRepository) Update(model *models.AIModel) error {
	return r.db.Save(model).Error
}

func (r *AIModelRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.AIModel{}, "id = ?", id).Error
}

// TenantAIConfigRepository 租户 AI 配置仓储
type TenantAIConfigRepository struct {
	db *gorm.DB
}

func NewTenantAIConfigRepository(db *gorm.DB) *TenantAIConfigRepository {
	return &TenantAIConfigRepository{db: db}
}

func (r *TenantAIConfigRepository) GetByTenantID(tenantID uuid.UUID) (*models.TenantAIConfig, error) {
	var config models.TenantAIConfig
	err := r.db.Preload("DefaultChatModel").Preload("DefaultChatModel.Provider").
		Preload("DefaultEmbeddingModel").Preload("DefaultEmbeddingModel.Provider").
		Preload("DefaultProvider").
		First(&config, "tenant_id = ?", tenantID).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *TenantAIConfigRepository) Create(config *models.TenantAIConfig) error {
	return r.db.Create(config).Error
}

func (r *TenantAIConfigRepository) Update(config *models.TenantAIConfig) error {
	return r.db.Save(config).Error
}

func (r *TenantAIConfigRepository) Upsert(config *models.TenantAIConfig) error {
	// 以 tenant_id 为唯一键做 upsert：PUT 请求体通常不含主键 id。
	// 若直接 Save，主键为空会被 GORM 当作新记录执行 INSERT，触发
	// uni_tenant_ai_configs_tenant_id 唯一约束冲突（duplicate key）。
	var existing models.TenantAIConfig
	err := r.db.First(&existing, "tenant_id = ?", config.TenantID).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(config).Error
	}
	if err != nil {
		return err
	}
	// 更新时保留系统字段，避免 Save 把 CreatedAt / 用量计数清零
	config.ID = existing.ID
	config.CreatedAt = existing.CreatedAt
	config.MonthlyTokenUsed = existing.MonthlyTokenUsed
	config.TokenLimitResetAt = existing.TokenLimitResetAt
	return r.db.Save(config).Error
}

// TenantAPIKeyRepository 租户 API Key 仓储
type TenantAPIKeyRepository struct {
	db *gorm.DB
}

func NewTenantAPIKeyRepository(db *gorm.DB) *TenantAPIKeyRepository {
	return &TenantAPIKeyRepository{db: db}
}

func (r *TenantAPIKeyRepository) List(tenantID uuid.UUID) ([]models.TenantAPIKey, error) {
	var keys []models.TenantAPIKey
	err := r.db.Table("tenant_api_keys k").
		Select("k.*, p.name as provider_name").
		Joins("LEFT JOIN ai_providers p ON p.id = k.provider_id").
		Where("k.tenant_id = ?", tenantID).
		Order("k.created_at DESC").
		Scan(&keys).Error
	return keys, err
}

func (r *TenantAPIKeyRepository) GetByID(id uuid.UUID) (*models.TenantAPIKey, error) {
	var key models.TenantAPIKey
	err := r.db.Preload("Provider").First(&key, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *TenantAPIKeyRepository) GetByTenantProvider(tenantID, providerID uuid.UUID) (*models.TenantAPIKey, error) {
	var key models.TenantAPIKey
	err := r.db.Where("tenant_id = ? AND provider_id = ? AND status = 'active'", tenantID, providerID).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *TenantAPIKeyRepository) Create(key *models.TenantAPIKey) error {
	return r.db.Create(key).Error
}

func (r *TenantAPIKeyRepository) Update(key *models.TenantAPIKey) error {
	return r.db.Save(key).Error
}

func (r *TenantAPIKeyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.TenantAPIKey{}, "id = ?", id).Error
}

// AIUsageLogRepository AI 使用日志仓储
type AIUsageLogRepository struct {
	db *gorm.DB
}

func NewAIUsageLogRepository(db *gorm.DB) *AIUsageLogRepository {
	return &AIUsageLogRepository{db: db}
}

func (r *AIUsageLogRepository) Create(log *models.AIUsageLog) error {
	return r.db.Create(log).Error
}

func (r *AIUsageLogRepository) GetTenantUsage(tenantID uuid.UUID, startDate, endDate string) ([]models.AIUsageLog, error) {
	var logs []models.AIUsageLog
	err := r.db.Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, startDate, endDate).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

func (r *AIUsageLogRepository) GetTenantDailySummary(tenantID uuid.UUID, startDate, endDate string) ([]map[string]interface{}, error) {
	var summary []map[string]interface{}
	err := r.db.Table("ai_usage_logs").
		Select("DATE(created_at) as date, SUM(total_tokens) as total_tokens, SUM(total_cost) as total_cost, COUNT(*) as request_count").
		Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, startDate, endDate).
		Group("DATE(created_at)").
		Order("date DESC").
		Find(&summary).Error
	return summary, err
}