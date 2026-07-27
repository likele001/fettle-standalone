package middleware

import (
	"ai-platform/shared/middleware"

	"github.com/gin-gonic/gin"
)

type Claims = middleware.Claims

var JWTAuthMiddleware = middleware.JWTAuthMiddleware
var TenantMiddleware = middleware.TenantMiddleware
var LoggerMiddleware = middleware.LoggerMiddleware
var CORSMiddleware = middleware.CORSMiddleware
var TraceMiddleware = middleware.TraceMiddleware

// RecoveryMiddleware panic恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(500, gin.H{
					"code":    5000,
					"message": "internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
