package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"ai-platform/shared/logger"
	"ai-platform/wechat-plugin/service"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type TextMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
	MsgID        int64    `xml:"MsgId"`
}

type ReplyMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
}

type WebhookHandler struct {
	token string
	nc    *nats.Conn
	api   *service.WechatAPI
}

func NewWebhookHandler(token string, nc *nats.Conn, api *service.WechatAPI) *WebhookHandler {
	return &WebhookHandler{token: token, nc: nc, api: api}
}

func (h *WebhookHandler) Verify(c *gin.Context) {
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	if h.checkSignature(signature, timestamp, nonce) {
		c.String(http.StatusOK, echostr)
		return
	}

	c.String(http.StatusForbidden, "invalid signature")
}

func (h *WebhookHandler) ReceiveMessage(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusOK, "")
		return
	}

	var msg TextMessage
	if err := xml.Unmarshal(body, &msg); err != nil {
		logger.Error("Failed to parse XML message", zap.Error(err))
		c.String(http.StatusOK, "")
		return
	}

	logger.Info("WeChat message received", zap.String("type", msg.MsgType), zap.String("from", msg.FromUserName), zap.String("content", msg.Content))

	if h.nc != nil {
		event := map[string]interface{}{
			"type":      "message.new",
			"channel":   "wechat",
			"timestamp": time.Now().Unix(),
			"data": map[string]interface{}{
				"from_user":    msg.FromUserName,
				"to_user":      msg.ToUserName,
				"msg_type":     msg.MsgType,
				"content":      msg.Content,
				"msg_id":       msg.MsgID,
				"create_time":  msg.CreateTime,
			},
		}

		eventData, err := json.Marshal(event)
		if err == nil {
			h.nc.Publish("chat.inbound.>", eventData)
		}
	}

	c.String(http.StatusOK, "")
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
	if event.Channel != "wechat" || h.api == nil {
		return
	}

	toUser, _ := event.Data["to_user"].(string)
	content, _ := event.Data["content"].(string)
	if toUser == "" || content == "" {
		return
	}

	if err := h.api.SendTextMessage(toUser, content); err != nil {
		logger.Error("WeChat send message failed", zap.Error(err))
	}
}

func (h *WebhookHandler) checkSignature(signature, timestamp, nonce string) bool {
	arr := []string{h.token, timestamp, nonce}
	sort.Strings(arr)
	str := strings.Join(arr, "")

	hash := sha1.New()
	hash.Write([]byte(str))
	expected := hex.EncodeToString(hash.Sum(nil))

	return signature == expected
}
