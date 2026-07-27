package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionMemory 短期记忆
type SessionMemory struct {
	redisClient *redis.Client
	ttl         time.Duration
}

// NewSessionMemory 创建短期记忆
func NewSessionMemory(redisClient *redis.Client) *SessionMemory {
	return &SessionMemory{
		redisClient: redisClient,
		ttl:         24 * time.Hour, // 24小时过期
	}
}

// SessionData 会话数据
type SessionData struct {
	ConversationID string                 `json:"conversation_id"`
	TenantID       string                 `json:"tenant_id"`
	UserID         string                 `json:"user_id"`
	AgentID        string                 `json:"agent_id"`
	Summary        string                 `json:"summary"`        // 会话摘要
	Preferences    map[string]interface{} `json:"preferences"`    // 用户偏好
	LastActiveAt   time.Time              `json:"last_active_at"`
	CreatedAt      time.Time              `json:"created_at"`
}

// SaveSession 保存会话记忆
func (m *SessionMemory) SaveSession(ctx context.Context, data *SessionData) error {
	key := fmt.Sprintf("session:%s", data.ConversationID)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return m.redisClient.Set(ctx, key, jsonData, m.ttl).Err()
}

// GetSession 获取会话记忆
func (m *SessionMemory) GetSession(ctx context.Context, conversationID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", conversationID)
	
	result, err := m.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var data SessionData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// UpdateSummary 更新会话摘要
func (m *SessionMemory) UpdateSummary(ctx context.Context, conversationID string, summary string) error {
	key := fmt.Sprintf("session:%s", conversationID)
	
	// 获取现有数据
	data, err := m.GetSession(ctx, conversationID)
	if err != nil {
		return err
	}
	if data == nil {
		return fmt.Errorf("session not found")
	}

	data.Summary = summary
	data.LastActiveAt = time.Now()

	return m.SaveSession(ctx, data)
}

// UpdatePreferences 更新用户偏好
func (m *SessionMemory) UpdatePreferences(ctx context.Context, conversationID string, preferences map[string]interface{}) error {
	key := fmt.Sprintf("session:%s", conversationID)
	
	data, err := m.GetSession(ctx, conversationID)
	if err != nil {
		return err
	}
	if data == nil {
		return fmt.Errorf("session not found")
	}

	if data.Preferences == nil {
		data.Preferences = make(map[string]interface{})
	}

	for k, v := range preferences {
		data.Preferences[k] = v
	}

	return m.SaveSession(ctx, data)
}

// DeleteSession 删除会话记忆
func (m *SessionMemory) DeleteSession(ctx context.Context, conversationID string) error {
	key := fmt.Sprintf("session:%s", conversationID)
	return m.redisClient.Del(ctx, key).Err()
}

// ExtendTTL 延长过期时间
func (m *SessionMemory) ExtendTTL(ctx context.Context, conversationID string) error {
	key := fmt.Sprintf("session:%s", conversationID)
	return m.redisClient.Expire(ctx, key, m.ttl).Err()
}
