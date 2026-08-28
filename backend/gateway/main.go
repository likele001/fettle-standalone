package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-platform/shared/cache"
	"ai-platform/shared/config"
	"ai-platform/shared/logger"
	"ai-platform/gateway/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logLevel := config.GetEnv("LOG_LEVEL", "info")
	if _, err := logger.InitLogger(logLevel); err != nil {
		panic(err)
	}

	// 环境变量
	port := config.GetEnvInt("GATEWAY_PORT", 9100)
	redisAddr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := config.GetEnv("REDIS_PASSWORD", "")
	redisDB := config.GetEnvInt("REDIS_DB", 8)
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required")
	}
	env := config.GetEnv("APP_ENV", "development")

	// 设置Gin模式
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化Redis
	redisClient, err := cache.NewRedisClient(redisAddr, redisPassword, redisDB)
	if err != nil {
		logger.Fatal("failed to connect redis", zap.Error(err))
	}
	defer redisClient.Close()

	// 创建路由
	r := router.NewRouter(redisClient, jwtSecret)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + config.GetEnv("GATEWAY_PORT", "9100"),
		Handler: r,
	}

	// 启动服务
	go func() {
		logger.Info("gateway starting", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("gateway listen failed", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("gateway shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("gateway shutdown error", zap.Error(err))
	}

	logger.Info("gateway stopped")
}
