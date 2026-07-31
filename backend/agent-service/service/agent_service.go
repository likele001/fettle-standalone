package service

import (
	"context"
	"ai-platform/agent-service/grpc_client"
	"ai-platform/agent-service/models"
	"ai-platform/agent-service/repository"
	"errors"
	"fmt"
)

// AgentService 智能体业务逻辑
type AgentService struct {
	repo     *repository.AgentRepository
	aiClient *grpc_client.AIEngineClient
}

func NewAgentService(repo *repository.AgentRepository, aiClient *grpc_client.AIEngineClient) *AgentService {
	return &AgentService{repo: repo, aiClient: aiClient}
}

// CreateAgentRequest 创建智能体请求
type CreateAgentRequest struct {
	Name              string         `json:"name" binding:"required"`
	Description       string         `json:"description"`
	AvatarURL         string         `json:"avatar_url"`
	AgentType         string         `json:"agent_type"`
	PersonalityConfig models.JSONMap `json:"personality_config"`
	WorkflowConfig    models.JSONMap `json:"workflow_config"`
	WorkflowID        string         `json:"workflow_id"`
	WelcomeMessage    string         `json:"welcome_message"`
	WorkHours         models.JSONMap `json:"work_hours"`
	ModelID           string         `json:"model_id"`
	KnowledgeBaseIDs  []string       `json:"knowledge_base_ids"`
	MaxConcurrent     int            `json:"max_concurrent"`
}

// UpdateAgentRequest 更新智能体请求
type UpdateAgentRequest struct {
	Name              *string        `json:"name"`
	Description       *string        `json:"description"`
	AvatarURL         *string        `json:"avatar_url"`
	PersonalityConfig models.JSONMap `json:"personality_config"`
	WorkflowConfig    models.JSONMap `json:"workflow_config"`
	WorkflowID        *string        `json:"workflow_id"`
	WelcomeMessage    *string        `json:"welcome_message"`
	WorkHours         models.JSONMap `json:"work_hours"`
	ModelID           *string        `json:"model_id"`
	KnowledgeBaseIDs  []string       `json:"knowledge_base_ids"`
	MaxConcurrent     *int           `json:"max_concurrent"`
}

func (s *AgentService) Create(tenantID string, req *CreateAgentRequest) (*models.Agent, error) {
	// Check agent quota
	count, _ := s.repo.CountByTenant(tenantID)
	maxAgents, _, _, _ := s.repo.GetPlanQuota(tenantID)
	if count >= maxAgents {
		return nil, fmt.Errorf("智能体数量已达上限（%d/%d），请升级套餐以创建更多智能体", count, maxAgents)
	}

	agent := &models.Agent{
		Name:              req.Name,
		Description:       req.Description,
		AvatarURL:         req.AvatarURL,
		AgentType:         req.AgentType,
		PersonalityConfig: req.PersonalityConfig,
		WorkflowConfig:    req.WorkflowConfig,
		WorkflowID:        req.WorkflowID,
		WelcomeMessage:    req.WelcomeMessage,
		WorkHours:         req.WorkHours,
		ModelID:           req.ModelID,
		MaxConcurrent:     req.MaxConcurrent,
		Status:            "idle",
	}

	if agent.AgentType == "" {
		agent.AgentType = "chat"
	}
	if agent.MaxConcurrent <= 0 {
		agent.MaxConcurrent = 10
	}
	if agent.PersonalityConfig == nil {
		agent.PersonalityConfig = models.JSONMap{}
	}
	if agent.WorkflowConfig == nil {
		agent.WorkflowConfig = models.JSONMap{}
	}
	if agent.WorkHours == nil {
		agent.WorkHours = models.JSONMap{
			"start": "09:00",
			"end":   "18:00",
			"days":  []int{1, 2, 3, 4, 5},
		}
	}

	// 解析知识库ID
	if len(req.KnowledgeBaseIDs) > 0 {
		// 存储为JSON数组
		agent.KnowledgeBaseIDs = nil
	}

	if err := s.repo.Create(agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *AgentService) GetByID(tenantID, agentID string) (*models.Agent, error) {
	agent, err := s.repo.GetByID(tenantID, agentID)
	if err != nil {
		return nil, errors.New("agent not found")
	}
	return agent, nil
}

func (s *AgentService) List(tenantID string, page, pageSize int) ([]models.Agent, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListByTenant(tenantID, page, pageSize)
}

func (s *AgentService) Update(tenantID, agentID string, req *UpdateAgentRequest) (*models.Agent, error) {
	agent, err := s.repo.GetByID(tenantID, agentID)
	if err != nil {
		return nil, errors.New("agent not found")
	}

	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Description != nil {
		agent.Description = *req.Description
	}
	if req.AvatarURL != nil {
		agent.AvatarURL = *req.AvatarURL
	}
	if req.PersonalityConfig != nil {
		agent.PersonalityConfig = req.PersonalityConfig
	}
	if req.WorkflowConfig != nil {
		agent.WorkflowConfig = req.WorkflowConfig
	}
	if req.WorkflowID != nil {
		agent.WorkflowID = *req.WorkflowID
	}
	if req.WelcomeMessage != nil {
		agent.WelcomeMessage = *req.WelcomeMessage
	}
	if req.WorkHours != nil {
		agent.WorkHours = req.WorkHours
	}
	if req.ModelID != nil {
		agent.ModelID = *req.ModelID
	}
	if req.MaxConcurrent != nil {
		agent.MaxConcurrent = *req.MaxConcurrent
	}

	if err := s.repo.Update(agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *AgentService) Delete(tenantID, agentID string) error {
	return s.repo.Delete(tenantID, agentID)
}

func (s *AgentService) GetAvailableAgent(tenantID string) (*models.Agent, error) {
	return s.repo.GetAvailableAgent(tenantID)
}

func (s *AgentService) UpdateStatus(agentID string, status string, currentSessions int) error {
	return s.repo.UpdateStatus(agentID, status, currentSessions)
}

// ListAllAgents lists all agents across all tenants (admin)
func (s *AgentService) ListAllAgents(page, pageSize int) ([]models.Agent, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListAllAgents(page, pageSize)
}

// GetAgentStats returns agent statistics (admin)
func (s *AgentService) GetAgentStats() (map[string]int64, error) {
	return s.repo.GetAgentStats()
}

// TestChat 测试智能体对话（通过 gRPC 调用 AI Engine）
func (s *AgentService) TestChat(ctx context.Context, tenantID, userID, agentID, message string) (string, error) {
	agent, err := s.repo.GetByID(tenantID, agentID)
	if err != nil {
		return "", fmt.Errorf("agent not found: %w", err)
	}

	reply, err := s.aiClient.GenerateReply(ctx, tenantID, userID, message, agent.ID.String(), "", nil)
	if err != nil {
		return "", fmt.Errorf("ai engine call failed: %w", err)
	}

	return reply, nil
}
