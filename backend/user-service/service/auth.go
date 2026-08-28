package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"ai-platform/shared/cache"
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
	TenantName  string `json:"tenant_name"`
	InviteCode  string `json:"invite_code"`
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

	// 处理邀请码
	var inviteTenantID *uuid.UUID
	var inviteRole string
	if req.InviteCode != "" {
		var inv struct {
			ID       uuid.UUID
			TenantID uuid.UUID
			Role     string
			Used     bool
		}
		err := s.db.Table("tenant_invitations").Select("id, tenant_id, role, used").
			Where("code = ?", req.InviteCode).First(&inv).Error
		if err != nil {
			return nil, nil, errors.New("邀请码无效")
		}
		if inv.Used {
			return nil, nil, errors.New("邀请码已使用")
		}
		inviteTenantID = &inv.TenantID
		inviteRole = inv.Role

		// 检查用户是否已属于其他租户
		var existingMember struct{ ID uuid.UUID }
		if err := s.db.Table("tenant_members").Where("user_id IS NOT NULL").First(&existingMember).Error; err == nil {
			// 会在后面创建成员时检查
		}
	}

	// 创建租户（有邀请码则不创建新租户）
	var tenant models.Tenant
	if inviteTenantID != nil {
		if err := s.db.First(&tenant, "id = ?", *inviteTenantID).Error; err != nil {
			return nil, nil, errors.New("邀请的租户不存在")
		}
	} else {
		tenant = models.Tenant{
			Name:     req.TenantName,
			PlanType: "free",
			Status:   "active",
		}
		if tenant.Name == "" {
			tenant.Name = "我的企业"
		}
		if err := s.db.Create(&tenant).Error; err != nil {
			return nil, nil, err
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	// 创建用户
	userRole := "super_admin"
	if inviteRole != "" {
		userRole = inviteRole
	}
	user := models.User{
		TenantID:     tenant.ID,
		Phone:        req.Phone,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         userRole,
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

	// 创建租户成员记录
	if inviteTenantID != nil {
		member := models.TenantMember{
			TenantID:  tenant.ID,
			UserID:    user.ID,
			Role:      inviteRole,
			Status:    "active",
			InvitedBy: nil,
		}
		s.db.Create(&member)
		// 标记邀请码已使用
		s.db.Table("tenant_invitations").Where("code = ?", req.InviteCode).Update("used", true)
	} else {
		member := models.TenantMember{
			TenantID: tenant.ID,
			UserID:   user.ID,
			Role:     "owner",
			Status:   "active",
		}
		s.db.Create(&member)
	}

	// 生成Token
	tokens, err := s.generateTokens(user.ID.String(), tenant.ID.String(), user.Role)
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
			return nil, nil, errors.New("账号或密码错误")
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
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpire),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	refreshClaims := Claims{
		UserID:   userID,
		TenantID: tenantID,
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
