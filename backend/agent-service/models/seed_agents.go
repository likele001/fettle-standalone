package models

import (
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedAgents 种子智能体数据（为平台租户预置常用智能体）
func SeedAgents(db *gorm.DB) error {
	platformTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// 检查是否已有智能体
	var count int64
	db.Model(&Agent{}).Where("tenant_id = ?", platformTenantID).Count(&count)
	if count > 0 {
		log.Printf("[SeedAgents] 平台租户已有 %d 个智能体，跳过种子", count)
		return nil
	}

	agents := []Agent{
		{
			TenantID:    platformTenantID,
			Name:        "智能客服小助",
			Description: "通用智能客服，支持多轮对话、问题解答和知识检索，可快速接入各类业务场景。",
			Status:      "idle",
			AgentType:   "chat",
			IsPublic:    true,
			RequiredPlan: "free",
			PersonalityConfig: JSONMap{
				"role":     "客服助手",
				"tone":     "友好专业",
				"language": "中文",
			},
			WelcomeMessage: "您好！我是智能客服小助，很高兴为您服务。请问有什么可以帮您的吗？",
			ModelID:        "default",
			MaxConcurrent:  10,
		},
		{
			TenantID:    platformTenantID,
			Name:        "代码助手",
			Description: "专业的编程助手，支持多种编程语言，可以帮助编写代码、调试问题、代码审查和技术咨询。",
			Status:      "idle",
			AgentType:   "chat",
			IsPublic:    true,
			RequiredPlan: "pro",
			PersonalityConfig: JSONMap{
				"role":     "编程专家",
				"tone":     "精确严谨",
				"language": "中文",
				"specialties": []string{"Python", "Go", "JavaScript", "SQL"},
			},
			WelcomeMessage: "你好！我是代码助手，可以帮你写代码、调试问题和解答技术疑问。请描述你的需求吧。",
			ModelID:        "default",
			MaxConcurrent:  10,
		},
		{
			TenantID:    platformTenantID,
			Name:        "文案写手",
			Description: "创意文案专家，擅长撰写营销文案、产品描述、社交媒体内容、公众号文章等，风格多变。",
			Status:      "idle",
			AgentType:   "chat",
			IsPublic:    true,
			RequiredPlan: "free",
			PersonalityConfig: JSONMap{
				"role":     "文案策划",
				"tone":     "创意生动",
				"language": "中文",
				"styles":  []string{"营销文案", "品牌故事", "社媒内容", "公众号文章"},
			},
			WelcomeMessage: "嗨！我是你的专属文案写手，无论是广告语、产品描述还是长篇推文，交给我吧！告诉我你想写什么？",
			ModelID:        "default",
			MaxConcurrent:  10,
		},
		{
			TenantID:    platformTenantID,
			Name:        "数据分析师",
			Description: "数据分析专家，能够解读数据、生成报告、提供洞察和建议，支持多种数据格式分析。",
			Status:      "idle",
			AgentType:   "chat",
			IsPublic:    true,
			RequiredPlan: "pro",
			PersonalityConfig: JSONMap{
				"role":     "数据分析师",
				"tone":     "理性客观",
				"language": "中文",
				"skills":  []string{"数据解读", "趋势分析", "报告生成", "可视化建议"},
			},
			WelcomeMessage: "你好！我是数据分析师，可以帮你分析数据、发现趋势、生成报告。请分享你的数据或分析需求吧。",
			ModelID:        "default",
			MaxConcurrent:  10,
		},
		{
			TenantID:    platformTenantID,
			Name:        "翻译专家",
			Description: "多语言翻译助手，支持中英日韩等多语种互译，保持原文语境和风格，适合商务和技术文档翻译。",
			Status:      "idle",
			AgentType:   "chat",
			IsPublic:    true,
			RequiredPlan: "pro",
			PersonalityConfig: JSONMap{
				"role":     "翻译专家",
				"tone":     "准确流畅",
				"languages": []string{"中文", "英文", "日文", "韩文"},
				"domains":  []string{"商务", "技术", "法律", "医学"},
			},
			WelcomeMessage: "你好！我是翻译专家，支持中英日韩等多语种翻译。请直接输入需要翻译的内容，并告诉我目标语言。",
			ModelID:        "default",
			MaxConcurrent:  10,
		},
		{
			TenantID:    platformTenantID,
			Name:        "会议纪要助手",
			Description: "自动整理会议内容，提取关键决策、待办事项和参会人发言要点，生成结构化会议纪要。",
			Status:      "idle",
			AgentType:   "chat",
			IsPublic:    true,
			RequiredPlan: "pro",
			PersonalityConfig: JSONMap{
				"role":     "会议秘书",
				"tone":     "简洁规范",
				"language": "中文",
				"output_format": []string{"决策事项", "待办清单", "讨论要点", "下次会议安排"},
			},
			WelcomeMessage: "你好！我是会议纪要助手。请将会议内容发给我，我会帮你整理出结构化的会议纪要，包括关键决策和待办事项。",
			ModelID:        "default",
			MaxConcurrent:  10,
		},
	}

	created := 0
	for _, agent := range agents {
		if err := db.Create(&agent).Error; err != nil {
			log.Printf("[SeedAgents] 创建智能体 %s 失败: %v", agent.Name, err)
			continue
		}
		created++
	}

	if created > 0 {
		log.Printf("[SeedAgents] 新增 %d 个预置智能体", created)
	}
	return nil
}
