package models

import (
	"time"

	"github.com/google/uuid"
)

// Role 角色模型
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Code        string    `gorm:"size:50;not null" json:"code"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 表名
func (Role) TableName() string {
	return "roles"
}

// RolePermission 角色权限关联
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 表名
func (RolePermission) TableName() string {
	return "role_permissions"
}
