package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-platform/agent-service/grpc_client"
	"ai-platform/agent-service/models"
	"ai-platform/agent-service/router"
	"ai-platform/agent-service/service"
	"ai-platform/agent-service/storage"
	"ai-platform/shared/config"
	"ai-platform/shared/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var gitVersion = "unknown"

func main() {
	logLevel := config.GetEnv("LOG_LEVEL", "info")
	if _, err := logger.InitLogger(logLevel); err != nil {
		panic(err)
	}

	port := config.GetEnvInt("AGENT_SERVICE_PORT", 9300)
	dbHost := config.GetEnv("DB_HOST", "localhost")
	dbPort := config.GetEnvInt("DB_PORT", 5432)
	dbUser := config.GetEnv("DB_USER", "ai_platform")
	dbPassword := config.GetEnv("DB_PASSWORD", "")
	if dbPassword == "" {
		logger.Fatal("DB_PASSWORD environment variable is required")
	}
	dbName := config.GetEnv("DB_NAME", "ai_platform")
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required")
	}
	aiEngineAddr := config.GetEnv("AI_ENGINE_GRPC_ADDR", "localhost:9701")
	// ai-engine 的 HTTP 地址，用于触发工作流子流程（主管 Agent → 子流程桥接）
	aiEngineURL := config.GetEnv("AI_ENGINE_URL", "http://localhost:9700")
	env := config.GetEnv("APP_ENV", "development")

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}

	if err := models.AutoMigrate(db); err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	// 种子智能体数据
	if err := models.SeedAgents(db); err != nil {
		logger.Error("failed to seed agents", zap.Error(err))
	}

	aiClient := grpc_client.NewAIEngineClient(aiEngineAddr, 30*time.Second)
	if err := aiClient.Connect(); err != nil {
		logger.Warn("failed to connect AI Engine gRPC", zap.String("addr", aiEngineAddr), zap.Error(err))
	} else {
		defer aiClient.Close()
		logger.Info("connected to AI Engine gRPC", zap.String("addr", aiEngineAddr))
	}

	var minioClient *storage.MinioClient
	minioEndpoint := config.GetEnv("MINIO_ENDPOINT", "")
	if minioEndpoint != "" && config.GetEnv("MINIO_SECRET_KEY", "") == "" {
		logger.Fatal("MINIO_SECRET_KEY environment variable is required when MINIO_ENDPOINT is set")
	}
	if minioEndpoint != "" {
		mc, err := storage.NewMinioClient(&storage.MinioConfig{
			Endpoint:  minioEndpoint,
			AccessKey: config.GetEnv("MINIO_ACCESS_KEY", "ai_platform"),
			SecretKey: config.GetEnv("MINIO_SECRET_KEY", ""),
			Bucket:    config.GetEnv("MINIO_BUCKET", "documents"),
			UseSSL:    config.GetEnv("MINIO_USE_SSL", "false") == "true",
		})
		if err != nil {
			logger.Warn("failed to connect MinIO, using local storage", zap.Error(err))
		} else {
			minioClient = mc
			logger.Info("connected to MinIO", zap.String("endpoint", minioEndpoint))
		}
	}

	// 初始化工作流 HTTP Client（用于主管 Agent 触发子流程）
	workflowClient := service.NewWorkflowHTTPClient(aiEngineURL)
	logger.Info("workflow HTTP client initialized", zap.String("url", aiEngineURL))

	r := router.NewRouter(db, jwtSecret, aiEngineAddr, aiClient, minioClient, workflowClient)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		logger.Info("agent-service starting", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("agent-service listen failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("agent-service shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("agent-service shutdown error", zap.Error(err))
	}
	logger.Info("agent-service stopped")
}
