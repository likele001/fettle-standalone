package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SessionData 会话数据
type SessionData struct {
	UserID    string                 `json:"user_id"`
	TenantID  string                 `json:"tenant_id"`
	Role      string                 `json:"role"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// SessionCache 会话缓存管理器
type SessionCache struct {
	client *RedisClient
	ttl    time.Duration
}

// NewSessionCache 创建会话缓存
func NewSessionCache(client *RedisClient, ttl time.Duration) *SessionCache {
	if ttl == 0 {
		ttl = 24 * time.Hour // 默认24小时
	}
	return &SessionCache{
		client: client,
		ttl:    ttl,
	}
}

// SetSession 设置会话
func (c *SessionCache) SetSession(ctx context.Context, sessionID string, data *SessionData) error {
	key := fmt.Sprintf(UserSessionKey, sessionID)
	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal session data failed: %w", err)
	}

	return c.client.Set(ctx, key, string(value), c.ttl)
}

// GetSession 获取会话
func (c *SessionCache) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	key := fmt.Sprintf(UserSessionKey, sessionID)
	value, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var data SessionData
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return nil, fmt.Errorf("unmarshal session data failed: %w", err)
	}

	return &data, nil
}

// DeleteSession 删除会话
func (c *SessionCache) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf(UserSessionKey, sessionID)
	return c.client.Delete(ctx, key)
}

// RefreshSession 刷新会话过期时间
func (c *SessionCache) RefreshSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf(UserSessionKey, sessionID)
	_, err := c.client.Expire(ctx, key, c.ttl)
	return err
}

// SetUserData 设置会话中的用户数据
func (c *SessionCache) SetUserData(ctx context.Context, sessionID, field string, value interface{}) error {
	session, err := c.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Data == nil {
		session.Data = make(map[string]interface{})
	}
	session.Data[field] = value

	return c.SetSession(ctx, sessionID, session)
}

// GetUserData 获取会话中的用户数据
func (c *SessionCache) GetUserData(ctx context.Context, sessionID, field string) (interface{}, error) {
	session, err := c.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Data == nil {
		return nil, nil
	}

	return session.Data[field], nil
}

// TokenBlacklist Token黑名单管理
type TokenBlacklist struct {
	client *RedisClient
}

// NewTokenBlacklist 创建Token黑名单
func NewTokenBlacklist(client *RedisClient) *TokenBlacklist {
	return &TokenBlacklist{client: client}
}

// AddToken 添加Token到黑名单
func (b *TokenBlacklist) AddToken(ctx context.Context, token string, expiresAt time.Duration) error {
	key := AuthTokenBlacklistFormat(token)
	return b.client.Set(ctx, key, "1", expiresAt)
}

// IsBlacklisted 检查Token是否在黑名单中
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := AuthTokenBlacklistFormat(token)
	_, err := b.client.Get(ctx, key)
	if err != nil {
		// 键不存在，说明不在黑名单中
		return false, nil
	}
	return true, nil
}

// RemoveToken 从黑名单移除Token
func (b *TokenBlacklist) RemoveToken(ctx context.Context, token string) error {
	key := AuthTokenBlacklistFormat(token)
	return b.client.Delete(ctx, key)
}
