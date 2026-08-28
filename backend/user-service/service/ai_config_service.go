package service

import (
	"ai-platform/user-service/models"
	"ai-platform/user-service/repository"
	"ai-platform/user-service/utils"
	"context"
	"errors"

	"github.com/google/uuid"
)

// AIConfigService AI 配置服务
type AIConfigService struct {
	providerRepo     *repository.AIProviderRepository
	modelRepo        *repository.AIModelRepository
	tenantConfigRepo *repository.TenantAIConfigRepository
	tenantKeyRepo    *repository.TenantAPIKeyRepository
	usageLogRepo     *repository.AIUsageLogRepository
}

func NewAIConfigService(
	providerRepo *repository.AIProviderRepository,
	modelRepo *repository.AIModelRepository,
	tenantConfigRepo *repository.TenantAIConfigRepository,
	tenantKeyRepo *repository.TenantAPIKeyRepository,
	usageLogRepo *repository.AIUsageLogRepository,
) *AIConfigService {
	return &AIConfigService{
		providerRepo:     providerRepo,
		modelRepo:        modelRepo,
		tenantConfigRepo: tenantConfigRepo,
		tenantKeyRepo:    tenantKeyRepo,
		usageLogRepo:     usageLogRepo,
	}
}

// ========== 平台层面 API ==========

// ListProviders 获取所有 AI 厂商列表
func (s *AIConfigService) ListProviders(ctx context.Context, status string) ([]models.AIProvider, error) {
	return s.providerRepo.List(status)
}

// GetProvider 获取厂商详情
func (s *AIConfigService) GetProvider(ctx context.Context, id uuid.UUID) (*models.AIProvider, error) {
	return s.providerRepo.GetByID(id)
}

// CreateProvider 创建厂商（超管）
func (s *AIConfigService) CreateProvider(ctx context.Context, provider *models.AIProvider) error {
	return s.providerRepo.Create(provider)
}

// UpdateProvider 更新厂商（超管）
func (s *AIConfigService) UpdateProvider(ctx context.Context, provider *models.AIProvider) error {
	return s.providerRepo.Update(provider)
}

// DeleteProvider 删除厂商（超管）
func (s *AIConfigService) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	// 先删除关联的模型
	modelsList, err := s.modelRepo.List(&id, "", "")
	if err != nil {
		return err
	}
	for _, m := range modelsList {
		s.modelRepo.Delete(m.ID)
	}
	return s.providerRepo.Delete(id)
}

// ListModels 获取所有 AI 模型列表
func (s *AIConfigService) ListModels(ctx context.Context, providerID *uuid.UUID, modelType string, status string) ([]models.AIModel, error) {
	return s.modelRepo.List(providerID, modelType, status)
}

// GetModel 获取模型详情
func (s *AIConfigService) GetModel(ctx context.Context, id uuid.UUID) (*models.AIModel, error) {
	return s.modelRepo.GetByID(id)
}

// CreateModel 创建模型（超管）
func (s *AIConfigService) CreateModel(ctx context.Context, model *models.AIModel) error {
	return s.modelRepo.Create(model)
}

// UpdateModel 更新模型（超管）
func (s *AIConfigService) UpdateModel(ctx context.Context, model *models.AIModel) error {
	return s.modelRepo.Update(model)
}

// DeleteModel 删除模型（超管）
func (s *AIConfigService) DeleteModel(ctx context.Context, id uuid.UUID) error {
	return s.modelRepo.Delete(id)
}

// ========== 租户层面 API ==========

// GetTenantAIConfig 获取租户 AI 配置
func (s *AIConfigService) GetTenantAIConfig(ctx context.Context, tenantID uuid.UUID) (*models.TenantAIConfig, error) {
	config, err := s.tenantConfigRepo.GetByTenantID(tenantID)
	if err != nil {
		// 如果不存在，创建默认配置
		config = &models.TenantAIConfig{
			TenantID:           tenantID,
			AIEnabled:          true,
			StreamingEnabled:   true,
			MonthlyTokenLimit:  100000,
			RateLimitPerMinute: 60,
			RateLimitPerDay:    1000,
		}
		if err := s.tenantConfigRepo.Create(config); err != nil {
			return nil, err
		}
	}
	return config, nil
}

// UpdateTenantAIConfig 更新租户 AI 配置
func (s *AIConfigService) UpdateTenantAIConfig(ctx context.Context, tenantID uuid.UUID, config *models.TenantAIConfig) error {
	existing, err := s.tenantConfigRepo.GetByTenantID(tenantID)
	if err != nil {
		return s.tenantConfigRepo.Create(&models.TenantAIConfig{
			TenantID:                tenantID,
			DefaultProviderID:       config.DefaultProviderID,
			DefaultChatModelID:      config.DefaultChatModelID,
			DefaultEmbeddingModelID: config.DefaultEmbeddingModelID,
			AIEnabled:               true,
			StreamingEnabled:        true,
			VisionEnabled:           false,
			FunctionCallEnabled:     false,
			MonthlyTokenLimit:       100000,
			RateLimitPerMinute:      60,
			RateLimitPerDay:         1000,
			Config:                  config.Config,
		})
	}
	if config.DefaultProviderID != nil {
		existing.DefaultProviderID = config.DefaultProviderID
	} else {
		existing.DefaultProviderID = nil
	}
	if config.DefaultChatModelID != nil {
		existing.DefaultChatModelID = config.DefaultChatModelID
	} else {
		existing.DefaultChatModelID = nil
	}
	if config.DefaultEmbeddingModelID != nil {
		existing.DefaultEmbeddingModelID = config.DefaultEmbeddingModelID
	} else {
		existing.DefaultEmbeddingModelID = nil
	}
	if config.Config != nil {
		for k, v := range config.Config {
			existing.Config[k] = v
		}
	}
	return s.tenantConfigRepo.Update(existing)
}

// ListTenantAPIKeys 获取租户 API Key 列表
func (s *AIConfigService) ListTenantAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]models.TenantAPIKey, error) {
	return s.tenantKeyRepo.List(tenantID)
}

// GetTenantAPIKey 获取租户 API Key 详情（包含实际 Key 值，仅内部调用）
func (s *AIConfigService) GetTenantAPIKey(ctx context.Context, id uuid.UUID) (*models.TenantAPIKey, error) {
	return s.tenantKeyRepo.GetByID(id)
}

// CreateTenantAPIKey 创建租户 API Key（自动加密）
func (s *AIConfigService) CreateTenantAPIKey(ctx context.Context, tenantID uuid.UUID, key *models.TenantAPIKey) error {
	key.TenantID = tenantID
	if key.APIKeyValue != "" {
		encrypted, err := utils.AESEncrypt(key.APIKeyValue)
		if err != nil {
			return err
		}
		key.APIKeyValue = encrypted
		key.APIKeyEncrypted = true
	}
	return s.tenantKeyRepo.Create(key)
}

// UpdateTenantAPIKey 更新租户 API Key（自动加密）
func (s *AIConfigService) UpdateTenantAPIKey(ctx context.Context, tenantID uuid.UUID, key *models.TenantAPIKey) error {
	existingKey, err := s.tenantKeyRepo.GetByID(key.ID)
	if err != nil {
		return err
	}
	if existingKey.TenantID != tenantID {
		return ErrUnauthorized
	}
	if key.APIKeyValue != "" && key.APIKeyValue != existingKey.APIKeyValue {
		encrypted, err := utils.AESEncrypt(key.APIKeyValue)
		if err != nil {
			return err
		}
		key.APIKeyValue = encrypted
		key.APIKeyEncrypted = true
	} else if key.APIKeyValue == "" {
		key.APIKeyValue = existingKey.APIKeyValue
		key.APIKeyEncrypted = existingKey.APIKeyEncrypted
	}
	return s.tenantKeyRepo.Update(key)
}

// DeleteTenantAPIKey 删除租户 API Key
func (s *AIConfigService) DeleteTenantAPIKey(ctx context.Context, tenantID uuid.UUID, keyID uuid.UUID) error {
	// 验证 Key 属于该租户
	existingKey, err := s.tenantKeyRepo.GetByID(keyID)
	if err != nil {
		return err
	}
	if existingKey.TenantID != tenantID {
		return ErrUnauthorized
	}
	return s.tenantKeyRepo.Delete(keyID)
}

// GetTenantAPIKeyForUse 获取租户可用的 API Key（内部调用，用于实际调用 AI）
func (s *AIConfigService) GetTenantAPIKeyForUse(ctx context.Context, tenantID, providerID uuid.UUID) (*models.TenantAPIKey, error) {
	return s.tenantKeyRepo.GetByTenantProvider(tenantID, providerID)
}

// ========== 使用日志 ==========

// CreateUsageLog 创建使用日志
func (s *AIConfigService) CreateUsageLog(ctx context.Context, log *models.AIUsageLog) error {
	return s.usageLogRepo.Create(log)
}

// GetTenantUsageSummary 获取租户使用统计
func (s *AIConfigService) GetTenantUsageSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate string) ([]map[string]interface{}, error) {
	return s.usageLogRepo.GetTenantDailySummary(tenantID, startDate, endDate)
}

// SeedProviders 预置常用 AI 厂商和模型
func (s *AIConfigService) SeedProviders(ctx context.Context) error {
	providers := []models.AIProvider{
		{
			Code:       "openai",
			Name:       "OpenAI",
			APIBaseURL: "https://api.openai.com/v1",
			Status:     "active",
		},
		{
			Code:       "anthropic",
			Name:       "Anthropic",
			APIBaseURL: "https://api.anthropic.com/v1",
			Status:     "active",
		},
		{
			Code:       "aliyun",
			Name:       "阿里云通义",
			APIBaseURL: "https://dashscope.aliyuncs.com/api/v1",
			Status:     "active",
		},
		{
			Code:       "baidu",
			Name:       "百度文心",
			APIBaseURL: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1",
			Status:     "active",
		},
		{
			Code:       "deepseek",
			Name:       "DeepSeek",
			APIBaseURL: "https://api.deepseek.com/v1",
			Status:     "active",
		},
	}

	for _, p := range providers {
		// 检查是否已存在
		existing, _ := s.providerRepo.GetByCode(p.Code)
		if existing != nil {
			continue
		}
		if err := s.providerRepo.Create(&p); err != nil {
			return err
		}
	}
	return nil
}

// TestAPIKey 测试 API Key 是否有效
func (s *AIConfigService) TestAPIKey(ctx context.Context, keyID uuid.UUID, tenantID uuid.UUID) (bool, error) {
	key, err := s.tenantKeyRepo.GetByID(keyID)
	if err != nil {
		return false, err
	}
	if key.TenantID != tenantID {
		return false, ErrUnauthorized
	}

	// 这里应该实际调用 AI API 验证，暂时返回 true
	// TODO: 实现实际的 API 调用验证
	return true, nil
}

var ErrUnauthorized = errors.New("unauthorized access")
