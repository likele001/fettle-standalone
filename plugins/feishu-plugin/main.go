package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-platform/feishu-plugin/client"
	"ai-platform/feishu-plugin/handler"
	"ai-platform/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	port := getEnv("APP_PORT", "20024")
	natsURL := getEnv("NATS_URL", "nats://localhost:20013")
	appID := getEnv("FEISHU_APP_ID", "")
	appSecret := getEnv("FEISHU_APP_SECRET", "")

	logger.InitLogger("info")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", zap.Error(err))
		os.Exit(1)
	}
	defer nc.Close()

	var feishuClient *client.FeishuClient
	if appID != "" && appSecret != "" {
		feishuClient = client.NewFeishuClient(appID, appSecret)
	}

	webhookHandler := handler.NewWebhookHandler(nc, feishuClient)

	if feishuClient != nil {
		nc.Subscribe("chat.outbound.>", webhookHandler.HandleNATSMsg)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "feishu-plugin"})
	})

	feishu := r.Group("/feishu")
	{
		feishu.POST("/event", webhookHandler.EventCallback)
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("feishu-plugin starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("feishu-plugin shutting down")
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
