package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-platform/shared/cache"
	"ai-platform/shared/config"
	"ai-platform/shared/logger"
	"ai-platform/user-service/models"
	"ai-platform/user-service/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logLevel := config.GetEnv("LOG_LEVEL", "info")
	if _, err := logger.InitLogger(logLevel); err != nil {
		panic(err)
	}

	port := config.GetEnvInt("USER_SERVICE_PORT", 9200)
	dbHost := config.GetEnv("DB_HOST", "localhost")
	dbPort := config.GetEnvInt("DB_PORT", 5432)
	dbUser := config.GetEnv("DB_USER", "ai_platform")
	dbPassword := config.GetEnv("DB_PASSWORD", "")
	if dbPassword == "" {
		logger.Fatal("DB_PASSWORD environment variable is required")
	}
	dbName := config.GetEnv("DB_NAME", "ai_platform")
	redisAddr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := config.GetEnv("REDIS_PASSWORD", "")
	redisDB := config.GetEnvInt("REDIS_DB", 9)
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required")
	}
	env := config.GetEnv("APP_ENV", "development")

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 数据库连接
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// 生产环境关闭自动迁移，使用手动迁移脚本
	})
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}

	// 自动迁移
	if err := models.AutoMigrate(db); err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	// 初始化数据
	if err := models.SeedData(db); err != nil {
		logger.Fatal("failed to seed database", zap.Error(err))
	}

	// 种子 AI 模型数据
	if err := models.SeedAIModels(db); err != nil {
		logger.Fatal("failed to seed AI models", zap.Error(err))
	}

	// Redis
	redisClient, err := cache.NewRedisClient(redisAddr, redisPassword, redisDB)
	if err != nil {
		logger.Fatal("failed to connect redis", zap.Error(err))
	}
	defer redisClient.Close()

	// 创建路由
	r := router.NewRouter(db, redisClient, jwtSecret)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		logger.Info("user-service starting", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("user-service listen failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("user-service shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("user-service shutdown error", zap.Error(err))
	}

	logger.Info("user-service stopped")
}
