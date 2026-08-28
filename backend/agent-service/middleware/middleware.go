package middleware

import (
	"net/http"
	"os"

	"ai-platform/shared/middleware"

	"github.com/gin-gonic/gin"
)

type Claims = middleware.Claims

var JWTAuthMiddleware = middleware.JWTAuthMiddleware
var TenantMiddleware = middleware.TenantMiddleware
var LoggerMiddleware = middleware.LoggerMiddleware
var CORSMiddleware = middleware.CORSMiddleware
var TraceMiddleware = middleware.TraceMiddleware


// InternalAuth 内部服务鉴权：优先 X-Internal-Token（与 billing 共用 BILLING_INTERNAL_TOKEN）；未配置时仅放行本机
func InternalAuth() gin.HandlerFunc {
	expected := os.Getenv("BILLING_INTERNAL_TOKEN")
	return func(c *gin.Context) {
		if expected != "" {
			if c.GetHeader("X-Internal-Token") != expected {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "invalid internal token"})
				return
			}
			c.Next()
			return
		}
		ip := c.ClientIP()
		if ip != "127.0.0.1" && ip != "::1" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 1003, "message": "internal endpoint: local only"})
			return
		}
		c.Next()
	}
}
