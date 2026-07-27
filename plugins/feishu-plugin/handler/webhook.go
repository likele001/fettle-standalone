package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"ai-platform/feishu-plugin/client"
	"ai-platform/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type ChallengeRequest struct {
	Challenge string `json:"challenge"`
	Token     string `json:"token"`
	Type      string `json:"type"`
}

type EventRequest struct {
	Schema string `json:"schema"`
	Header struct {
		EventID   string `json:"event_id"`
		EventType string `json:"event_type"`
	} `json:"header"`
	Event struct {
		Sender   Sender  `json:"sender"`
		Message  Message `json:"message"`
	} `json:"event"`
}

type Sender struct {
	SenderID   SenderID `json:"sender_id"`
	SenderType string   `json:"sender_type"`
}

type SenderID struct {
	OpenID string `json:"open_id"`
	UserID string `json:"user_id"`
}

type Message struct {
	Content string `json:"content"`
	MsgType string `json:"msg_type"`
}

type WebhookHandler struct {
	nc     *nats.Conn
	client *client.FeishuClient
}

func NewWebhookHandler(nc *nats.Conn, cl *client.FeishuClient) *WebhookHandler {
	return &WebhookHandler{nc: nc, client: cl}
}

func (h *WebhookHandler) EventCallback(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "read body failed"})
		return
	}

	var challenge ChallengeRequest
	if err := json.Unmarshal(body, &challenge); err == nil && challenge.Type == "url_verification" {
		c.JSON(http.StatusOK, gin.H{"challenge": challenge.Challenge})
		return
	}

	var event EventRequest
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("Feishu parse event failed", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	if event.Header.EventType == "im.message.receive_v1" {
		openID := event.Event.Sender.SenderID.OpenID
		contentStr := event.Event.Message.Content

		var contentData map[string]string
		if err := json.Unmarshal([]byte(contentStr), &contentData); err != nil {
			logger.Error("Feishu parse content failed", zap.Error(err))
			c.JSON(http.StatusOK, gin.H{"code": 0})
			return
		}
		text := contentData["text"]

		logger.Info("Feishu message received", zap.String("from", openID), zap.String("content", text))

		if h.nc != nil && text != "" {
			eventData, _ := json.Marshal(map[string]interface{}{
				"type":      "message.new",
				"channel":   "feishu",
				"timestamp": time.Now().Unix(),
				"data": map[string]interface{}{
					"from_user": openID,
					"content":   text,
				},
			})
			h.nc.Publish("chat.inbound.>", eventData)
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0})
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
	if event.Channel != "feishu" || h.client == nil {
		return
	}

	toUser, _ := event.Data["to_user"].(string)
	content, _ := event.Data["content"].(string)
	if toUser == "" || content == "" {
		return
	}

	if err := h.client.SendText(toUser, content); err != nil {
		logger.Error("Feishu send message failed", zap.Error(err))
	}
}
