package middleware

import (
	"net/http"
	"os"
	"strings"

	"ai-platform/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims = middleware.Claims

type AdminClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var JWTAuthMiddleware = middleware.JWTAuthMiddleware
var TenantMiddleware = middleware.TenantMiddleware
var LoggerMiddleware = middleware.LoggerMiddleware
var CORSMiddleware = middleware.CORSMiddleware
var CORS = middleware.CORSMiddleware
var TraceMiddleware = middleware.TraceMiddleware

// Auth 别名，兼容 billing/skill 服务的调用方式
var Auth = middleware.JWTAuthMiddleware

// AdminAuth 管理员认证中间件，不要求 tenant_id
func AdminAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "invalid authorization format"})
			c.Abort()
			return
		}

		claims := &AdminClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "invalid or expired token"})
			c.Abort()
			return
		}

		if claims.Role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}


// InternalAuth 内部服务鉴权：优先 X-Internal-Token；未配置 token 时仅放行本机来源
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
