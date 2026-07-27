package repository

import (
	"errors"

	"ai-platform/user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户数据访问层
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(userID string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByPhone 根据手机号获取用户
func (r *UserRepository) GetByPhone(phone string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByWxOpenID 根据微信OpenID获取用户
func (r *UserRepository) GetByWxOpenID(openID string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("wx_open_id = ?", openID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}

// ListByTenant 获取租户下的用户列表
func (r *UserRepository) ListByTenant(tenantID string, page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Where("tenant_id = ?", tenantID)
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateLastLoginAt 更新最后登录时间
func (r *UserRepository) UpdateLastLoginAt(userID string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("last_login_at", gorm.Expr("NOW()")).Error
}

// UpdatePassword 更新密码
func (r *UserRepository) UpdatePassword(userID, hashedPassword string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("password_hash", hashedPassword).Error
}

// UpdateRole 更新用户角色
func (r *UserRepository) UpdateRole(userID, role string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("role", role).Error
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(userID, status string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("status", status).Error
}

// ExistsByPhone 检查手机号是否存在
func (r *UserRepository) ExistsByPhone(phone string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountByTenant 统计租户下用户数
func (r *UserRepository) CountByTenant(tenantID string) (int64, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
