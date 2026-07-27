package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Channel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Type      string         `gorm:"size:20;not null;uniqueIndex:idx_tenant_channel" json:"type"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Config    string         `gorm:"type:jsonb;default:'{}'" json:"config"`
	Status    string         `gorm:"size:20;default:'inactive'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Channel) TableName() string {
	return "channels"
}
