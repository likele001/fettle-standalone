package models

import (
	"time"

	"github.com/google/uuid"
)

// Permission 权限模型
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string    `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Module      string    `gorm:"size:50;not null" json:"module"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 表名
func (Permission) TableName() string {
	return "permissions"
}
