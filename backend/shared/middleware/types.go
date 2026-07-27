package middleware

import "github.com/golang-jwt/jwt/v5"

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
