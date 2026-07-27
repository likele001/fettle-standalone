package models

import (
	"time"

	"github.com/google/uuid"
)

// IndustryBundle 行业套餐
type IndustryBundle struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Industry    string    `gorm:"size:50;not null;uniqueIndex" json:"industry"` // ecommerce, dining, legal, education, medical
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Icon        string    `gorm:"size:100" json:"icon"`
	Config      JSONMap   `gorm:"type:jsonb;default:'{}'" json:"config"` // {agents: [...], kb_templates: [...]}
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (IndustryBundle) TableName() string {
	return "industry_bundles"
}

// BundleAgentConfig 套餐内智能体配置模板
type BundleAgentConfig struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	AgentType         string   `json:"agent_type"`
	WelcomeMessage    string   `json:"welcome_message"`
	PersonalityConfig JSONMap  `json:"personality_config"`
	RequiredPlan      string   `json:"required_plan"`
	KBTemplateNames   []string `json:"kb_template_names"` // 关联的知识库模板名称
}

// BundleKBTemplate 套餐内知识库模板
type BundleKBTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
