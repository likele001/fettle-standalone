package nats_consumer

import (
	"ai-platform/chat-service/models"
	"ai-platform/chat-service/service"
	"encoding/json"
	"fmt"
	"log"

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

	log.Printf("Received chat event: type=%s, conv=%s", event.Type, event.ConversationID)

	// 处理不同类型的消息
	switch event.Type {
	case "message.new":
		c.handleNewMessage(&event)
	case "conversation.created":
		c.handleConversationCreated(&event)
	default:
		log.Printf("Unknown event type: %s", event.Type)
	}
}

func (c *MessageConsumer) handleNewMessage(event *models.ChatEvent) {
	content, _ := event.Data["content"].(string)
	senderID, _ := event.Data["sender_id"].(string)
	senderName, _ := event.Data["sender_name"].(string)

	req := &service.SendMessageRequest{
		ContentType: "text",
		Content:     content,
	}

	_, err := c.chatService.SendMessage(event.TenantID, event.ConversationID, senderID, "user", senderName, req)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
	}
}

func (c *MessageConsumer) handleConversationCreated(event *models.ChatEvent) {
	log.Printf("New conversation created: %s", event.ConversationID)
}

// PublishReply 发布回复消息
func (c *MessageConsumer) PublishReply(tenantID, convID string, reply string) error {
	event := models.ChatEvent{
		Type:           "message.new",
		ConversationID: convID,
		TenantID:       tenantID,
		Data: map[string]interface{}{
			"content":     reply,
			"sender_type": "agent",
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("chat.outbound.%s", tenantID)
	return c.conn.Publish(subject, data)
}
