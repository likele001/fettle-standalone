package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-platform/chat-service/models"
	"ai-platform/chat-service/router"
	"ai-platform/shared/config"
	"ai-platform/shared/logger"

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

	port := config.GetEnvInt("APP_PORT", 20004)
	dbHost := config.GetEnv("DB_HOST", "localhost")
	dbPort := config.GetEnvInt("DB_PORT", 5432)
	dbUser := config.GetEnv("DB_USER", "ai_platform")
	dbPassword := config.GetEnv("DB_PASSWORD", "ai_platform_secret")
	dbName := config.GetEnv("DB_NAME", "ai_platform")
	aiEngineURL := config.GetEnv("AI_ENGINE_URL", "http://localhost:9107")
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required")
	}
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
	logger.Info("database migrated successfully")

	r := router.NewRouter(db, jwtSecret, aiEngineURL)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		logger.Info("chat-service starting", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("chat-service listen failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("chat-service shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("chat-service shutdown error", zap.Error(err))
	}

	logger.Info("chat-service stopped")
}
