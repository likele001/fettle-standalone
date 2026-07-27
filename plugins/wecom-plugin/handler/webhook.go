package handler

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"time"

	"ai-platform/shared/logger"
	"ai-platform/wecom-plugin/client"
	"ai-platform/wecom-plugin/crypt"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type TextMessageXML struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime int64    `xml:"CreateTime"`
	MsgType    string   `xml:"MsgType"`
	Content    string   `xml:"Content"`
	MsgID      int64    `xml:"MsgId"`
	AgentID    int      `xml:"AgentID"`
}

type WebhookHandler struct {
	nc     *nats.Conn
	crypt  *crypt.WXBizMsgCrypt
	client *client.WeComClient
}

func NewWebhookHandler(nc *nats.Conn, c *crypt.WXBizMsgCrypt, cl *client.WeComClient) *WebhookHandler {
	return &WebhookHandler{nc: nc, crypt: c, client: cl}
}

func (h *WebhookHandler) Verify(c *gin.Context) {
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	plaintext, err := h.crypt.VerifyURL(msgSignature, timestamp, nonce, echostr)
	if err != nil {
		logger.Error("WeCom verify failed", zap.Error(err))
		c.String(http.StatusForbidden, "verify failed")
		return
	}
	c.String(http.StatusOK, plaintext)
}

func (h *WebhookHandler) ReceiveMessage(c *gin.Context) {
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusOK, "success")
		return
	}

	decrypted, err := h.crypt.DecryptMsg(msgSignature, timestamp, nonce, string(body))
	if err != nil {
		logger.Error("WeCom decrypt failed", zap.Error(err))
		c.String(http.StatusOK, "success")
		return
	}

	var msg TextMessageXML
	if err := xml.Unmarshal([]byte(decrypted), &msg); err != nil {
		logger.Error("WeCom parse msg failed", zap.Error(err))
		c.String(http.StatusOK, "success")
		return
	}

	logger.Info("WeCom message received", zap.String("from", msg.FromUserName), zap.String("content", msg.Content))

	if msg.MsgType == "text" && h.nc != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"type":      "message.new",
			"channel":   "wecom",
			"timestamp": time.Now().Unix(),
			"data": map[string]interface{}{
				"from_user":    msg.FromUserName,
				"to_user":      msg.ToUserName,
				"content":      msg.Content,
				"msg_id":       msg.MsgID,
				"agent_id":     msg.AgentID,
			},
		})
		h.nc.Publish("chat.inbound.>", eventData)
	}

	c.String(http.StatusOK, "success")
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
	if event.Channel != "wecom" || h.client == nil {
		return
	}

	toUser, _ := event.Data["to_user"].(string)
	content, _ := event.Data["content"].(string)
	if toUser == "" || content == "" {
		return
	}

	if err := h.client.SendText(toUser, content); err != nil {
		logger.Error("WeCom send message failed", zap.Error(err))
	}
}
