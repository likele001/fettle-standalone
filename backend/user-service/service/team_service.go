package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"ai-platform/user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamService 团队管理服务
type TeamService struct {
	DB *gorm.DB
}

func NewTeamService(db *gorm.DB) *TeamService {
	return &TeamService{DB: db}
}

// --- 邀请 ---

type CreateInvitationRequest struct {
	InviteePhone string `json:"invitee_phone"`
	InviteeEmail string `json:"invitee_email"`
	Role         string `json:"role"` // admin / member
}

type InvitationResponse struct {
	Code      string    `json:"code"`
	InviteURL string    `json:"invite_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *TeamService) CreateInvitation(tenantID, inviterID uuid.UUID, req *CreateInvitationRequest) (*InvitationResponse, error) {
	// 检查邀请人权限
	var member models.TenantMember
	if err := s.DB.Where("tenant_id = ? AND user_id = ?", tenantID, inviterID).First(&member).Error; err != nil {
		return nil, errors.New("您不是该租户的成员")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, errors.New("只有所有者和管理员可以邀请成员")
	}

	role := req.Role
	if role != "admin" && role != "member" {
		role = "member"
	}

	// 生成邀请码
	code := generateInviteCode()

	invitation := &models.TenantInvitation{
		TenantID:     tenantID,
		Code:         code,
		InviteePhone: req.InviteePhone,
		InviteeEmail: req.InviteeEmail,
		Role:         role,
		InvitedBy:    inviterID,
	}
	if err := s.DB.Create(invitation).Error; err != nil {
		return nil, err
	}

	return &InvitationResponse{
		Code:      code,
		InviteURL: fmt.Sprintf("/invite?code=%s", code),
		Role:      role,
		CreatedAt: invitation.CreatedAt,
	}, nil
}

func (s *TeamService) GetInvitation(code string) (*models.TenantInvitation, error) {
	var inv models.TenantInvitation
	if err := s.DB.Where("code = ? AND used = ?", code, false).First(&inv).Error; err != nil {
		return nil, errors.New("邀请码无效或已使用")
	}
	return &inv, nil
}

// --- 接受邀请 ---

func (s *TeamService) AcceptInvitation(userID uuid.UUID, code string) error {
	inv, err := s.GetInvitation(code)
	if err != nil {
		return err
	}

	// 检查用户是否已属于其他租户
	var existingMember models.TenantMember
	if err := s.DB.Where("user_id = ?", userID).First(&existingMember).Error; err == nil {
		if existingMember.Status == "active" {
			return errors.New("您已加入其他公司，无法重复加入")
		}
		// 如果之前离职了，更新记录
		s.DB.Model(&existingMember).Updates(map[string]interface{}{
			"tenant_id":   inv.TenantID,
			"role":        inv.Role,
			"status":      "active",
			"invited_by":  inv.InvitedBy,
			"joined_at":   time.Now(),
			"resigned_at": nil,
		})
		// 同时更新 users 表的 tenant_id
		s.DB.Model(&models.User{}).Where("id = ?", userID).Update("tenant_id", inv.TenantID)
	} else {
		// 新成员
		member := &models.TenantMember{
			TenantID:  inv.TenantID,
			UserID:    userID,
			Role:      inv.Role,
			Status:    "active",
			InvitedBy: &inv.InvitedBy,
		}
		if err := s.DB.Create(member).Error; err != nil {
			return err
		}
		// 更新 users 表的 tenant_id
		s.DB.Model(&models.User{}).Where("id = ?", userID).Update("tenant_id", inv.TenantID)
	}

	// 标记邀请码已使用
	s.DB.Model(&models.TenantInvitation{}).Where("id = ?", inv.ID).Update("used", true)

	return nil
}

// --- 成员列表 ---

type MemberListResponse struct {
	Items []models.MemberWithUser `json:"items"`
	Total int64                   `json:"total"`
}

func (s *TeamService) ListMembers(tenantID uuid.UUID, status string, page, pageSize int) (*MemberListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := s.DB.Table("tenant_members tm").
		Select("tm.*, u.name as user_name, u.phone as user_phone, u.email as user_email, u.avatar_url").
		Joins("JOIN users u ON u.id = tm.user_id").
		Where("tm.tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("tm.status = ?", status)
	}

	var total int64
	query.Count(&total)

	var items []models.MemberWithUser
	offset := (page - 1) * pageSize
	query.Order("tm.role DESC, tm.joined_at ASC").
		Offset(offset).Limit(pageSize).
		Scan(&items)

	return &MemberListResponse{Items: items, Total: total}, nil
}

// --- 修改角色 ---

func (s *TeamService) UpdateMemberRole(tenantID, operatorID, memberID uuid.UUID, newRole string) error {
	// 检查操作人权限
	var operator models.TenantMember
	if err := s.DB.Where("tenant_id = ? AND user_id = ?", tenantID, operatorID).First(&operator).Error; err != nil {
		return errors.New("您不是该租户的成员")
	}
	if operator.Role != "owner" {
		return errors.New("只有所有者可以修改成员角色")
	}
	if newRole != "admin" && newRole != "member" {
		return errors.New("无效的角色")
	}

	var member models.TenantMember
	if err := s.DB.Where("tenant_id = ? AND id = ?", tenantID, memberID).First(&member).Error; err != nil {
		return errors.New("成员不存在")
	}
	if member.Role == "owner" {
		return errors.New("不能修改所有者的角色")
	}

	return s.DB.Model(&member).Update("role", newRole).Error
}

// --- 设为离职 ---

func (s *TeamService) ResignMember(tenantID, operatorID, memberID uuid.UUID) error {
	var operator models.TenantMember
	if err := s.DB.Where("tenant_id = ? AND user_id = ?", tenantID, operatorID).First(&operator).Error; err != nil {
		return errors.New("您不是该租户的成员")
	}
	if operator.Role != "owner" && operator.Role != "admin" {
		return errors.New("只有所有者和管理员可以操作成员离职")
	}

	var member models.TenantMember
	if err := s.DB.Where("tenant_id = ? AND id = ?", tenantID, memberID).First(&member).Error; err != nil {
		return errors.New("成员不存在")
	}
	if member.Role == "owner" {
		return errors.New("不能操作所有者离职")
	}

	now := time.Now()
	return s.DB.Model(&member).Updates(map[string]interface{}{
		"status":       "resigned",
		"resigned_at":  now,
	}).Error
}

// --- 取消邀请 ---

func (s *TeamService) CancelInvitation(tenantID, operatorID uuid.UUID, invitationID uuid.UUID) error {
	var operator models.TenantMember
	if err := s.DB.Where("tenant_id = ? AND user_id = ?", tenantID, operatorID).First(&operator).Error; err != nil {
		return errors.New("您不是该租户的成员")
	}
	if operator.Role != "owner" && operator.Role != "admin" {
		return errors.New("只有所有者和管理员可以取消邀请")
	}

	var inv models.TenantInvitation
	if err := s.DB.Where("id = ? AND tenant_id = ? AND used = ?", invitationID, tenantID, false).First(&inv).Error; err != nil {
		return errors.New("邀请不存在或已使用")
	}

	return s.DB.Delete(&inv).Error
}

// --- 邀请列表 ---

func (s *TeamService) ListInvitations(tenantID uuid.UUID) ([]models.TenantInvitation, error) {
	var invitations []models.TenantInvitation
	err := s.DB.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&invitations).Error
	return invitations, err
}

// --- 工具 ---

func generateInviteCode() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
