//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"ai-platform/billing-service/models"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(postgres.Open(seedTestDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("连接失败:", err)
	}

	plans := []models.Plan{
		{
			ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:        "免费版",
			Description: "适合个人体验",
			Price:       0,
			Period:      "month",
			MaxAgents:   1,
			MaxMessages: 100,
			Features:    `["1个智能体","100条消息/月","基础功能"]`,
			IsActive:    true,
		},
		{
			ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "专业版",
			Description: "适合小型团队",
			Price:       299,
			Period:      "month",
			MaxAgents:   5,
			MaxMessages: 10000,
			Features:    `["5个智能体","10000条消息/月","高级功能","优先支持","API接入"]`,
			IsActive:    true,
		},
		{
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Name:        "企业版",
			Description: "适合大型企业",
			Price:       999,
			Period:      "month",
			MaxAgents:   999,
			MaxMessages: 999999,
			Features:    `["无限智能体","无限消息","全部功能","专属支持","私有部署","SLA保障"]`,
			IsActive:    true,
		},
	}

	for _, p := range plans {
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
		result := db.Where("id = ?", p.ID).FirstOrCreate(&p)
		if result.Error != nil {
			fmt.Println("插入失败:", p.Name, result.Error)
		} else {
			fmt.Println("已插入/存在:", p.Name)
		}
	}
}

// seedTestDSN 从环境变量组装数据库 DSN（禁止在源码中硬编码口令）。
func seedTestDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		seedEnvOr("DB_HOST", "127.0.0.1"),
		seedEnvOr("DB_PORT", "5432"),
		seedEnvOr("DB_USER", "ai_platform"),
		os.Getenv("DB_PASSWORD"),
		seedEnvOr("DB_NAME", "ai_platform"))
}

func seedEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
