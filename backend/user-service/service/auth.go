package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"ai-platform/shared/cache"
	"ai-platform/shared/middleware"
	"ai-platform/user-service/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db          *gorm.DB
	redisClient *cache.RedisClient
	jwtSecret   string
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, redisClient *cache.RedisClient, jwtSecret string) *AuthService {
	return &AuthService{
		db:          db,
		redisClient: redisClient,
		jwtSecret:   jwtSecret,
	}
}

// RegisterRequest 注册请求（手机号/邮箱注册，无需用户名）
type RegisterRequest struct {
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Password    string `json:"password" binding:"required,min=6"`
	Name        string `json:"name"`
	CaptchaId   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

// LoginRequest 登录请求（手机号/邮箱登录）
type LoginRequest struct {
	Account     string `json:"account" binding:"required"` // 手机号或邮箱
	Password    string `json:"password" binding:"required"`
	CaptchaId   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// TokenResponse Token响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        string `json:"id"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	TenantID  string `json:"tenant_id"`
	AvatarURL string `json:"avatar_url"`
}

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Register 注册（手机号或邮箱）
func (s *AuthService) Register(req *RegisterRequest) (*TokenResponse, *UserInfo, error) {
	// 验证码
	if !s.verifyCaptcha(req.CaptchaId, req.CaptchaCode) {
		return nil, nil, errors.New("验证码错误或已过期")
	}

	// 至少填一个
	if req.Phone == "" && req.Email == "" {
		return nil, nil, errors.New("手机号和邮箱至少填写一项")
	}

	var existingUser models.User

	// 检查手机号
	if req.Phone != "" {
		if err := s.db.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
			return nil, nil, errors.New("手机号已被注册")
		} else if err != gorm.ErrRecordNotFound {
			return nil, nil, err
		}
	}

	// 检查邮箱
	if req.Email != "" {
		if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			return nil, nil, errors.New("邮箱已被注册")
		} else if err != gorm.ErrRecordNotFound {
			return nil, nil, err
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	tenantID := uuid.MustParse(middleware.DefaultTenantID)

	user := models.User{
		TenantID:     tenantID,
		Phone:        req.Phone,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         "super_admin",
		Status:       "active",
	}
	if user.Name == "" {
		if req.Phone != "" {
			user.Name = req.Phone
		} else {
			user.Name = req.Email
		}
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, nil, err
	}

	// 生成Token
	tokens, err := s.generateTokens(user.ID.String(), tenantID.String(), user.Role)
	if err != nil {
		return nil, nil, err
	}

	userInfo := &UserInfo{
		ID:       user.ID.String(),
		Phone:    user.Phone,
		Email:    user.Email,
		Name:     user.Name,
		Role:     user.Role,
		TenantID: user.TenantID.String(),
	}

	return tokens, userInfo, nil
}

// Login 登录（手机号或邮箱）
func (s *AuthService) Login(req *LoginRequest) (*TokenResponse, *UserInfo, error) {
	// 验证码
	if !s.verifyCaptcha(req.CaptchaId, req.CaptchaCode) {
		return nil, nil, errors.New("验证码错误或已过期")
	}

	var user models.User
	var err error

	if phoneRegex.MatchString(req.Account) {
		err = s.db.Where("phone = ?", req.Account).First(&user).Error
	} else if emailRegex.MatchString(req.Account) {
		err = s.db.Where("email = ?", req.Account).First(&user).Error
	} else {
		return nil, nil, errors.New("请输入正确的手机号或邮箱")
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 普通用户表没找到，尝试管理员表
			return s.loginAsAdmin(req)
		}
		return nil, nil, err
	}

	if user.Status != "active" {
		return nil, nil, errors.New("账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, errors.New("账号或密码错误")
	}

	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	tokens, err := s.generateTokens(user.ID.String(), user.TenantID.String(), user.Role)
	if err != nil {
		return nil, nil, err
	}

	userInfo := &UserInfo{
		ID:        user.ID.String(),
		Phone:     user.Phone,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		TenantID:  user.TenantID.String(),
		AvatarURL: user.AvatarURL,
	}

	return tokens, userInfo, nil
}

// loginAsAdmin 尝试从管理员表登录
func (s *AuthService) loginAsAdmin(req *LoginRequest) (*TokenResponse, *UserInfo, error) {
	var admin models.AdminUser
	var err error

	if phoneRegex.MatchString(req.Account) {
		err = s.db.Where("phone = ?", req.Account).First(&admin).Error
	} else {
		err = s.db.Where("email = ?", req.Account).First(&admin).Error
	}

	if err != nil {
		return nil, nil, errors.New("账号或密码错误")
	}

	if admin.Status != "active" {
		return nil, nil, errors.New("账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, errors.New("账号或密码错误")
	}

	now := time.Now()
	admin.LastLoginAt = &now
	s.db.Save(&admin)

	tenantID := middleware.DefaultTenantID
	tokens, err := s.generateTokens(admin.ID.String(), tenantID, admin.Role)
	if err != nil {
		return nil, nil, err
	}

	userInfo := &UserInfo{
		ID:        admin.ID.String(),
		Phone:     admin.Phone,
		Email:     admin.Email,
		Name:      admin.Name,
		Role:      admin.Role,
		TenantID:  tenantID,
		AvatarURL: admin.AvatarURL,
	}

	return tokens, userInfo, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID string, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&user).Update("password_hash", string(hashedPassword)).Error
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	blacklistKey := fmt.Sprintf("refresh_token_blacklist:%s", refreshToken)
	exists, _ := s.redisClient.Exists(context.Background(), blacklistKey)
	if exists > 0 {
		return nil, errors.New("refresh token has been revoked")
	}
	return s.generateTokens(claims.UserID, claims.TenantID, claims.Role)
}

func (s *AuthService) generateTokens(userID, tenantID, role string) (*TokenResponse, error) {
	accessExpire := time.Now().Add(15 * time.Minute)
	refreshExpire := time.Now().Add(7 * 24 * time.Hour)

	accessClaims := Claims{
		UserID:   userID,
		TenantID: middleware.DefaultTenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpire),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	refreshClaims := Claims{
		UserID:   userID,
		TenantID: middleware.DefaultTenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpire),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenString)
	if err := s.redisClient.Set(context.Background(), refreshKey, userID, 7*24*time.Hour); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	blacklistKey := fmt.Sprintf("refresh_token_blacklist:%s", refreshToken)
	return s.redisClient.Set(context.Background(), blacklistKey, "1", 7*24*time.Hour)
}

func (s *AuthService) verifyCaptcha(captchaId, code string) bool {
	if captchaId == "" || code == "" {
		return false
	}
	key := fmt.Sprintf("captcha:%s", captchaId)
	storedCode, err := s.redisClient.Get(context.Background(), key)
	if err != nil {
		return false
	}
	s.redisClient.Del(context.Background(), key)
	return storedCode == code
}
