package models

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&Tenant{},
		&User{},
		&AdminUser{},
		&Role{},
		&Permission{},
		&RolePermission{},
		&SysConfig{},
		// AI 配置相关表
		&AIProvider{},
		&AIModel{},
		&TenantAIConfig{},
		&TenantAPIKey{},
		&AIUsageLog{},
	); err != nil {
		return err
	}

	// 手动迁移：为已有表添加 username 字段
	// users 表
	if !db.Migrator().HasColumn(&User{}, "username") {
		if err := db.Migrator().AddColumn(&User{}, "username"); err != nil {
			log.Printf("[Migrate] 添加 users.username 字段失败: %v", err)
		} else {
			log.Println("[Migrate] 已添加 users.username 字段")
		}
	}

	// admin_users 表
	if !db.Migrator().HasColumn(&AdminUser{}, "username") {
		if err := db.Migrator().AddColumn(&AdminUser{}, "username"); err != nil {
			log.Printf("[Migrate] 添加 admin_users.username 字段失败: %v", err)
		} else {
			log.Println("[Migrate] 已添加 admin_users.username 字段")
		}
	}

	return nil
}

// SeedData 初始化基础数据
func SeedData(db *gorm.DB) error {
	// 1. 初始化系统权限
	permissions := []Permission{
		{Code: "user:read", Name: "查看用户", Module: "user"},
		{Code: "user:write", Name: "管理用户", Module: "user"},
		{Code: "agent:read", Name: "查看智能体", Module: "agent"},
		{Code: "agent:write", Name: "管理智能体", Module: "agent"},
		{Code: "chat:read", Name: "查看会话", Module: "chat"},
		{Code: "chat:write", Name: "管理会话", Module: "chat"},
		{Code: "skill:read", Name: "查看技能", Module: "skill"},
		{Code: "skill:write", Name: "管理技能", Module: "skill"},
		{Code: "billing:read", Name: "查看账单", Module: "billing"},
		{Code: "billing:write", Name: "管理账单", Module: "billing"},
		{Code: "tenant:read", Name: "查看租户", Module: "tenant"},
		{Code: "tenant:write", Name: "管理租户", Module: "tenant"},
		{Code: "plan:read", Name: "查看套餐", Module: "plan"},
		{Code: "plan:write", Name: "管理套餐", Module: "plan"},
		{Code: "system:read", Name: "查看系统配置", Module: "system"},
		{Code: "system:write", Name: "管理系统配置", Module: "system"},
		{Code: "admin:read", Name: "查看管理员", Module: "admin"},
		{Code: "admin:write", Name: "管理管理员", Module: "admin"},
	}

	for _, perm := range permissions {
		var existing Permission
		if err := db.Where("code = ?", perm.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&perm).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// 2. 创建平台租户（超管所属的特殊租户）
	platformTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	var platformTenant Tenant
	if err := db.Where("id = ?", platformTenantID).First(&platformTenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			platformTenant = Tenant{
				ID:       platformTenantID,
				Name:     "平台管理",
				PlanType: "enterprise",
				Status:   "active",
				Config:   JSONMap{"is_platform": true},
			}
			if err := db.Create(&platformTenant).Error; err != nil {
				return fmt.Errorf("创建平台租户失败: %w", err)
			}
			log.Println("[Seed] 平台租户已创建")
		} else {
			return err
		}
	}

	// 3. 创建超级管理员（写入 admin_users 表，独立于租户用户）
	var superAdmin AdminUser
	if err := db.Where("role = ?", "super_admin").First(&superAdmin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 生成 bcrypt 密码哈希
			defaultPassword := "CHANGE_ME_ADMIN_PASSWORD"
			hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("生成密码哈希失败: %w", err)
			}

			superAdmin = AdminUser{
				Username:     "admin",
				Phone:        "13800000000",
				Email:        "admin@fettle.com",
				PasswordHash: string(hash),
				Name:         "超级管理员",
				Role:         "super_admin",
				Status:       "active",
			}
			if err := db.Create(&superAdmin).Error; err != nil {
				return fmt.Errorf("创建超级管理员失败: %w", err)
			}
			log.Println("[Seed] 超级管理员已创建 (用户名: admin, 手机号: 13800000000, 密码: CHANGE_ME_ADMIN_PASSWORD)")
		} else {
			return err
		}
	} else {
		// 如果超管已存在但没有 username，补充设置
		if superAdmin.Username == "" {
			db.Model(&superAdmin).Update("username", "admin")
			log.Println("[Seed] 已为超级管理员补充 username: admin")
		}
	}

	return nil
}

// seedModel 种子模型定义
type seedModel struct {
	ProviderCode   string
	ModelCode      string
	ModelName      string
	ModelType      string
	MaxInputTokens int
	MaxOutputTokens int
	InputPrice     float64
	OutputPrice    float64
	Capabilities   []string
	IsDefault      bool
	Priority       int
}

// SeedAIModels 种子 AI 模型数据（幂等：按 provider_id+model_code 去重）
func SeedAIModels(db *gorm.DB) error {
	models := []seedModel{
		// === 阿里云 ===
		// chat
		{"aliyun", "qwen-turbo", "通义千问-Turbo", "chat", 4096, 2048, 0.002, 0.006, []string{"chat", "code"}, false, 1},
		{"aliyun", "qwen-plus", "通义千问-Plus", "chat", 32768, 2048, 0.004, 0.012, []string{"chat", "code", "reasoning"}, false, 2},
		{"aliyun", "qwen-max", "通义千问-Max", "chat", 32768, 2048, 0.02, 0.06, []string{"chat", "vision", "function_call", "code", "reasoning"}, true, 0},
		{"aliyun", "qwen-max-longcontext", "通义千问-Max-长文本", "chat", 28000, 2048, 0.02, 0.06, []string{"chat", "long_context"}, false, 3},
		// embedding
		{"aliyun", "text-embedding-v2", "通义文本向量-v2", "embedding", 2048, 0, 0.0007, 0, []string{"embedding"}, true, 0},
		// image
		{"aliyun", "wanx-v1", "通义万相-文生图", "image", 0, 0, 0.02, 0, []string{"image_generation"}, false, 10},
		{"aliyun", "wanx-2.1-imageedit", "通义万相-图像编辑", "image", 0, 0, 0.02, 0, []string{"image_edit"}, false, 11},
		// tts
		{"aliyun", "cosyvoice-v1", "CosyVoice语音合成", "tts", 0, 0, 0.002, 0, []string{"tts"}, false, 20},
		// vision
		{"aliyun", "qwen-vl-max", "Qwen-VL-Max", "vision", 32768, 2048, 0.003, 0.009, []string{"chat", "vision"}, false, 5},
		{"aliyun", "qwen-vl-plus", "Qwen-VL-Plus", "vision", 8192, 2048, 0.0015, 0.005, []string{"chat", "vision"}, false, 6},
		// stt
		{"aliyun", "paraformer-realtime-v2", "Paraformer实时语音识别", "stt", 0, 0, 0.001, 0, []string{"stt"}, false, 21},

		// === 百度 ===
		{"baidu", "ernie-bot-4", "文心一言 4.0", "chat", 8192, 2048, 0.03, 0.06, []string{"chat", "reasoning"}, true, 0},
		{"baidu", "ernie-bot-3.5", "文心一言 3.5", "chat", 4096, 2048, 0.004, 0.008, []string{"chat"}, false, 1},
		{"baidu", "ernie-bot-turbo", "文心一言 Turbo", "chat", 4096, 2048, 0.001, 0.002, []string{"chat"}, false, 2},
		{"baidu", "ernie-vilg-v2", "文心一格", "image", 0, 0, 0.02, 0, []string{"image_generation"}, false, 10},
		{"baidu", "ernie-vi", "文心千帆视觉理解", "vision", 8192, 2048, 0.003, 0.009, []string{"chat", "vision"}, false, 5},

		// === 腾讯 ===
		{"tencent", "hunyuan-lite", "混元 Lite", "chat", 4096, 2048, 0.001, 0.002, []string{"chat"}, false, 2},
		{"tencent", "hunyuan-standard", "混元 Standard", "chat", 8192, 2048, 0.004, 0.008, []string{"chat", "function_call"}, true, 0},
		{"tencent", "hunyuan-pro", "混元 Pro", "chat", 8192, 2048, 0.01, 0.02, []string{"chat", "function_call", "reasoning"}, false, 1},
		{"tencent", "hunyuan-image", "混元图像生成", "image", 0, 0, 0.02, 0, []string{"image_generation"}, false, 10},
		{"tencent", "hunyuan-vision", "混元视觉理解", "vision", 8192, 2048, 0.003, 0.009, []string{"chat", "vision"}, false, 5},
		{"tencent", "hunyuan-video", "混元视频生成", "video", 0, 0, 0.05, 0, []string{"video_generation"}, false, 15},

		// === OpenAI ===
		{"openai", "gpt-4o", "GPT-4o", "chat", 128000, 4096, 0.005, 0.015, []string{"chat", "vision", "function_call", "code"}, true, 0},
		{"openai", "gpt-4o-mini", "GPT-4o Mini", "chat", 128000, 4096, 0.00015, 0.0006, []string{"chat", "vision", "function_call"}, false, 1},
		{"openai", "gpt-4-turbo", "GPT-4 Turbo", "chat", 128000, 4096, 0.01, 0.03, []string{"chat", "vision", "function_call", "code"}, false, 2},
		{"openai", "gpt-3.5-turbo", "GPT-3.5 Turbo", "chat", 16385, 4096, 0.0005, 0.0015, []string{"chat", "function_call"}, false, 3},
		{"openai", "text-embedding-3-large", "Embedding Large", "embedding", 8191, 0, 0.00013, 0, []string{"embedding"}, true, 0},
		{"openai", "text-embedding-3-small", "Embedding Small", "embedding", 8191, 0, 0.00002, 0, []string{"embedding"}, false, 1},
		{"openai", "dall-e-3", "DALL-E 3", "image", 0, 0, 0.04, 0, []string{"image_generation"}, false, 10},
		{"openai", "whisper-1", "Whisper语音识别", "stt", 0, 0, 0.006, 0, []string{"stt"}, false, 20},
		{"openai", "tts-1", "TTS-1语音合成", "tts", 0, 0, 0.015, 0, []string{"tts"}, false, 21},
		{"openai", "sora", "Sora视频生成", "video", 0, 0, 0.10, 0, []string{"video_generation"}, false, 15},

		// === Anthropic ===
		{"anthropic", "claude-3-opus", "Claude 3 Opus", "chat", 200000, 4096, 0.015, 0.075, []string{"chat", "vision", "code", "reasoning"}, false, 1},
		{"anthropic", "claude-3-sonnet", "Claude 3 Sonnet", "chat", 200000, 4096, 0.003, 0.015, []string{"chat", "vision", "function_call", "code"}, true, 0},
		{"anthropic", "claude-3-haiku", "Claude 3 Haiku", "chat", 200000, 4096, 0.00025, 0.00125, []string{"chat", "vision"}, false, 2},
		{"anthropic", "claude-3.5-sonnet", "Claude 3.5 Sonnet", "chat", 200000, 8192, 0.003, 0.015, []string{"chat", "vision", "function_call", "code", "reasoning"}, false, 0},
		{"anthropic", "claude-3-5-sonnet-vision", "Claude 3.5 Sonnet (Vision)", "vision", 200000, 8192, 0.003, 0.015, []string{"chat", "vision"}, false, 5},
		{"anthropic", "claude-3-opus-vision", "Claude 3 Opus (Vision)", "vision", 200000, 4096, 0.015, 0.075, []string{"chat", "vision"}, false, 6},
		{"anthropic", "claude-3-haiku-vision", "Claude 3 Haiku (Vision)", "vision", 200000, 4096, 0.00025, 0.00125, []string{"chat", "vision"}, false, 7},

		// === Google ===
		{"google", "gemini-1.5-pro", "Gemini 1.5 Pro", "chat", 1000000, 8192, 0.0035, 0.0105, []string{"chat", "vision", "function_call", "code"}, true, 0},
		{"google", "gemini-1.5-flash", "Gemini 1.5 Flash", "chat", 1000000, 8192, 0.0008, 0.003, []string{"chat", "vision"}, false, 1},
		{"google", "gemini-1.5-pro-vision", "Gemini 1.5 Pro (Vision)", "vision", 1000000, 8192, 0.0035, 0.0105, []string{"chat", "vision"}, false, 5},
		{"google", "gemini-1.5-flash-vision", "Gemini 1.5 Flash (Vision)", "vision", 1000000, 8192, 0.0008, 0.003, []string{"chat", "vision"}, false, 6},
		{"google", "imagen-3", "Imagen 3", "image", 0, 0, 0.03, 0, []string{"image_generation"}, false, 10},

		// === MiniMax ===
		{"minimax", "abab6.5s-chat", "Abab6.5S Chat", "chat", 8192, 2048, 0.001, 0.002, []string{"chat"}, true, 0},
		{"minimax", "abab5.5-chat", "Abab5.5 Chat", "chat", 16384, 4096, 0.005, 0.015, []string{"chat", "function_call"}, false, 1},
		{"minimax", "embo-01", "Embo-01 向量", "embedding", 4096, 0, 0.0005, 0, []string{"embedding"}, true, 0},
		{"minimax", "speech-01", "Speech-01语音合成", "tts", 0, 0, 0.002, 0, []string{"tts"}, false, 20},
		{"minimax", "video-01", "Video-01视频生成", "video", 0, 0, 0.05, 0, []string{"video_generation"}, false, 15},

		// === 智谱AI ===
		{"zhipu", "glm-4", "GLM-4", "chat", 128000, 4096, 0.01, 0.03, []string{"chat", "function_call", "code"}, true, 0},
		{"zhipu", "glm-3-turbo", "GLM-3-Turbo", "chat", 128000, 4096, 0.001, 0.003, []string{"chat"}, false, 1},
		{"zhipu", "embedding-2", "Embedding-2", "embedding", 8192, 0, 0.0005, 0, []string{"embedding"}, true, 0},
		{"zhipu", "cogview-3", "CogView-3", "image", 0, 0, 0.02, 0, []string{"image_generation"}, false, 10},
		{"zhipu", "glm-4v", "GLM-4V (Vision)", "vision", 8192, 2048, 0.003, 0.009, []string{"chat", "vision"}, false, 5},
		{"zhipu", "cogvideox", "CogVideoX", "video", 0, 0, 0.05, 0, []string{"video_generation"}, false, 15},

		// === Moonshot ===
		{"moonshot", "moonshot-v1-8k", "Moonshot V1 8K", "chat", 8192, 4096, 0.0012, 0.0012, []string{"chat"}, false, 2},
		{"moonshot", "moonshot-v1-32k", "Moonshot V1 32K", "chat", 32768, 4096, 0.0024, 0.0024, []string{"chat"}, true, 0},
		{"moonshot", "moonshot-v1-128k", "Moonshot V1 128K", "chat", 131072, 4096, 0.004, 0.012, []string{"chat", "long_context"}, false, 1},
		{"moonshot", "moonshot-v1-128k-vision", "Moonshot V1 128K (Vision)", "vision", 128000, 4096, 0.004, 0.012, []string{"chat", "vision"}, false, 5},

		// === DeepSeek ===
		{"deepseek", "deepseek-chat", "DeepSeek Chat", "chat", 32000, 4096, 0.00014, 0.00028, []string{"chat", "code"}, true, 0},
		{"deepseek", "deepseek-coder", "DeepSeek Coder", "chat", 16000, 4096, 0.00014, 0.00028, []string{"chat", "code"}, false, 1},
		{"deepseek", "deepseek-vl", "DeepSeek-VL", "vision", 16000, 4096, 0.002, 0.006, []string{"chat", "vision"}, false, 5},
	}

	// 先查出所有 provider code → ID 映射
	var providers []AIProvider
	if err := db.Find(&providers).Error; err != nil {
		return fmt.Errorf("查询厂商列表失败: %w", err)
	}
	providerMap := make(map[string]uuid.UUID)
	for _, p := range providers {
		providerMap[p.Code] = p.ID
	}

	created := 0
	for _, sm := range models {
		providerID, ok := providerMap[sm.ProviderCode]
		if !ok {
			log.Printf("[SeedAI] 厂商 %s 不存在，跳过模型 %s", sm.ProviderCode, sm.ModelCode)
			continue
		}

		var existing AIModel
		err := db.Where("provider_id = ? AND model_code = ?", providerID, sm.ModelCode).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			m := AIModel{
				ProviderID:      providerID,
				ModelCode:       sm.ModelCode,
				ModelName:       sm.ModelName,
				ModelType:       sm.ModelType,
				MaxInputTokens:  sm.MaxInputTokens,
				MaxOutputTokens: sm.MaxOutputTokens,
				InputPricePer1k: sm.InputPrice,
				OutputPricePer1k: sm.OutputPrice,
				Capabilities:    sm.Capabilities,
				Status:          "active",
				IsDefault:       sm.IsDefault,
				Priority:        sm.Priority,
			}
			if err := db.Create(&m).Error; err != nil {
				log.Printf("[SeedAI] 创建模型 %s/%s 失败: %v", sm.ProviderCode, sm.ModelCode, err)
				continue
			}
			created++
		} else if err != nil {
			log.Printf("[SeedAI] 查询模型 %s/%s 失败: %v", sm.ProviderCode, sm.ModelCode, err)
		}
	}

	if created > 0 {
		log.Printf("[SeedAI] 新增 %d 个 AI 模型", created)
	}
	return nil
}
