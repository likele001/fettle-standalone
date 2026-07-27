package service

import (
	"context"
	"fmt"
	"time"

	"ai-platform/shared/cache"

	"github.com/mojocn/base64Captcha"
)

// CaptchaService 验证码服务
type CaptchaService struct {
	captcha *base64Captcha.Captcha
	redis   *cache.RedisClient
}

// redisStore 实现 base64Captcha.Store 接口
type redisStore struct {
	redis *cache.RedisClient
}

func (r *redisStore) Set(id string, value string) error {
	key := fmt.Sprintf("captcha:%s", id)
	return r.redis.Set(context.Background(), key, value, 5*time.Minute)
}

func (r *redisStore) Get(id string, clear bool) string {
	key := fmt.Sprintf("captcha:%s", id)
	val, err := r.redis.Get(context.Background(), key)
	if err != nil {
		return ""
	}
	if clear {
		r.redis.Del(context.Background(), key)
	}
	return val
}

func (r *redisStore) Verify(id, answer string, clear bool) bool {
	stored := r.Get(id, clear)
	return stored == answer
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService(redis *cache.RedisClient) *CaptchaService {
	store := &redisStore{redis: redis}
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)
	return &CaptchaService{captcha: captcha, redis: redis}
}

// Generate 生成验证码，返回 (captchaId, base64Image, error)
func (s *CaptchaService) Generate() (string, string, error) {
	id, b64, _, err := s.captcha.Generate()
	if err != nil {
		return "", "", err
	}
	return id, b64, nil
}
