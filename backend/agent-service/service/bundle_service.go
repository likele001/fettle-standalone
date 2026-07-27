package service

import (
	"ai-platform/agent-service/models"
	"ai-platform/agent-service/repository"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type BundleService struct {
	bundleRepo *repository.BundleRepository
}

func NewBundleService(bundleRepo *repository.BundleRepository) *BundleService {
	return &BundleService{bundleRepo: bundleRepo}
}

// ListBundles returns all active industry bundles
func (s *BundleService) ListBundles() ([]models.IndustryBundle, error) {
	return s.bundleRepo.ListActive()
}

// GetBundle returns a single bundle by industry
func (s *BundleService) GetBundle(industry string) (*models.IndustryBundle, error) {
	return s.bundleRepo.GetByIndustry(industry)
}

// ApplyBundleRequest 应用行业套餐请求
type ApplyBundleRequest struct {
	Industry string `json:"industry" binding:"required"`
}

// ApplyBundleResult 应用结果
type ApplyBundleResult struct {
	AgentsCreated   int      `json:"agents_created"`
	KBCreated       int      `json:"kb_created"`
	AgentNames      []string `json:"agent_names"`
	KBNames         []string `json:"kb_names"`
	SkippedReason   string   `json:"skipped_reason,omitempty"`
}

// ApplyBundle applies an industry bundle to a tenant
func (s *BundleService) ApplyBundle(tenantID string, industry string) (*ApplyBundleResult, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}

	// Get the bundle
	bundle, err := s.bundleRepo.GetByIndustry(industry)
	if err != nil {
		return nil, fmt.Errorf("行业套餐 '%s' 不存在", industry)
	}

	// Parse bundle config
	configData, err := json.Marshal(bundle.Config)
	if err != nil {
		return nil, errors.New("套餐配置解析失败")
	}

	var config struct {
		Agents      []models.BundleAgentConfig `json:"agents"`
		KBTemplates []models.BundleKBTemplate  `json:"kb_templates"`
	}
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, errors.New("套餐配置格式错误")
	}

	// Check agent limit
	currentCount, err := s.bundleRepo.CountTenantAgents(tID)
	if err != nil {
		return nil, errors.New("查询当前智能体数量失败")
	}
	maxAgents := s.bundleRepo.GetTenantMaxAgents(tID)
	agentsToAdd := len(config.Agents)

	result := &ApplyBundleResult{
		AgentNames: make([]string, 0),
		KBNames:    make([]string, 0),
	}

	if int(currentCount)+agentsToAdd > maxAgents {
		result.SkippedReason = fmt.Sprintf("当前套餐最多容纳 %d 个智能体，已有 %d 个，本套餐包含 %d 个，请先升级套餐或删除部分智能体", maxAgents, currentCount, agentsToAdd)
		return result, nil
	}

	// Create knowledge base templates first, collect their IDs
	kbIDMap := make(map[string]uuid.UUID) // template name -> KB ID
	for _, tmpl := range config.KBTemplates {
		kb := &models.KnowledgeBase{
			TenantID:    tID,
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Status:      "active",
			Config:      models.JSONMap{},
		}
		if err := s.bundleRepo.CreateKB(kb); err != nil {
			log.Printf("[BundleService] 创建知识库模板 '%s' 失败: %v", tmpl.Name, err)
			continue
		}
		kbIDMap[tmpl.Name] = kb.ID
		result.KBCreated++
		result.KBNames = append(result.KBNames, tmpl.Name)
	}

	// Create agents
	for _, agentCfg := range config.Agents {
		agent := &models.Agent{
			TenantID:          tID,
			Name:              agentCfg.Name,
			Description:       agentCfg.Description,
			AgentType:         agentCfg.AgentType,
			WelcomeMessage:    agentCfg.WelcomeMessage,
			PersonalityConfig: agentCfg.PersonalityConfig,
			RequiredPlan:      agentCfg.RequiredPlan,
			Status:            "idle",
			ModelID:           "default",
			MaxConcurrent:     10,
		}

		if agent.AgentType == "" {
			agent.AgentType = "chat"
		}
		if agent.PersonalityConfig == nil {
			agent.PersonalityConfig = models.JSONMap{}
		}

		// Link knowledge bases
		var kbIDs []string
		for _, kbName := range agentCfg.KBTemplateNames {
			if kbID, ok := kbIDMap[kbName]; ok {
				kbIDs = append(kbIDs, kbID.String())
			}
		}
		if len(kbIDs) > 0 {
			agent.KnowledgeBaseIDs = kbIDs
		}

		if err := s.bundleRepo.CreateAgent(agent); err != nil {
			log.Printf("[BundleService] 创建智能体 '%s' 失败: %v", agentCfg.Name, err)
			continue
		}
		result.AgentsCreated++
		result.AgentNames = append(result.AgentNames, agentCfg.Name)
	}

	return result, nil
}
