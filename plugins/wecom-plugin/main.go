package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"ai-platform/shared/logger"
	"ai-platform/wecom-plugin/client"
	"ai-platform/wecom-plugin/crypt"
	"ai-platform/wecom-plugin/handler"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	port := getEnv("APP_PORT", "20022")
	natsURL := getEnv("NATS_URL", "nats://localhost:20013")
	corpID := getEnv("WECOM_CORP_ID", "")
	corpSecret := getEnv("WECOM_CORP_SECRET", "")
	agentID, _ := strconv.Atoi(getEnv("WECOM_AGENT_ID", "0"))
	token := getEnv("WECOM_TOKEN", "")
	encodingAESKey := getEnv("WECOM_ENCODING_AES_KEY", "")

	logger.InitLogger("info")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", zap.Error(err))
		os.Exit(1)
	}
	defer nc.Close()

	var wecomClient *client.WeComClient
	if corpID != "" && corpSecret != "" && agentID > 0 {
		wecomClient = client.NewWeComClient(corpID, corpSecret, agentID)
	}

	var cryptor *crypt.WXBizMsgCrypt
	if token != "" && encodingAESKey != "" && corpID != "" {
		cryptor = crypt.NewWXBizMsgCrypt(token, encodingAESKey, corpID)
	}

	webhookHandler := handler.NewWebhookHandler(nc, cryptor, wecomClient)

	// 订阅 NATS outbound 消息（来自 chat-service 的回复）
	if wecomClient != nil {
		nc.Subscribe("chat.outbound.>", webhookHandler.HandleNATSMsg)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "wecom-plugin"})
	})

	wecom := r.Group("/wecom")
	{
		wecom.GET("/callback", webhookHandler.Verify)
		wecom.POST("/callback", webhookHandler.ReceiveMessage)
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("wecom-plugin starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("wecom-plugin shutting down")
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
