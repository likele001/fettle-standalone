package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connect failed: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// NewRedisClientFromConfig 通过统一配置创建 Redis 客户端
func NewRedisClientFromConfig(addr, password string, db int) (*RedisClient, error) {
	return NewRedisClient(addr, password, db)
}

// Close 关闭连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Client 返回原始redis客户端
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Set 设置键值
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Get 获取键值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del 删除键
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Delete 删除键（Del的别名）
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists 判断键是否存在
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Incr 自增
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Expire 设置过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, ttl).Result()
}

// HSet Hash设置字段
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGetAll Hash获取全部字段
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// RateLimitCheck 限流检查（滑动窗口）
func (r *RedisClient) RateLimitCheck(ctx context.Context, key string, maxRequests int64, window time.Duration) (bool, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	pipe := r.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, key, window)

	cmders, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	countCmd := cmders[1].(*redis.IntCmd)
	count := countCmd.Val()

	return count <= maxRequests, nil
}

// BuildKey 构建带前缀的键
func BuildKey(parts ...string) string {
	return strings.Join(parts, ":")
}
