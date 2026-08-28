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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminAuthService 超级管理员认证服务
type AdminAuthService struct {
	db          *gorm.DB
	redisClient *cache.RedisClient
	jwtSecret   string
}

func NewAdminAuthService(db *gorm.DB, redisClient *cache.RedisClient, jwtSecret string) *AdminAuthService {
	return &AdminAuthService{db: db, redisClient: redisClient, jwtSecret: jwtSecret}
}

// AdminLoginRequest 超管登录请求（手机号/邮箱）
type AdminLoginRequest struct {
	Account     string `json:"account" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaId   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

// AdminUserInfo 超管用户信息
type AdminUserInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url"`
}

// AdminClaims JWT声明
type AdminClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var adminPhoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
var adminEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Login 超管登录（手机号或邮箱）
func (s *AdminAuthService) Login(req *AdminLoginRequest) (*TokenResponse, *AdminUserInfo, error) {
	if !s.verifyCaptcha(req.CaptchaId, req.CaptchaCode) {
		return nil, nil, errors.New("验证码错误或已过期")
	}

	var admin models.AdminUser
	var err error

	if adminPhoneRegex.MatchString(req.Account) {
		err = s.db.Where("phone = ?", req.Account).First(&admin).Error
	} else if adminEmailRegex.MatchString(req.Account) {
		err = s.db.Where("email = ?", req.Account).First(&admin).Error
	} else {
		return nil, nil, errors.New("请输入正确的手机号或邮箱")
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.New("账号或密码错误")
		}
		return nil, nil, err
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

	tokens, err := s.generateAdminTokens(admin.ID.String(), admin.Role)
	if err != nil {
		return nil, nil, err
	}

	userInfo := &AdminUserInfo{
		ID:        admin.ID.String(),
		Username:  admin.Username,
		Phone:     admin.Phone,
		Email:     admin.Email,
		Name:      admin.Name,
		Role:      admin.Role,
		AvatarURL: admin.AvatarURL,
	}

	return tokens, userInfo, nil
}

// ChangePassword 修改密码
func (s *AdminAuthService) ChangePassword(userID string, oldPassword, newPassword string) error {
	var admin models.AdminUser
	if err := s.db.Where("id = ?", userID).First(&admin).Error; err != nil {
		return errors.New("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&admin).Update("password_hash", string(hashedPassword)).Error
}

func (s *AdminAuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	claims := &AdminClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	blacklistKey := fmt.Sprintf("admin_refresh_token_blacklist:%s", refreshToken)
	exists, _ := s.redisClient.Exists(context.Background(), blacklistKey)
	if exists > 0 {
		return nil, errors.New("refresh token has been revoked")
	}
	return s.generateAdminTokens(claims.UserID, claims.Role)
}

func (s *AdminAuthService) generateAdminTokens(userID, role string) (*TokenResponse, error) {
	accessExpire := time.Now().Add(15 * time.Minute)
	refreshExpire := time.Now().Add(7 * 24 * time.Hour)

	accessClaims := AdminClaims{
		UserID:   userID,
		TenantID: "00000000-0000-0000-0000-000000000001",
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpire),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	refreshClaims := AdminClaims{
		UserID:   userID,
		TenantID: "00000000-0000-0000-0000-000000000001",
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

	refreshKey := fmt.Sprintf("admin_refresh_token:%s", refreshTokenString)
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

func (s *AdminAuthService) ParseAdminToken(tokenString string) (*AdminClaims, error) {
	claims := &AdminClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AdminAuthService) Logout(refreshToken string) error {
	blacklistKey := fmt.Sprintf("admin_refresh_token_blacklist:%s", refreshToken)
	return s.redisClient.Set(context.Background(), blacklistKey, "1", 7*24*time.Hour)
}

func (s *AdminAuthService) verifyCaptcha(captchaId, code string) bool {
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

// ---- 管理员用户管理 ----

type CreateAdminUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type UpdateAdminUserRequest struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Email string `json:"email"`
}

func (s *AdminAuthService) ListAdminUsers(page, pageSize int) ([]*AdminUserInfo, int64, error) {
	var admins []models.AdminUser
	var total int64
	s.db.Model(&models.AdminUser{}).Count(&total)
	err := s.db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&admins).Error
	if err != nil {
		return nil, 0, err
	}
	result := make([]*AdminUserInfo, 0, len(admins))
	for _, admin := range admins {
		result = append(result, &AdminUserInfo{
			ID: admin.ID.String(), Username: admin.Username, Phone: admin.Phone,
			Email: admin.Email, Name: admin.Name, Role: admin.Role, AvatarURL: admin.AvatarURL,
		})
	}
	return result, total, nil
}

func (s *AdminAuthService) GetAdminUserByID(userID string) (*AdminUserInfo, error) {
	var admin models.AdminUser
	if err := s.db.Where("id = ?", userID).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &AdminUserInfo{
		ID: admin.ID.String(), Username: admin.Username, Phone: admin.Phone,
		Email: admin.Email, Name: admin.Name, Role: admin.Role, AvatarURL: admin.AvatarURL,
	}, nil
}

func (s *AdminAuthService) CreateAdminUser(req *CreateAdminUserRequest) (*models.AdminUser, error) {
	var existing models.AdminUser
	if err := s.db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errors.New("username already exists")
	}
	if err := s.db.Where("phone = ?", req.Phone).First(&existing).Error; err == nil {
		return nil, errors.New("phone already exists")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	admin := models.AdminUser{
		Username: req.Username, Phone: req.Phone,
		PasswordHash: string(hashedPassword), Name: req.Name, Role: req.Role, Status: "active",
	}
	if err := s.db.Create(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminAuthService) UpdateAdminUser(userID string, req *UpdateAdminUserRequest) (*models.AdminUser, error) {
	var admin models.AdminUser
	if err := s.db.Where("id = ?", userID).First(&admin).Error; err != nil {
		return nil, err
	}
	if req.Name != "" {
		admin.Name = req.Name
	}
	if req.Role != "" {
		admin.Role = req.Role
	}
	if req.Email != "" {
		admin.Email = req.Email
	}
	if err := s.db.Save(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminAuthService) UpdateAdminUserStatus(userID, status string) error {
	return s.db.Model(&models.AdminUser{}).Where("id = ?", userID).Update("status", status).Error
}

func (s *AdminAuthService) ResetAdminPassword(userID, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&models.AdminUser{}).Where("id = ?", userID).Update("password_hash", string(hashedPassword)).Error
}
