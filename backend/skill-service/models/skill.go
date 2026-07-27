package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Skill 技能模型
type Skill struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name             string         `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description      string         `gorm:"size:500" json:"description"`
	Category         string         `gorm:"size:50;not null" json:"category"` // data, tool, content, other, workflow
	Icon             string         `gorm:"size:255" json:"icon"`
	Version          string         `gorm:"size:20;default:'1.0.0'" json:"version"`
	Author           string         `gorm:"size:100" json:"author"`
	IsPublic         bool           `gorm:"default:false" json:"is_public"`
	IsBuiltIn        bool           `gorm:"default:false" json:"is_built_in"`
	Code             string         `gorm:"type:text" json:"-"` // Python 代码实现
	Config           string         `gorm:"type:jsonb;default:'{}'" json:"config"`
	IsMCPTool        bool           `gorm:"default:false" json:"is_mcp_tool"`
	MCPToolName      string         `gorm:"size:100;default:''" json:"mcp_tool_name"`
	MCPDescription   string         `gorm:"size:500;default:''" json:"mcp_description"`
	MCPParameters    string         `gorm:"type:jsonb;default:'{}'" json:"mcp_parameters"`
	WorkflowTemplate string         `gorm:"type:text;default:''" json:"workflow_template"`
	InstallCount     int64          `gorm:"default:0" json:"install_count"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// SkillInstallation 技能安装记录
type SkillInstallation struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	SkillID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"skill_id"`
	Status    string         `gorm:"size:20;not null;default:'active'" json:"status"` // active, disabled
	Config    string         `gorm:"type:jsonb;default:'{}'" json:"config"`
	InstalledAt time.Time    `json:"installed_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// SkillMarketItem 技能市场项
type SkillMarketItem struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Icon        string    `json:"icon"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	InstallCount int64    `json:"install_count"`
	Installed   bool      `json:"installed"`
}
