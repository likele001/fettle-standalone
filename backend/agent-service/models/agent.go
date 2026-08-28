package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Agent 智能体模型
type Agent struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name              string         `gorm:"size:100;not null" json:"name"`
	Description       string         `gorm:"size:500" json:"description"`
	AvatarURL         string         `gorm:"size:255" json:"avatar_url"`
	Status            string         `gorm:"size:20;not null;default:'idle'" json:"status"` // idle, processing, waiting, error
	AgentType         string         `gorm:"size:50;not null;default:'chat'" json:"agent_type"`
	PersonalityConfig JSONMap        `gorm:"type:jsonb;default:'{}'" json:"personality_config"`
	WorkflowConfig    JSONMap        `gorm:"type:jsonb;default:'{}'" json:"workflow_config"`
	WorkflowID        string         `gorm:"size:100;default:''" json:"workflow_id"`
	WelcomeMessage    string         `gorm:"size:500" json:"welcome_message"`
	WorkHours         JSONMap        `gorm:"type:jsonb;default:'{}'" json:"work_hours"` // {"start":"09:00","end":"18:00","days":[1,2,3,4,5]}
	ModelID           string         `gorm:"size:50;default:'default'" json:"model_id"`
	KnowledgeBaseIDs  []string         `gorm:"type:jsonb;default:'[]';serializer:json" json:"knowledge_base_ids"`
	MaxConcurrent     int            `gorm:"default:10" json:"max_concurrent"`
	CurrentSessions   int            `gorm:"default:0" json:"current_sessions"`
	TotalConversations int64         `gorm:"default:0" json:"total_conversations"`
	TotalMessages     int64          `gorm:"default:0" json:"total_messages"`
	IsPublic          bool           `gorm:"default:false;index" json:"is_public"`
	RequiredPlan      string         `gorm:"size:20;default:'free';index" json:"required_plan"` // free, pro, enterprise
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Agent) TableName() string {
	return "agents"
}

// KnowledgeBase 知识库模型
type KnowledgeBase struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Status      string         `gorm:"size:20;not null;default:'active'" json:"status"` // active, processing, error
	DocCount    int            `gorm:"default:0" json:"doc_count"`
	ChunkCount  int            `gorm:"default:0" json:"chunk_count"`
	TotalSize   int64          `gorm:"default:0" json:"total_size"` // bytes
	Config      JSONMap        `gorm:"type:jsonb;default:'{}'" json:"config"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

// KnowledgeDocument 知识库文档
type KnowledgeDocument struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KnowledgeBaseID uuid.UUID `gorm:"type:uuid;not null;index" json:"knowledge_base_id"`
	TenantID       uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	FileName       string    `gorm:"size:255;not null" json:"file_name"`
	FileType       string    `gorm:"size:20;not null" json:"file_type"` // pdf, docx, txt, xlsx
	FileSize       int64     `gorm:"not null" json:"file_size"`
	FileURL        string    `gorm:"size:500" json:"file_url"` // MinIO URL
	Status         string    `gorm:"size:20;not null;default:'pending'" json:"status"` // pending, processing, completed, failed
	ChunkCount     int       `gorm:"default:0" json:"chunk_count"`
	ErrorMessage   string    `gorm:"size:500" json:"error_message"`
	ProcessedAt    *time.Time `json:"processed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (KnowledgeDocument) TableName() string {
	return "knowledge_documents"
}

// JSONMap JSON类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	case map[string]interface{}:
		*j = JSONMap(v)
		return nil
	default:
		*j = JSONMap{}
		return nil
	}
}

// AutoMigrate 自动迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Agent{},
		&KnowledgeBase{},
		&KnowledgeDocument{},
	)
}
