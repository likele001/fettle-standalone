package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"ai-platform/shared/middleware"
)

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager JWT管理器
type JWTManager struct {
	config *JWTConfig
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(config *JWTConfig) *JWTManager {
	return &JWTManager{config: config}
}

// GenerateAccessToken 生成访问令牌（tenant_id 固定为 DefaultTenantID）
func (m *JWTManager) GenerateAccessToken(userID, tenantID, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		TenantID: middleware.DefaultTenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// GenerateRefreshToken 生成刷新令牌（tenant_id 固定为 DefaultTenantID）
func (m *JWTManager) GenerateRefreshToken(userID, tenantID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		TenantID: middleware.DefaultTenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// ParseToken 解析令牌
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ValidateToken 验证令牌
func (m *JWTManager) ValidateToken(tokenString string) bool {
	_, err := m.ParseToken(tokenString)
	return err == nil
}

// ExtractTokenFromHeader 从Authorization头提取token
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization format")
	}

	return parts[1], nil
}
