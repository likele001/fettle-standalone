package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Conversation struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	AgentID        uuid.UUID      `gorm:"type:uuid;index" json:"agent_id"`
	Channel        string         `gorm:"size:20;not null;default:'web'" json:"channel"`
	ChannelID      string         `gorm:"size:100" json:"channel_id"`
	CustomerID     string         `gorm:"size:100;index" json:"customer_id"`
	CustomerName   string         `gorm:"size:100" json:"customer_name"`
	CustomerAvatar string         `gorm:"size:500" json:"customer_avatar"`
	Title          string         `gorm:"size:200" json:"title"`
	Status         string         `gorm:"size:20;not null;default:'active'" json:"status"`
	Priority       int            `gorm:"default:0" json:"priority"`
	Tags           string         `gorm:"size:500" json:"tags"`
	AssignedTo     *uuid.UUID     `gorm:"type:uuid" json:"assigned_to"`
	LastMessageAt  *time.Time     `json:"last_message_at"`
	MessageCount   int64          `gorm:"default:0" json:"message_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Conversation) TableName() string {
	return "conversations"
}

type Message struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;index" json:"conversation_id"`
	TenantID       uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	SenderID       string    `gorm:"size:100" json:"sender_id"`
	SenderType     string    `gorm:"size:20;not null" json:"sender_type"`
	SenderName     string    `gorm:"size:100" json:"sender_name"`
	ContentType    string    `gorm:"size:20;not null;default:'text'" json:"content_type"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	Metadata       string    `gorm:"type:text" json:"metadata"`
	Status         string    `gorm:"size:20;not null;default:'sent'" json:"status"`
	ModelUsed      string    `gorm:"size:100" json:"model_used"`
	TokensUsed     int       `gorm:"default:0" json:"tokens_used"`
	LatencyMs      int64     `gorm:"default:0" json:"latency_ms"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Message) TableName() string {
	return "messages"
}

type ChatEvent struct {
	Type           string                 `json:"type"`
	ConversationID string                 `json:"conversation_id"`
	TenantID       string                 `json:"tenant_id"`
	Data           map[string]interface{} `json:"data"`
	Timestamp      time.Time              `json:"timestamp"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Conversation{}, &Message{}, &Channel{})
}
