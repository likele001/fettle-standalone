package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"ai-platform/dingtalk-plugin/client"
	"ai-platform/dingtalk-plugin/handler"
	"ai-platform/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	port := getEnv("APP_PORT", "20025")
	natsURL := getEnv("NATS_URL", "nats://localhost:20013")
	appKey := getEnv("DINGTALK_APP_KEY", "")
	appSecret := getEnv("DINGTALK_APP_SECRET", "")
	agentID, _ := strconv.ParseInt(getEnv("DINGTALK_AGENT_ID", "0"), 10, 64)

	logger.InitLogger("info")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", zap.Error(err))
		os.Exit(1)
	}
	defer nc.Close()

	var dingtalkClient *client.DingTalkClient
	if appKey != "" && appSecret != "" {
		dingtalkClient = client.NewDingTalkClient(appKey, appSecret)
	}

	webhookHandler := handler.NewWebhookHandler(nc, dingtalkClient, agentID)

	if dingtalkClient != nil {
		nc.Subscribe("chat.outbound.>", webhookHandler.HandleNATSMsg)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "dingtalk-plugin"})
	})

	dingtalk := r.Group("/dingtalk")
	{
		dingtalk.POST("/callback", webhookHandler.ReceiveMessage)
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("dingtalk-plugin starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("dingtalk-plugin shutting down")
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
