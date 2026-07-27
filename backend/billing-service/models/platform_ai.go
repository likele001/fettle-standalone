package models

import (
	"time"

	"github.com/google/uuid"
)

// PlatformModelPricing 平台模型定价
type PlatformModelPricing struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ModelID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"model_id"`
	ModelName         string    `gorm:"size:200" json:"model_name"`
	Provider          string    `gorm:"size:100" json:"provider"`
	InputPricePer1K   float64   `gorm:"type:decimal(10,6);not null;default:0" json:"input_price_per_1k"`
	OutputPricePer1K  float64   `gorm:"type:decimal(10,6);not null;default:0" json:"output_price_per_1k"`
	IsEnabled         bool      `gorm:"not null;default:true" json:"is_enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (PlatformModelPricing) TableName() string {
	return "platform_model_pricing"
}

// TenantBalance 租户余额
type TenantBalance struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"tenant_id"`
	Balance        float64   `gorm:"type:decimal(12,4);not null;default:0" json:"balance"`
	TotalRecharged float64   `gorm:"type:decimal(12,4);not null;default:0" json:"total_recharged"`
	TotalConsumed  float64   `gorm:"type:decimal(12,4);not null;default:0" json:"total_consumed"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (TenantBalance) TableName() string {
	return "tenant_balances"
}

// ResourcePackage 资源包定义
type ResourcePackage struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	TokenAmount int64     `gorm:"not null" json:"token_amount"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Description string    `gorm:"size:500" json:"description"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	SortOrder   int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ResourcePackage) TableName() string {
	return "resource_packages"
}

// TenantResourcePackage 租户已购资源包
type TenantResourcePackage struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	PackageID      uuid.UUID  `gorm:"type:uuid;not null" json:"package_id"`
	TokenName      string     `gorm:"size:100" json:"token_name"`
	TokenAmount    int64      `gorm:"not null" json:"token_amount"`
	TokenUsed      int64      `gorm:"not null;default:0" json:"token_used"`
	TokenRemaining int64      `gorm:"not null" json:"token_remaining"`
	Price          float64    `gorm:"type:decimal(10,2);not null" json:"price"`
	Status         string     `gorm:"size:20;not null;default:'active'" json:"status"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (TenantResourcePackage) TableName() string {
	return "tenant_resource_packages"
}

// TenantUsageDetail 用量明细
type TenantUsageDetail struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ModelID      uuid.UUID  `gorm:"type:uuid" json:"model_id"`
	InputTokens  int64      `gorm:"not null;default:0" json:"input_tokens"`
	OutputTokens int64      `gorm:"not null;default:0" json:"output_tokens"`
	TotalTokens  int64      `gorm:"not null;default:0" json:"total_tokens"`
	Cost         float64    `gorm:"type:decimal(10,6);not null;default:0" json:"cost"`
	BillingMode  string     `gorm:"size:20;not null" json:"billing_mode"`
	PackageID    *uuid.UUID `gorm:"type:uuid" json:"package_id"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (TenantUsageDetail) TableName() string {
	return "tenant_usage_details"
}

// TenantRechargeRecord 充值记录
type TenantRechargeRecord struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Amount        float64   `gorm:"type:decimal(12,4);not null" json:"amount"`
	PaymentMethod string    `gorm:"size:50" json:"payment_method"`
	Status        string    `gorm:"size:20;not null;default:'completed'" json:"status"`
	Remark        string    `gorm:"size:500" json:"remark"`
	OperatedBy    string    `gorm:"size:100" json:"operated_by"`
	CreatedAt     time.Time `json:"created_at"`
}

func (TenantRechargeRecord) TableName() string {
	return "tenant_recharge_records"
}
