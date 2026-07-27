package repository

import (
	"ai-platform/agent-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentRepository 智能体数据访问
type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(agent *models.Agent) error {
	return r.db.Create(agent).Error
}

func (r *AgentRepository) GetByID(tenantID, agentID string) (*models.Agent, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	aID, err := uuid.Parse(agentID)
	if err != nil {
		return nil, err
	}

	var agent models.Agent
	err = r.db.Where("id = ? AND (tenant_id = ? OR is_public = true)", aID, tID).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// getTenantPlanType 获取租户的套餐类型
func (r *AgentRepository) getTenantPlanType(tenantID uuid.UUID) string {
	var planType string
	err := r.db.Table("tenants").Select("plan_type").Where("id = ?", tenantID).Scan(&planType).Error
	if err != nil || planType == "" {
		return "free" // 默认免费版
	}
	return planType
}

// getAccessiblePlans 根据租户套餐类型获取可访问的智能体级别
func getAccessiblePlans(tenantPlan string) []string {
	switch tenantPlan {
	case "enterprise":
		return []string{"free", "pro", "enterprise"}
	case "pro":
		return []string{"free", "pro"}
	default:
		return []string{"free"}
	}
}

func (r *AgentRepository) ListByTenant(tenantID string, page, pageSize int) ([]models.Agent, int64, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, 0, err
	}

	// 获取租户套餐类型
	tenantPlan := r.getTenantPlanType(tID)
	accessiblePlans := getAccessiblePlans(tenantPlan)

	var agents []models.Agent
	var total int64

	// Count: tenant's own agents + public agents within plan limits
	r.db.Model(&models.Agent{}).
		Where("(tenant_id = ? OR (is_public = true AND required_plan IN ?))", tID, accessiblePlans).
		Count(&total)

	// Find
	offset := (page - 1) * pageSize
	err = r.db.Model(&models.Agent{}).
		Where("(tenant_id = ? OR (is_public = true AND required_plan IN ?))", tID, accessiblePlans).
		Offset(offset).Limit(pageSize).Order("is_public DESC, created_at DESC").Find(&agents).Error
	return agents, total, err
}

func (r *AgentRepository) Update(agent *models.Agent) error {
	return r.db.Save(agent).Error
}

func (r *AgentRepository) Delete(tenantID, agentID string) error {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return err
	}
	aID, err := uuid.Parse(agentID)
	if err != nil {
		return err
	}
	return r.db.Where("id = ? AND tenant_id = ?", aID, tID).Delete(&models.Agent{}).Error
}

func (r *AgentRepository) UpdateStatus(agentID string, status string, currentSessions int) error {
	aID, err := uuid.Parse(agentID)
	if err != nil {
		return err
	}
	return r.db.Model(&models.Agent{}).Where("id = ?", aID).
		Updates(map[string]interface{}{
			"status":           status,
			"current_sessions": currentSessions,
		}).Error
}

func (r *AgentRepository) IncrementStats(agentID string, messages int64) error {
	aID, err := uuid.Parse(agentID)
	if err != nil {
		return err
	}
	return r.db.Model(&models.Agent{}).Where("id = ?", aID).
		Updates(map[string]interface{}{
			"total_messages":      gorm.Expr("total_messages + ?", messages),
			"total_conversations": gorm.Expr("total_conversations + 1"),
		}).Error
}

func (r *AgentRepository) GetAvailableAgent(tenantID string) (*models.Agent, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}

	// 获取租户套餐类型
	tenantPlan := r.getTenantPlanType(tID)
	accessiblePlans := getAccessiblePlans(tenantPlan)

	var agent models.Agent
	err = r.db.Model(&models.Agent{}).
		Where("((tenant_id = ? OR is_public = true) AND required_plan IN ?) AND status != 'error' AND current_sessions < max_concurrent", tID, accessiblePlans).
		Order("current_sessions ASC").
		First(&agent).Error
	return &agent, err
}

// CountByTenant counts agents for a tenant
func (r *AgentRepository) CountByTenant(tenantID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Agent{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// GetPlanQuota returns max_agents for the tenant's plan
func (r *AgentRepository) GetPlanQuota(tenantID string) (maxAgents, maxKB, maxDocsPerKB int64, err error) {
	var planType string
	err = r.db.Table("tenants").Select("plan_type").Where("id = ?", tenantID).Scan(&planType).Error
	if err != nil || planType == "" {
		planType = "free"
	}
	var plan struct {
		MaxAgents        int64
		MaxKnowledgeBases int64
		MaxDocumentsPerKB int64
	}
	err = r.db.Table("plans").Select("max_agents, max_knowledge_bases as max_knowledge_bases, max_documents_per_kb as max_documents_per_kb").
		Where("type = ?", planType).Scan(&plan).Error
	if err != nil {
		return 2, 1, 10, nil // defaults
	}
	if plan.MaxAgents <= 0 {
		plan.MaxAgents = 2
	}
	if plan.MaxKnowledgeBases <= 0 {
		plan.MaxKnowledgeBases = 1
	}
	if plan.MaxDocumentsPerKB <= 0 {
		plan.MaxDocumentsPerKB = 10
	}
	return plan.MaxAgents, plan.MaxKnowledgeBases, plan.MaxDocumentsPerKB, nil
}

// CountKBByTenant counts knowledge bases for a tenant
func (r *AgentRepository) CountKBByTenant(tenantID string) (int64, error) {
	var count int64
	err := r.db.Table("knowledge_bases").Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// CountDocsByKB counts documents in a knowledge base
func (r *AgentRepository) CountDocsByKB(kbID string) (int64, error) {
	var count int64
	err := r.db.Table("knowledge_documents").Where("knowledge_base_id = ?", kbID).Count(&count).Error
	return count, err
}

// ListAllAgents lists all agents across all tenants with tenant info (for admin)
func (r *AgentRepository) ListAllAgents(page, pageSize int) ([]models.Agent, int64, error) {
	var agents []models.Agent
	var total int64

	r.db.Model(&models.Agent{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&agents).Error
	return agents, total, err
}

// AgentStatusCount represents a status group count
type AgentStatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// GetAgentStats returns total count and counts by status (for admin)
func (r *AgentRepository) GetAgentStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total count
	var total int64
	r.db.Model(&models.Agent{}).Count(&total)
	stats["total"] = total

	// By status
	var statusCounts []AgentStatusCount
	err := r.db.Model(&models.Agent{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error
	if err != nil {
		return stats, err
	}
	for _, sc := range statusCounts {
		stats[sc.Status] = sc.Count
	}

	return stats, nil
}
