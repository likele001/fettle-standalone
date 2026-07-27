package service

import (
	"errors"

	"ai-platform/user-service/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Phone  string `json:"phone" binding:"required"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar_url"`
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers 用户列表
func (s *UserService) ListUsers(tenantID string, page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	query := s.db.Where("tenant_id = ?", tenantID)
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(tenantID string, req *CreateUserRequest) (*models.User, error) {
	var existing models.User
	if err := s.db.Where("phone = ?", req.Phone).First(&existing).Error; err == nil {
		return nil, errors.New("phone already registered")
	}
	if req.Email != "" {
		if err := s.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
			return nil, errors.New("email already registered")
		}
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := models.User{
		TenantID:     tenantUUID,
		Phone:        req.Phone,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         req.Role,
		Status:       req.Status,
	}
	if user.Role == "" {
		user.Role = "member"
	}
	if user.Status == "" {
		user.Status = "active"
	}
	if user.Name == "" {
		user.Name = req.Phone
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(userID string, req *UpdateUserRequest) (*models.User, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// 检查邮箱是否被占用
		var existing models.User
		if err := s.db.Where("email = ? AND id != ?", req.Email, user.ID).First(&existing).Error; err == nil {
			return nil, errors.New("email already in use")
		}
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.AvatarURL = req.Avatar
	}
	if err := s.db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(tenantID, userID string) error {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user id")
	}
	result := s.db.Where("id = ? AND tenant_id = ?", parsedUserID, tenantID).Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateUserRole 更新用户角色
func (s *UserService) UpdateUserRole(tenantID, userID, role string) (*models.User, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user.TenantID.String() != tenantID {
		return nil, errors.New("user not found in tenant")
	}
	user.Role = role
	if err := s.db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
