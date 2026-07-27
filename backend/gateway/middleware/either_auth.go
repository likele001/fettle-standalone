package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	sharedMiddleware "ai-platform/shared/middleware"
)

// TenantClaims 租户 JWT 声明
type TenantClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// EitherAuthMiddleware 同时支持管理员 JWT 和租户 JWT
// 管理员 JWT 不含 tenant_id，租户 JWT 含 tenant_id
func EitherAuthMiddleware(secret string) gin.HandlerFunc {
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

		// 先尝试解析为租户 JWT
		tenantClaims := &TenantClaims{}
		token, err := jwt.ParseWithClaims(parts[1], tenantClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err == nil && token.Valid {
			c.Set("user_id", tenantClaims.UserID)
			c.Set("tenant_id", sharedMiddleware.DefaultTenantID)
			c.Set("role", tenantClaims.Role)
			c.Next()
			return
		}

		// 再尝试解析为管理员 JWT
		adminClaims := &AdminClaims{}
		token, err = jwt.ParseWithClaims(parts[1], adminClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err == nil && token.Valid {
			c.Set("user_id", adminClaims.UserID)
			c.Set("role", adminClaims.Role)
			c.Set("tenant_id", sharedMiddleware.DefaultTenantID)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "invalid or expired token"})
		c.Abort()
	}
}
