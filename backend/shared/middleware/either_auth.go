package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

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

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err == nil && token.Valid {
			c.Set("user_id", claims.UserID)
			c.Set("tenant_id", DefaultTenantID)
			c.Set("role", claims.Role)
			c.Next()
			return
		}

		adminClaims := &AdminClaims{}
		token, err = jwt.ParseWithClaims(parts[1], adminClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err == nil && token.Valid {
			c.Set("user_id", adminClaims.UserID)
			c.Set("role", adminClaims.Role)
			c.Set("tenant_id", DefaultTenantID)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "invalid or expired token"})
		c.Abort()
	}
}
