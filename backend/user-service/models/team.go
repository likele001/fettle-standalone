package models

import (
	"time"

	"github.com/google/uuid"
)

// TenantMember 租户成员
type TenantMember struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	Role       string     `gorm:"size:20;not null;default:'member'" json:"role"` // owner/admin/member
	Status     string     `gorm:"size:20;not null;default:'active'" json:"status"` // active/resigned
	InvitedBy  *uuid.UUID `gorm:"type:uuid" json:"invited_by"`
	JoinedAt   time.Time  `gorm:"not null;default:now()" json:"joined_at"`
	ResignedAt *time.Time `json:"resigned_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (TenantMember) TableName() string { return "tenant_members" }

// TenantInvitation 租户邀请码
type TenantInvitation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Code         string     `gorm:"size:64;not null;uniqueIndex" json:"code"`
	InviteePhone string     `gorm:"size:20" json:"invitee_phone"`
	InviteeEmail string     `gorm:"size:100" json:"invitee_email"`
	Role         string     `gorm:"size:20;not null;default:'member'" json:"role"`
	InvitedBy    uuid.UUID  `gorm:"type:uuid;not null" json:"invited_by"`
	Used         bool       `gorm:"not null;default:false" json:"used"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (TenantInvitation) TableName() string { return "tenant_invitations" }

// MemberWithUser 成员列表（带用户信息）
type MemberWithUser struct {
	TenantMember
	UserName  string `json:"user_name"`
	UserPhone string `json:"user_phone"`
	UserEmail string `json:"user_email"`
	AvatarURL string `json:"avatar_url"`
}
