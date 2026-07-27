package repository

import (
	"ai-platform/agent-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KnowledgeBaseRepository 知识库数据访问
type KnowledgeBaseRepository struct {
	db *gorm.DB
}

func NewKnowledgeBaseRepository(db *gorm.DB) *KnowledgeBaseRepository {
	return &KnowledgeBaseRepository{db: db}
}

func (r *KnowledgeBaseRepository) Create(kb *models.KnowledgeBase) error {
	return r.db.Create(kb).Error
}

func (r *KnowledgeBaseRepository) GetByID(tenantID, kbID string) (*models.KnowledgeBase, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	kID, err := uuid.Parse(kbID)
	if err != nil {
		return nil, err
	}

	var kb models.KnowledgeBase
	err = r.db.Where("id = ? AND tenant_id = ?", kID, tID).First(&kb).Error
	return &kb, err
}

func (r *KnowledgeBaseRepository) ListByTenant(tenantID string, page, pageSize int) ([]models.KnowledgeBase, int64, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, 0, err
	}

	var kbs []models.KnowledgeBase
	var total int64

	// Count
	r.db.Model(&models.KnowledgeBase{}).Where("tenant_id = ?", tID).Count(&total)

	// Find
	offset := (page - 1) * pageSize
	err = r.db.Model(&models.KnowledgeBase{}).Where("tenant_id = ?", tID).
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&kbs).Error
	return kbs, total, err
}

func (r *KnowledgeBaseRepository) Update(kb *models.KnowledgeBase) error {
	return r.db.Save(kb).Error
}

func (r *KnowledgeBaseRepository) Delete(tenantID, kbID string) error {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return err
	}
	kID, err := uuid.Parse(kbID)
	if err != nil {
		return err
	}
	return r.db.Where("id = ? AND tenant_id = ?", kID, tID).Delete(&models.KnowledgeBase{}).Error
}

// Document methods

func (r *KnowledgeBaseRepository) CreateDocument(doc *models.KnowledgeDocument) error {
	return r.db.Create(doc).Error
}

func (r *KnowledgeBaseRepository) ListDocuments(kbID string, page, pageSize int) ([]models.KnowledgeDocument, int64, error) {
	kID, err := uuid.Parse(kbID)
	if err != nil {
		return nil, 0, err
	}

	var docs []models.KnowledgeDocument
	var total int64

	// Count
	r.db.Model(&models.KnowledgeDocument{}).Where("knowledge_base_id = ?", kID).Count(&total)

	// Find
	offset := (page - 1) * pageSize
	err = r.db.Model(&models.KnowledgeDocument{}).Where("knowledge_base_id = ?", kID).
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&docs).Error
	return docs, total, err
}

func (r *KnowledgeBaseRepository) UpdateDocumentStatus(docID uuid.UUID, status string) error {
	return r.db.Model(&models.KnowledgeDocument{}).Where("id = ?", docID).Update("status", status).Error
}

func (r *KnowledgeBaseRepository) DeleteDocument(docID string) error {
	dID, err := uuid.Parse(docID)
	if err != nil {
		return err
	}
	return r.db.Where("id = ?", dID).Delete(&models.KnowledgeDocument{}).Error
}
