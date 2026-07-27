package models

import (
	"time"

	"github.com/google/uuid"
)

// AIProvider AI 厂商配置（平台层面）
type AIProvider struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code              string    `gorm:"size:50;not null;unique" json:"code"`
	Name              string    `gorm:"size:100;not null" json:"name"`
	NameEn            string    `gorm:"size:100" json:"name_en"`
	LogoURL           string    `gorm:"size:500" json:"logo_url"`
	Description       string    `gorm:"type:text" json:"description"`
	APIBaseURL        string    `gorm:"size:500" json:"api_base_url"`
	AuthType          string    `gorm:"size:20;not null;default:'api_key'" json:"auth_type"`
	AuthConfig        JSONMap   `gorm:"type:jsonb;default:'{}'" json:"auth_config"`
	Status            string    `gorm:"size:20;not null;default:'active'" json:"status"`
	IsDomestic        bool      `gorm:"not null;default:false" json:"is_domestic"`
	SupportStreaming  bool      `gorm:"not null;default:true" json:"support_streaming"`
	SupportVision     bool      `gorm:"not null;default:false" json:"support_vision"`
	SupportFunctionCall bool    `gorm:"not null;default:false" json:"support_function_call"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (AIProvider) TableName() string {
	return "ai_providers"
}

// AIModel AI 模型配置（平台层面）
type AIModel struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProviderID      uuid.UUID `gorm:"type:uuid;not null;index" json:"provider_id"`
	Provider        *AIProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	ModelCode       string    `gorm:"size:100;not null" json:"model_code"`
	ModelName       string    `gorm:"size:200;not null" json:"model_name"`
	ModelType       string    `gorm:"size:20;not null;default:'chat'" json:"model_type"`
	MaxInputTokens  int       `gorm:"not null;default:4096" json:"max_input_tokens"`
	MaxOutputTokens int       `gorm:"not null;default:2048" json:"max_output_tokens"`
	InputPricePer1k float64   `gorm:"not null;default:0.01" json:"input_price_per_1k"`
	OutputPricePer1k float64  `gorm:"not null;default:0.03" json:"output_price_per_1k"`
	Capabilities    []string  `gorm:"type:jsonb;default:'[]';serializer:json" json:"capabilities"`
	Description     string    `gorm:"type:text" json:"description"`
	Status          string    `gorm:"size:20;not null;default:'active'" json:"status"`
	IsDefault       bool      `gorm:"not null;default:false" json:"is_default"`
	Priority        int       `gorm:"not null;default:0" json:"priority"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (AIModel) TableName() string {
	return "ai_models"
}

// TenantAIConfig 租户 AI 配置（租户层面）
type TenantAIConfig struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID             uuid.UUID  `gorm:"type:uuid;not null;unique;index" json:"tenant_id"`
	DefaultChatModelID   *uuid.UUID `gorm:"type:uuid" json:"default_chat_model_id"`
	DefaultChatModel     *AIModel   `gorm:"foreignKey:DefaultChatModelID" json:"default_chat_model,omitempty"`
	DefaultEmbeddingModelID *uuid.UUID `gorm:"type:uuid" json:"default_embedding_model_id"`
	DefaultEmbeddingModel *AIModel  `gorm:"foreignKey:DefaultEmbeddingModelID" json:"default_embedding_model,omitempty"`
	DefaultProviderID    *uuid.UUID `gorm:"type:uuid" json:"default_provider_id"`
	DefaultProvider      *AIProvider `gorm:"foreignKey:DefaultProviderID" json:"default_provider,omitempty"`
	AIEnabled            bool       `gorm:"not null;default:true" json:"ai_enabled"`
	StreamingEnabled     bool       `gorm:"not null;default:true" json:"streaming_enabled"`
	VisionEnabled        bool       `gorm:"not null;default:false" json:"vision_enabled"`
	FunctionCallEnabled  bool       `gorm:"not null;default:false" json:"function_call_enabled"`
	MonthlyTokenLimit    int        `gorm:"not null;default:100000" json:"monthly_token_limit"`
	MonthlyTokenUsed     int        `gorm:"not null;default:0" json:"monthly_token_used"`
	TokenLimitResetAt    *time.Time `json:"token_limit_reset_at"`
	RateLimitPerMinute   int        `gorm:"not null;default:60" json:"rate_limit_per_minute"`
	RateLimitPerDay      int        `gorm:"not null;default:1000" json:"rate_limit_per_day"`
	Config               JSONMap    `gorm:"type:jsonb;default:'{}'" json:"config"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (TenantAIConfig) TableName() string {
	return "tenant_ai_configs"
}

// TenantAPIKey 租户 API Key 配置（租户层面）
type TenantAPIKey struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ProviderID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"provider_id"`
	Provider        *AIProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	APIKeyName      string     `gorm:"size:100;not null" json:"api_key_name"`
	APIKeyValue     string     `gorm:"type:text;not null" json:"api_key_value"` // 返回时手动清空
	APIKeyEncrypted bool       `gorm:"not null;default:false" json:"api_key_encrypted"`
	MonthlyQuota    int        `gorm:"not null;default:0" json:"monthly_quota"`
	MonthlyUsed     int        `gorm:"not null;default:0" json:"monthly_used"`
	Status          string     `gorm:"size:20;not null;default:'active'" json:"status"`
	ExpiredAt       *time.Time `json:"expired_at"`
	CustomBaseURL   string     `gorm:"size:500" json:"custom_base_url"`
	CustomHeaders   JSONMap    `gorm:"type:jsonb;default:'{}'" json:"custom_headers"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastUsedAt      *time.Time `json:"last_used_at"`
}

func (TenantAPIKey) TableName() string {
	return "tenant_api_keys"
}

// AIUsageLog AI 使用日志（用于计费和统计）
type AIUsageLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID         *uuid.UUID `gorm:"type:uuid" json:"user_id"`
	ConversationID *uuid.UUID `gorm:"type:uuid" json:"conversation_id"`
	MessageID      *uuid.UUID `gorm:"type:uuid" json:"message_id"`
	ProviderID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"provider_id"`
	Provider       *AIProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	ModelID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"model_id"`
	Model          *AIModel   `gorm:"foreignKey:ModelID" json:"model,omitempty"`
	InputTokens    int        `gorm:"not null;default:0" json:"input_tokens"`
	OutputTokens   int        `gorm:"not null;default:0" json:"output_tokens"`
	TotalTokens    int        `gorm:"not null;default:0" json:"total_tokens"`
	InputCost      float64    `gorm:"not null;default:0" json:"input_cost"`
	OutputCost     float64    `gorm:"not null;default:0" json:"output_cost"`
	TotalCost      float64    `gorm:"not null;default:0" json:"total_cost"`
	RequestType    string     `gorm:"size:20;not null" json:"request_type"`
	LatencyMs      int        `gorm:"not null;default:0" json:"latency_ms"`
	Success        bool       `gorm:"not null;default:true" json:"success"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	CreatedAt      time.Time  `gorm:"index" json:"created_at"`
}

func (AIUsageLog) TableName() string {
	return "ai_usage_logs"
}