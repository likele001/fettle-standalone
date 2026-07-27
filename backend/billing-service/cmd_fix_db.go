//go:build ignore

package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(postgres.Open("host=127.0.0.1 port=5432 user=ai_platform password=CHANGE_ME dbname=ai_platform sslmode=disable"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("连接失败:", err)
	}

	// 删除旧表
	db.Exec("DROP TABLE IF EXISTS billing_records CASCADE")
	db.Exec("DROP TABLE IF EXISTS subscriptions CASCADE")
	db.Exec("DROP TABLE IF EXISTS plans CASCADE")
	fmt.Println("表已删除")

	// 重新迁移
	err = db.AutoMigrate(
		&struct {
			ID          string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
			Name        string  `gorm:"size:50;not null;uniqueIndex"`
			Description string  `gorm:"size:500"`
			Price       float64 `gorm:"type:decimal(10,2);not null"`
			Period      string  `gorm:"size:20;not null"`
			MaxAgents   int
			MaxMessages int64
			Features    string `gorm:"type:jsonb;default:'[]'"`
			IsActive    bool   `gorm:"default:true"`
		}{},
		&struct {
			ID           string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
			TenantID     string `gorm:"type:uuid;not null;index:idx_sub_tenant,unique"`
			PlanID       string `gorm:"type:uuid;not null"`
			Status       string `gorm:"size:20;not null;default:'active'"`
			MessagesUsed int64  `gorm:"default:0"`
		}{},
		&struct {
			ID             string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
			TenantID       string  `gorm:"type:uuid;not null;index"`
			SubscriptionID string  `gorm:"type:uuid;not null"`
			Type           string  `gorm:"size:20;not null"`
			Amount         int64
			Cost           float64 `gorm:"type:decimal(10,4);not null"`
		}{},
	)
	if err != nil {
		log.Fatal("迁移失败:", err)
	}
	fmt.Println("迁移完成")

	// 插入默认套餐
	db.Exec("INSERT INTO plans (id, name, description, price, period, max_agents, max_messages, features, is_active) VALUES ('11111111-1111-1111-1111-111111111111', '免费版', '适合个人体验', 0, 'month', 1, 100, '[\"1个智能体\",\"100条消息/月\",\"基础功能\"]', true) ON CONFLICT DO NOTHING")
	db.Exec("INSERT INTO plans (id, name, description, price, period, max_agents, max_messages, features, is_active) VALUES ('22222222-2222-2222-2222-222222222222', '专业版', '适合小型团队', 299, 'month', 5, 10000, '[\"5个智能体\",\"10000条消息/月\",\"高级功能\",\"优先支持\",\"API接入\"]', true) ON CONFLICT DO NOTHING")
	db.Exec("INSERT INTO plans (id, name, description, price, period, max_agents, max_messages, features, is_active) VALUES ('33333333-3333-3333-3333-333333333333', '企业版', '适合大型企业', 999, 'month', 999, 999999, '[\"无限智能体\",\"无限消息\",\"全部功能\",\"专属支持\",\"私有部署\",\"SLA保障\"]', true) ON CONFLICT DO NOTHING")
	fmt.Println("套餐已插入")
}
