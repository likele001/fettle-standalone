package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"ai-platform/douyin-plugin/client"
	"ai-platform/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type CallbackRequest struct {
	Event   string `json:"event"`
	ClientKey string `json:"client_key"`
	FromUser string `json:"from_user"`
	ToUser   string `json:"to_user"`
	Content  string `json:"content"`
	MsgID    string `json:"msg_id"`
}

type WebhookHandler struct {
	nc     *nats.Conn
	client *client.DouyinClient
}

func NewWebhookHandler(nc *nats.Conn, cl *client.DouyinClient) *WebhookHandler {
	return &WebhookHandler{nc: nc, client: cl}
}

func (h *WebhookHandler) Verify(c *gin.Context) {
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echo_str")
	signature := c.Query("signature")

	if h.client != nil && !h.client.VerifySignature(timestamp, nonce, echostr, signature) {
		logger.Warn("Douyin verify signature mismatch")
		c.String(http.StatusForbidden, "signature mismatch")
		return
	}

	c.String(http.StatusOK, echostr)
}

func (h *WebhookHandler) ReceiveMessage(c *gin.Context) {
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	signature := c.Query("signature")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"err_no": 0, "err_msg": "ok"})
		return
	}

	if h.client != nil && !h.client.VerifySignature(timestamp, nonce, string(body), signature) {
		logger.Warn("Douyin message signature mismatch")
		c.JSON(http.StatusOK, gin.H{"err_no": 0, "err_msg": "ok"})
		return
	}

	var msg CallbackRequest
	if err := json.Unmarshal(body, &msg); err != nil {
		logger.Error("Douyin parse message failed", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"err_no": 0, "err_msg": "ok"})
		return
	}

	logger.Info("Douyin message received", zap.String("from", msg.FromUser), zap.String("content", msg.Content))

	if h.nc != nil && msg.Content != "" {
		eventData, _ := json.Marshal(map[string]interface{}{
			"type":      "message.new",
			"channel":   "douyin",
			"timestamp": time.Now().Unix(),
			"data": map[string]interface{}{
				"from_user": msg.FromUser,
				"content":   msg.Content,
				"msg_id":    msg.MsgID,
			},
		})
		h.nc.Publish("chat.inbound.>", eventData)
	}

	c.JSON(http.StatusOK, gin.H{"err_no": 0, "err_msg": "ok"})
}

func (h *WebhookHandler) HandleNATSMsg(msg *nats.Msg) {
	var event struct {
		Type    string                 `json:"type"`
		Channel string                 `json:"channel"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return
	}
	if event.Channel != "douyin" || h.client == nil {
		return
	}

	toUser, _ := event.Data["to_user"].(string)
	content, _ := event.Data["content"].(string)
	if toUser == "" || content == "" {
		return
	}

	if err := h.client.SendText(toUser, content); err != nil {
		logger.Error("Douyin send message failed", zap.Error(err))
	}
}
