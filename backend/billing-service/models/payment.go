package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentConfig 支付配置
type PaymentConfig struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppID      string         `gorm:"size:32;not null" json:"app_id"`
	AppSecret  string         `gorm:"size:64;not null" json:"app_secret"`
	GatewayURL string         `gorm:"size:128;not null" json:"gateway_url"`
	NotifyURL  string         `gorm:"size:256;not null" json:"notify_url"`
	ReturnURL  string         `gorm:"size:256;not null" json:"return_url"`
	Enabled    bool           `gorm:"not null;default:false" json:"enabled"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// PaymentOrder 支付订单
type PaymentOrder struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	PlanID         string         `gorm:"size:32;not null" json:"plan_id"`
	PlanName       string         `gorm:"size:64;not null" json:"plan_name"`
	Amount         float64        `gorm:"type:decimal(18,2);not null" json:"amount"`
	TradeOrderID   string         `gorm:"size:64;uniqueIndex;not null" json:"trade_order_id"`
	Status         string         `gorm:"size:16;not null;default:'pending'" json:"status"` // pending, paid, failed, cancelled
	PaymentURL     string         `gorm:"type:text;not null" json:"payment_url"`
	TransactionID  string         `gorm:"size:64;not null" json:"transaction_id"`
	PaidAt         *time.Time     `json:"paid_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
