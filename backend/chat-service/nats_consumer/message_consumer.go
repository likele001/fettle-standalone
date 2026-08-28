package nats_consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"ai-platform/chat-service/models"
	"ai-platform/chat-service/service"

	"github.com/nats-io/nats.go"
)

// MessageConsumer NATS 消息消费者
type MessageConsumer struct {
	conn        *nats.Conn
	chatService *service.ChatService
	tenantID    string
}

func NewMessageConsumer(conn *nats.Conn, chatService *service.ChatService, tenantID string) *MessageConsumer {
	return &MessageConsumer{
		conn:        conn,
		chatService: chatService,
		tenantID:    tenantID,
	}
}

// Start 启动消息消费
func (c *MessageConsumer) Start() error {
	// 订阅入站消息
	subject := fmt.Sprintf("chat.inbound.%s", c.tenantID)
	_, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		c.handleInboundMessage(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	log.Printf("Message consumer started, listening on: %s", subject)
	return nil
}

func (c *MessageConsumer) handleInboundMessage(msg *nats.Msg) {
	var event models.ChatEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	log.Printf("Received chat event: type=%s, conv=%s, tenant=%s", event.Type, event.ConversationID, event.TenantID)

	switch event.Type {
	case "message.new":
		c.handleNewMessage(&event)
	case "conversation.created":
		c.handleConversationCreated(&event)
	default:
		log.Printf("Unknown event type: %s", event.Type)
	}
}

// handleNewMessage 渠道入站消息：自动建会话 → 保存消息 → 生成 AI 回复 → 出站回发
func (c *MessageConsumer) handleNewMessage(event *models.ChatEvent) {
	content, _ := event.Data["content"].(string)
	senderID, _ := event.Data["sender_id"].(string)
	senderName, _ := event.Data["sender_name"].(string)
	if content == "" || senderID == "" {
		return
	}

	channel, _ := event.Data["channel"].(string)
	if channel == "" {
		channel = "web"
	}

	// 1. 确保会话存在（按 channel+客户ID 幂等）
	conv, err := c.chatService.EnsureChannelConversation(event.TenantID, event.ConversationID, channel, senderID, senderName)
	if err != nil || conv == nil {
		log.Printf("Failed to ensure conversation: %v", err)
		return
	}
	convID := conv.ID.String()

	// 2. 保存用户消息
	req := &service.SendMessageRequest{ContentType: "text", Content: content}
	if _, err := c.chatService.SendMessage(event.TenantID, convID, senderID, "user", senderName, req); err != nil {
		log.Printf("Failed to save message: %v", err)
		return
	}

	// 3. 生成 AI 回复（收集完整流式回复）
	agentID := ""
	if conv.AgentID.String() != "" {
		agentID = conv.AgentID.String()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	streamCh, _, err := c.chatService.GenerateAIMessage(ctx, event.TenantID, convID, senderID, content, "", agentID, "")
	if err != nil {
		log.Printf("AI generate failed: %v", err)
		return
	}
	var sb strings.Builder
	for tok := range streamCh {
		if tok.Error != "" {
			log.Printf("AI stream error: %s", tok.Error)
		}
		sb.WriteString(tok.Content)
	}
	reply := strings.TrimSpace(sb.String())
	if reply == "" {
		return
	}

	// 4. 回复出站（携带客户标识，供渠道插件回发）
	if err := c.PublishReply(event.TenantID, convID, senderID, reply); err != nil {
		log.Printf("Publish reply failed: %v", err)
	}
}

func (c *MessageConsumer) handleConversationCreated(event *models.ChatEvent) {
	log.Printf("New conversation created: %s", event.ConversationID)
}

// PublishReply 发布回复消息（toUser 为渠道客户标识，供插件回发）
func (c *MessageConsumer) PublishReply(tenantID, convID, toUser, reply string) error {
	event := models.ChatEvent{
		Type:           "message.new",
		ConversationID: convID,
		TenantID:       tenantID,
		Data: map[string]interface{}{
			"content":     reply,
			"sender_type": "agent",
			"to_user":     toUser,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("chat.outbound.%s", tenantID)
	return c.conn.Publish(subject, data)
}
