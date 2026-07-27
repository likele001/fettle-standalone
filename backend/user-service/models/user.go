package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Username     string         `gorm:"size:50;uniqueIndex" json:"username"`
	Phone        string         `gorm:"size:20;uniqueIndex" json:"phone"`
	Email        string         `gorm:"size:100" json:"email"`
	PasswordHash string         `gorm:"size:255" json:"-"`
	WxOpenID     string         `gorm:"size:64" json:"wx_openid"`
	Name         string         `gorm:"size:50" json:"name"`
	AvatarURL    string         `json:"avatar_url"`
	Role         string         `gorm:"size:20;not null;default:'member'" json:"role"`
	Status       string         `gorm:"size:10;not null;default:'active'" json:"status"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}
