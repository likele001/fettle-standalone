package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"ai-platform/dingtalk-plugin/client"
	"ai-platform/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type CallbackRequest struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	SenderID       string `json:"senderId"`
	SenderNick     string `json:"senderNick"`
	ConversationID string `json:"conversationId"`
	ChatbotUserID  string `json:"chatbotUserId"`
	MsgID          string `json:"msgId"`
	CreateAt       int64  `json:"createAt"`
}

type WebhookHandler struct {
	nc      *nats.Conn
	client  *client.DingTalkClient
	agentID int64
}

func NewWebhookHandler(nc *nats.Conn, cl *client.DingTalkClient, agentID int64) *WebhookHandler {
	return &WebhookHandler{nc: nc, client: cl, agentID: agentID}
}

func (h *WebhookHandler) ReceiveMessage(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"errcode": 0, "errmsg": "ok"})
		return
	}

	var req CallbackRequest
	if err := json.Unmarshal(body, &req); err != nil {
		logger.Error("DingTalk parse failed", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"errcode": 0, "errmsg": "ok"})
		return
	}

	logger.Info("DingTalk message received", zap.String("from", req.SenderNick), zap.String("content", req.Text.Content))

	if req.MsgType == "text" && h.nc != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"type":      "message.new",
			"channel":   "dingtalk",
			"timestamp": time.Now().Unix(),
			"data": map[string]interface{}{
				"from_user":    req.SenderID,
				"sender_name":  req.SenderNick,
				"content":      req.Text.Content,
				"msg_id":       req.MsgID,
				"conversation": req.ConversationID,
			},
		})
		h.nc.Publish("chat.inbound.>", eventData)
	}

	c.JSON(http.StatusOK, gin.H{"errcode": 0, "errmsg": "ok"})
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
	if event.Channel != "dingtalk" || h.client == nil {
		return
	}

	toUser, _ := event.Data["to_user"].(string)
	content, _ := event.Data["content"].(string)
	if toUser == "" || content == "" {
		return
	}

	if err := h.client.SendText(h.agentID, toUser, content); err != nil {
		logger.Error("DingTalk send message failed", zap.Error(err))
	}
}
