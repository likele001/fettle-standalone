package service

import (
	"errors"

	"ai-platform/user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionService 权限服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建权限服务
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// GetPermissionsByRole 获取角色的权限列表
func (s *PermissionService) GetPermissionsByRole(tenantID, roleCode string) ([]models.Permission, error) {
	var permissions []models.Permission

	// 通过 role_permissions 关联表查询
	err := s.db.Table("permissions").
		Select("permissions.*").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.tenant_id = ? AND roles.code = ?", tenantID, roleCode).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// CheckPermission 检查用户是否有指定权限
func (s *PermissionService) CheckPermission(tenantID, roleCode, permissionCode string) (bool, error) {
	var count int64

	err := s.db.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.tenant_id = ? AND roles.code = ? AND permissions.code = ?", tenantID, roleCode, permissionCode).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetAllPermissions 获取所有权限
func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	if err := s.db.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetPermissionsByModule 按模块获取权限
func (s *PermissionService) GetPermissionsByModule(module string) ([]models.Permission, error) {
	var permissions []models.Permission
	if err := s.db.Where("module = ?", module).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// CreateRole 创建角色
func (s *PermissionService) CreateRole(tenantID, name, code, description string) (*models.Role, error) {
	// 检查角色是否已存在
	var existing models.Role
	if err := s.db.Where("tenant_id = ? AND code = ?", tenantID, code).First(&existing).Error; err == nil {
		return nil, errors.New("role already exists")
	}

	role := models.Role{
		TenantID:    uuid.MustParse(tenantID),
		Name:        name,
		Code:        code,
		Description: description,
	}

	if err := s.db.Create(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

// AssignPermissionsToRole 为角色分配权限
func (s *PermissionService) AssignPermissionsToRole(roleID string, permissionIDs []string) error {
	// 先删除旧的权限关联
	if err := s.db.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}

	// 创建新的权限关联
	for _, permID := range permissionIDs {
		rp := models.RolePermission{
			RoleID:       uuid.MustParse(roleID),
			PermissionID: uuid.MustParse(permID),
		}
		if err := s.db.Create(&rp).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetRolesByTenant 获取租户下的所有角色
func (s *PermissionService) GetRolesByTenant(tenantID string) ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoleByID 根据ID获取角色
func (s *PermissionService) GetRoleByID(roleID string) (*models.Role, error) {
	id, err := uuid.Parse(roleID)
	if err != nil {
		return nil, errors.New("invalid role id")
	}

	var role models.Role
	if err := s.db.Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetUserRole 获取用户在租户中的角色
func (s *PermissionService) GetUserRole(tenantID, userID string) (*models.Role, error) {
	var role models.Role

	// 通过 users 表关联 roles 表
	err := s.db.Table("roles").
		Select("roles.*").
		Joins("JOIN users ON users.role = roles.code").
		Where("users.tenant_id = ? AND users.id = ?", tenantID, userID).
		First(&role).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user role not found")
		}
		return nil, err
	}

	return &role, nil
}

// DeleteRole 删除角色（不可删除内置角色）
func (s *PermissionService) DeleteRole(tenantID, roleID string) error {
	var role models.Role
	if err := s.db.Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error; err != nil {
		return err
	}
	if role.Code == "super_admin" || role.Code == "admin" || role.Code == "operator" || role.Code == "member" {
		return errors.New("不能删除系统内置角色")
	}
	return s.db.Where("id = ?", roleID).Delete(&models.Role{}).Error
}
