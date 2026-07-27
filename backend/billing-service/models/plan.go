package models

import (
	"time"

	"github.com/google/uuid"
)

// Plan 套餐计划 - maps to the unified 'plans' table
type Plan struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name                  string         `gorm:"size:100;not null" json:"name"`
	Type                  string         `gorm:"column:type;size:20;not null;default:'free'" json:"type"`
	Description           string         `gorm:"size:500" json:"description"`
	Price                 float64        `gorm:"column:price;type:decimal(10,2);not null;default:0" json:"price"`
	PriceMonthly          float64        `gorm:"column:price_monthly;type:decimal(10,2);not null;default:0" json:"price_monthly"`
	PriceYearly           float64        `gorm:"type:decimal(10,2);not null;default:0" json:"price_yearly"`
	Period                string         `gorm:"size:20;not null;default:'month'" json:"period"`
	MaxAgents             int            `gorm:"not null;default:2" json:"max_agents"`
	MaxMessages           int64          `gorm:"not null;default:100" json:"max_messages"`
	MaxKnowledgeBases     int            `gorm:"not null;default:1" json:"max_knowledge_bases"`
	MaxDocumentsPerKB     int            `gorm:"not null;default:10" json:"max_documents_per_kb"`
	MaxMessagesPerMonth   int64          `gorm:"column:max_messages_per_month;not null;default:100" json:"max_messages_per_month"`
	MaxConcurrentSessions int            `gorm:"column:max_concurrent_sessions;not null;default:5" json:"max_concurrent_sessions"`
	Features              string         `gorm:"type:jsonb;default:'[]'" json:"features"`
	AllowedIndustries     string         `gorm:"type:jsonb;default:'[]'" json:"allowed_industries"`
	AllowedFeatures       string         `gorm:"type:jsonb;default:'[]'" json:"allowed_features"`
	IsActive              bool           `gorm:"not null;default:true" json:"is_active"`
	Status                string         `gorm:"size:20;not null;default:'active'" json:"status"`
	SortOrder             int            `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             *time.Time     `json:"deleted_at"`
}

func (Plan) TableName() string {
	return "plans"
}

// Subscription 订阅记录
type Subscription struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID            uuid.UUID      `gorm:"type:uuid;not null;index:idx_sub_tenant,unique" json:"tenant_id"`
	PlanID              uuid.UUID      `gorm:"type:uuid;not null" json:"plan_id"`
	Status              string         `gorm:"size:20;not null;default:'active'" json:"status"`
	StartDate           time.Time      `gorm:"column:start_date;not null;default:now()" json:"start_date"`
	EndDate             time.Time      `json:"end_date"`
	NextBillingDate     time.Time      `json:"next_billing_date"`
	CurrentPeriodStart  time.Time      `gorm:"not null;default:now()" json:"current_period_start"`
	CurrentPeriodEnd    time.Time      `json:"current_period_end"`
	MessagesUsed        int64          `gorm:"column:total_tokens_used;not null;default:0" json:"messages_used"`
	TokenLimit          int            `gorm:"not null;default:0" json:"token_limit"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// BillingRecord 计费记录
type BillingRecord struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	SubscriptionID uuid.UUID     `gorm:"type:uuid" json:"subscription_id"`
	Type          string         `gorm:"column:record_type;size:20;not null;default:'usage'" json:"type"`
	Amount        float64        `gorm:"type:decimal(10,2);not null;default:0" json:"amount"`
	Currency      string         `gorm:"size:10;not null;default:'CNY'" json:"currency"`
	Description   string         `gorm:"size:500" json:"description"`
	TokensUsed    int            `gorm:"not null;default:0" json:"tokens_used"`
	InputTokens   int            `gorm:"not null;default:0" json:"input_tokens"`
	OutputTokens  int            `gorm:"not null;default:0" json:"output_tokens"`
	ModelID       uuid.UUID      `json:"model_id"`
	ProviderID    uuid.UUID      `json:"provider_id"`
	Status        string         `gorm:"size:20;not null;default:'pending'" json:"status"`
	TransactionID string         `gorm:"size:200" json:"transaction_id"`
	CreatedAt     time.Time      `json:"created_at"`
}

func (BillingRecord) TableName() string {
	return "billing_records"
}
