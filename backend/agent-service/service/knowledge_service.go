package service

import (
	"ai-platform/agent-service/grpc_client"
	"ai-platform/agent-service/models"
	"ai-platform/agent-service/repository"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// KnowledgeService 知识库业务逻辑
type KnowledgeService struct {
	repo      *repository.KnowledgeBaseRepository
	agentRepo *repository.AgentRepository
	aiClient  *grpc_client.AIEngineClient
}

func NewKnowledgeService(repo *repository.KnowledgeBaseRepository, agentRepo *repository.AgentRepository, aiClient *grpc_client.AIEngineClient) *KnowledgeService {
	return &KnowledgeService{repo: repo, agentRepo: agentRepo, aiClient: aiClient}
}

type CreateKBRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Config      models.JSONMap `json:"config"`
}

type UpdateKBRequest struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	Config      models.JSONMap `json:"config"`
}

func (s *KnowledgeService) Create(tenantID string, req *CreateKBRequest) (*models.KnowledgeBase, error) {
	// Check KB quota
	kbCount, _ := s.agentRepo.CountKBByTenant(tenantID)
	_, maxKB, _, _ := s.agentRepo.GetPlanQuota(tenantID)
	if kbCount >= maxKB {
		return nil, fmt.Errorf("知识库数量已达上限（%d/%d），请升级套餐以创建更多知识库", kbCount, maxKB)
	}

	kb := &models.KnowledgeBase{
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		Config:      req.Config,
	}
	if kb.Config == nil {
		kb.Config = models.JSONMap{}
	}

	if tid, err := uuid.Parse(tenantID); err == nil {
		kb.TenantID = tid
	}

	if err := s.repo.Create(kb); err != nil {
		return nil, err
	}
	return kb, nil
}

func (s *KnowledgeService) GetByID(tenantID, kbID string) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetByID(tenantID, kbID)
	if err != nil {
		return nil, errors.New("knowledge base not found")
	}
	return kb, nil
}

func (s *KnowledgeService) List(tenantID string, page, pageSize int) ([]models.KnowledgeBase, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListByTenant(tenantID, page, pageSize)
}

func (s *KnowledgeService) Update(tenantID, kbID string, req *UpdateKBRequest) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetByID(tenantID, kbID)
	if err != nil {
		return nil, errors.New("knowledge base not found")
	}

	if req.Name != nil {
		kb.Name = *req.Name
	}
	if req.Description != nil {
		kb.Description = *req.Description
	}
	if req.Config != nil {
		kb.Config = req.Config
	}

	if err := s.repo.Update(kb); err != nil {
		return nil, err
	}
	return kb, nil
}

func (s *KnowledgeService) Delete(tenantID, kbID string) error {
	return s.repo.Delete(tenantID, kbID)
}

// Document methods

func (s *KnowledgeService) AddDocument(tenantID, kbID string, doc *models.KnowledgeDocument) error {
	// Check document quota per KB
	docCount, _ := s.agentRepo.CountDocsByKB(kbID)
	_, _, maxDocsPerKB, _ := s.agentRepo.GetPlanQuota(tenantID)
	if docCount >= maxDocsPerKB {
		return fmt.Errorf("该知识库文档数已达上限（%d/%d），请升级套餐以上传更多文档", docCount, maxDocsPerKB)
	}

	// 验证知识库存在
	_, err := s.repo.GetByID(tenantID, kbID)
	if err != nil {
		return errors.New("knowledge base not found")
	}

	if err := s.repo.CreateDocument(doc); err != nil {
		return err
	}

	// 更新知识库文档计数
	kb, _ := s.repo.GetByID(tenantID, kbID)
	if kb != nil {
		kb.DocCount++
		kb.TotalSize += doc.FileSize
		s.repo.Update(kb)
	}

	return nil
}

func (s *KnowledgeService) ListDocuments(kbID string, page, pageSize int) ([]models.KnowledgeDocument, int64, error) {
	return s.repo.ListDocuments(kbID, page, pageSize)
}

func (s *KnowledgeService) DeleteDocument(tenantID, kbID, docID string) error {
	return s.repo.DeleteDocument(docID)
}

func (s *KnowledgeService) UpdateDocumentStatus(docID, status string) error {
	docIDParsed, err := uuid.Parse(docID)
	if err != nil {
		return err
	}
	return s.repo.UpdateDocumentStatus(docIDParsed, status)
}
