package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-platform/douyin-plugin/client"
	"ai-platform/douyin-plugin/handler"
	"ai-platform/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	port := getEnv("APP_PORT", "20023")
	natsURL := getEnv("NATS_URL", "nats://localhost:20013")
	appID := getEnv("DOUYIN_APP_ID", "")
	appSecret := getEnv("DOUYIN_APP_SECRET", "")

	logger.InitLogger("info")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", zap.Error(err))
		os.Exit(1)
	}
	defer nc.Close()

	var douyinClient *client.DouyinClient
	if appID != "" && appSecret != "" {
		douyinClient = client.NewDouyinClient(appID, appSecret)
	}

	webhookHandler := handler.NewWebhookHandler(nc, douyinClient)

	if douyinClient != nil {
		nc.Subscribe("chat.outbound.>", webhookHandler.HandleNATSMsg)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "douyin-plugin"})
	})

	douyin := r.Group("/douyin")
	{
		douyin.GET("/webhook", webhookHandler.Verify)
		douyin.POST("/webhook", webhookHandler.ReceiveMessage)
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("douyin-plugin starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("douyin-plugin shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
