package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-platform/shared/config"
	"ai-platform/shared/logger"
	"ai-platform/billing-service/models"
	"ai-platform/billing-service/router"
	"ai-platform/billing-service/repository"
	"ai-platform/billing-service/service"

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

	port := config.GetEnvInt("BILLING_SERVICE_PORT", 9600)
	dbHost := config.GetEnv("DB_HOST", "localhost")
	dbPort := config.GetEnvInt("DB_PORT", 5432)
	dbUser := config.GetEnv("DB_USER", "ai_platform")
	dbPassword := config.GetEnv("DB_PASSWORD", "CHANGE_ME")
	dbName := config.GetEnv("DB_NAME", "ai_platform")
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

	if err := db.AutoMigrate(
		&models.Plan{},
		&models.Subscription{},
		&models.BillingRecord{},
		&models.PaymentConfig{},
		&models.PaymentOrder{},
		&models.PlatformModelPricing{},
		&models.TenantBalance{},
		&models.ResourcePackage{},
		&models.TenantResourcePackage{},
		&models.TenantUsageDetail{},
		&models.TenantRechargeRecord{},
	); err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	billingService := service.NewBillingService(db)
	exportService := service.NewExportService(repository.NewBillingRepository(db))
	paymentService := service.NewPaymentService(db)
	aiBillingService := service.NewAIBillingService(db)
	r := router.Setup(billingService, paymentService, exportService, aiBillingService, jwtSecret)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		logger.Info("billing-service starting", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("billing-service listen failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("billing-service shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("billing-service shutdown error", zap.Error(err))
	}
	logger.Info("billing-service stopped")
}
