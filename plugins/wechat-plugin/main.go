package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-platform/shared/logger"
	"ai-platform/wechat-plugin/handler"
	"ai-platform/wechat-plugin/service"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	port := getEnv("APP_PORT", "20021")
	wechatToken := getEnv("WECHAT_TOKEN", "")
	wechatAppID := getEnv("WECHAT_APP_ID", "")
	wechatAppSecret := getEnv("WECHAT_APP_SECRET", "")
	natsURL := getEnv("NATS_URL", "nats://localhost:20013")

	logger.InitLogger("info")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", zap.Error(err))
		os.Exit(1)
	}
	defer nc.Close()

	var wechatAPI *service.WechatAPI
	if wechatAppID != "" && wechatAppSecret != "" {
		wechatAPI = service.NewWechatAPI(wechatAppID, wechatAppSecret, wechatToken)
	}

	webhookHandler := handler.NewWebhookHandler(wechatToken, nc, wechatAPI)

	if wechatAPI != nil {
		nc.Subscribe("chat.outbound.>", webhookHandler.HandleNATSMsg)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "wechat-plugin",
		})
	})

	wechat := r.Group("/wechat")
	{
		wechat.GET("/webhook", webhookHandler.Verify)
		wechat.POST("/webhook", webhookHandler.ReceiveMessage)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		logger.Info("wechat-plugin starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("wechat-plugin shutting down")
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
