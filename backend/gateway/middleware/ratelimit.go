package middleware

import (
	"net/http"
	"time"

	"ai-platform/shared/cache"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(redisClient *cache.RedisClient, maxRequests int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取限流key：优先使用租户ID，其次使用IP
		key := "rate_limit:"
		if tenantID, exists := c.Get("tenant_id"); exists {
			if tid, ok := tenantID.(string); ok && tid != "" {
				key += "tenant:" + tid
			}
		}
		if key == "rate_limit:" {
			key += "ip:" + c.ClientIP()
		}

		allowed, err := redisClient.RateLimitCheck(c.Request.Context(), key, maxRequests, window)
		if err != nil {
			// Redis错误时放行，避免误拦截
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 3001, "message": "too many requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}
