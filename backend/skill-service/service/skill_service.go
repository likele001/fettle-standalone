package service

import (
	"ai-platform/skill-service/models"
	"ai-platform/skill-service/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SkillService struct {
	repo *repository.SkillRepository
}

func NewSkillService(db *gorm.DB) *SkillService {
	return &SkillService{
		repo: repository.NewSkillRepository(db),
	}
}

// CreateSkill 创建技能
func (s *SkillService) CreateSkill(skill *models.Skill) error {
	return s.repo.Create(skill)
}

// GetSkill 获取技能详情
func (s *SkillService) GetSkill(id uuid.UUID) (*models.Skill, error) {
	return s.repo.GetByID(id)
}

// ListSkills 列出所有技能
func (s *SkillService) ListSkills(page, pageSize int) ([]models.Skill, int64, error) {
	return s.repo.List(page, pageSize)
}

// UpdateSkill 更新技能
func (s *SkillService) UpdateSkill(skill *models.Skill) error {
	return s.repo.Update(skill)
}

// DeleteSkill 删除技能
func (s *SkillService) DeleteSkill(id uuid.UUID) error {
	return s.repo.Delete(id)
}

// InstallSkill 安装技能
func (s *SkillService) InstallSkill(tenantID, skillID uuid.UUID, config string) error {
	// 检查是否已安装
	if _, err := s.repo.GetInstallation(tenantID, skillID); err == nil {
		return nil // 已安装
	}

	installation := &models.SkillInstallation{
		TenantID: tenantID,
		SkillID:  skillID,
		Status:   "active",
		Config:   config,
	}

	if err := s.repo.Install(installation); err != nil {
		return err
	}

	// 增加安装次数
	return s.repo.IncrementInstallCount(skillID)
}

// GetInstalledSkills 获取租户已安装的技能
func (s *SkillService) GetInstalledSkills(tenantID uuid.UUID) ([]models.SkillMarketItem, error) {
	installations, err := s.repo.ListInstallations(tenantID)
	if err != nil {
		return nil, err
	}

	var items []models.SkillMarketItem
	for _, inst := range installations {
		skill, err := s.repo.GetByID(inst.SkillID)
		if err != nil {
			continue
		}
		items = append(items, models.SkillMarketItem{
			ID:           skill.ID,
			Name:         skill.Name,
			Description:  skill.Description,
			Category:     skill.Category,
			Icon:         skill.Icon,
			Version:      skill.Version,
			Author:       skill.Author,
			InstallCount: skill.InstallCount,
			Installed:    true,
		})
	}
	return items, nil
}

// UninstallSkill 卸载技能
func (s *SkillService) UninstallSkill(tenantID, skillID uuid.UUID) error {
	return s.repo.Uninstall(tenantID, skillID)
}

// ToggleSkillStatus 切换技能状态
func (s *SkillService) ToggleSkillStatus(tenantID, skillID uuid.UUID, status string) error {
	installation, err := s.repo.GetInstallation(tenantID, skillID)
	if err != nil {
		return err
	}
	installation.Status = status
	return s.repo.UpdateInstallation(installation)
}

// GetSkillConfig 获取技能配置
func (s *SkillService) GetSkillConfig(tenantID, skillID uuid.UUID) (string, error) {
	installation, err := s.repo.GetInstallation(tenantID, skillID)
	if err != nil {
		return "", err
	}
	return installation.Config, nil
}

// UpdateSkillConfig 更新技能配置
func (s *SkillService) UpdateSkillConfig(tenantID, skillID uuid.UUID, config string) error {
	installation, err := s.repo.GetInstallation(tenantID, skillID)
	if err != nil {
		return err
	}
	installation.Config = config
	return s.repo.UpdateInstallation(installation)
}


// ListAllSkills lists all skills (admin)
func (s *SkillService) ListAllSkills(page, pageSize int) ([]models.Skill, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListAllSkills(page, pageSize)
}

// GetSkillStats returns skill statistics (admin)
func (s *SkillService) GetSkillStats() (map[string]int64, error) {
	return s.repo.GetSkillStats()
}
