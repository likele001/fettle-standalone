package repository

import (
	"ai-platform/skill-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SkillRepository struct {
	db *gorm.DB
}

func NewSkillRepository(db *gorm.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

// Create 创建技能
func (r *SkillRepository) Create(skill *models.Skill) error {
	return r.db.Create(skill).Error
}

// GetByID 根据 ID 获取技能
func (r *SkillRepository) GetByID(id uuid.UUID) (*models.Skill, error) {
	var skill models.Skill
	err := r.db.First(&skill, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

// List 列出所有技能
func (r *SkillRepository) List(page, pageSize int) ([]models.Skill, int64, error) {
	var skills []models.Skill
	var total int64

	r.db.Model(&models.Skill{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&skills).Error
	return skills, total, err
}

// Update 更新技能
func (r *SkillRepository) Update(skill *models.Skill) error {
	return r.db.Save(skill).Error
}

// Delete 删除技能
func (r *SkillRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Skill{}, "id = ?", id).Error
}

// IncrementInstallCount 增加安装次数
func (r *SkillRepository) IncrementInstallCount(id uuid.UUID) error {
	return r.db.Model(&models.Skill{}).Where("id = ?", id).
		UpdateColumn("install_count", gorm.Expr("install_count + 1")).Error
}

// Install 安装技能
func (r *SkillRepository) Install(installation *models.SkillInstallation) error {
	return r.db.Create(installation).Error
}

// GetInstallation 获取安装记录
func (r *SkillRepository) GetInstallation(tenantID, skillID uuid.UUID) (*models.SkillInstallation, error) {
	var installation models.SkillInstallation
	err := r.db.Where("tenant_id = ? AND skill_id = ?", tenantID, skillID).First(&installation).Error
	if err != nil {
		return nil, err
	}
	return &installation, nil
}

// ListInstallations 列出租户安装的技能
func (r *SkillRepository) ListInstallations(tenantID uuid.UUID) ([]models.SkillInstallation, error) {
	var installations []models.SkillInstallation
	err := r.db.Where("tenant_id = ?", tenantID).Find(&installations).Error
	return installations, err
}

// UpdateInstallation 更新安装配置
func (r *SkillRepository) UpdateInstallation(installation *models.SkillInstallation) error {
	return r.db.Save(installation).Error
}

// Uninstall 卸载技能
func (r *SkillRepository) Uninstall(tenantID, skillID uuid.UUID) error {
	return r.db.Where("tenant_id = ? AND skill_id = ?", tenantID, skillID).
		Delete(&models.SkillInstallation{}).Error
}


// ListAllSkills lists all skills (for admin)
func (r *SkillRepository) ListAllSkills(page, pageSize int) ([]models.Skill, int64, error) {
	var skills []models.Skill
	var total int64

	r.db.Model(&models.Skill{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&skills).Error
	return skills, total, err
}

// GetSkillStats returns skill statistics (for admin)
func (r *SkillRepository) GetSkillStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total count
	var total int64
	r.db.Model(&models.Skill{}).Count(&total)
	stats["total"] = total

	// Installed count (unique skills with at least one installation)
	var installed int64
	r.db.Table("skill_installations").
		Where("deleted_at IS NULL").
		Distinct("skill_id").
		Count(&installed)
	stats["installed"] = installed

	// Total installations
	var totalInstalls int64
	r.db.Table("skill_installations").
		Where("deleted_at IS NULL").
		Count(&totalInstalls)
	stats["total_installations"] = totalInstalls

	return stats, nil
}
