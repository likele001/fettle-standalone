package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ai-platform/shared/cache"
	"ai-platform/user-service/models"
	"ai-platform/user-service/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MiniappAuthService 小程序认证服务
type MiniappAuthService struct {
	db          *gorm.DB
	redisClient *cache.RedisClient
	authService *AuthService
	baseWechat  *utils.WechatClient // 平台默认微信配置（优先使用兜底）
}

// NewMiniappAuthService 创建小程序认证服务
func NewMiniappAuthService(db *gorm.DB, redisClient *cache.RedisClient, authService *AuthService) *MiniappAuthService {
	return &MiniappAuthService{
		db:          db,
		redisClient: redisClient,
		authService: authService,
	}
}

// getWechatClient 根据租户配置或平台配置获取微信客户端
func (s *MiniappAuthService) getWechatClient(tenantID string) (*utils.WechatClient, error) {
	// 优先从平台配置获取
	var cfg models.SysConfig
	if err := s.db.First(&cfg, "id = ?", 1).Error; err == nil {
		if cfg.Config.WechatMiniAppId != "" && cfg.Config.WechatMiniAppSecret != "" {
			return utils.NewWechatClient(&utils.WechatConfig{
				AppID:     cfg.Config.WechatMiniAppId,
				AppSecret: cfg.Config.WechatMiniAppSecret,
			}), nil
		}
	}

	// 兜底：使用 base 配置
	if s.baseWechat != nil {
		return s.baseWechat, nil
	}

	return nil, errors.New("微信小程序未配置，请在平台管理后台设置 AppID 和 AppSecret")
}

// MiniappLogin 小程序微信登录
// 通过 code 换取 openid，检查是否已绑定用户
func (s *MiniappAuthService) MiniappLogin(code, tenantCode string) (*TokenResponse, *UserInfo, bool, string, error) {
	// 1. 查找租户
	var tenant models.Tenant
	if err := s.db.Where("code = ?", tenantCode).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, false, "", errors.New("租户不存在，请检查租户编码")
		}
		return nil, nil, false, "", err
	}

	if tenant.Status != "active" {
		return nil, nil, false, "", errors.New("该租户已被禁用")
	}

	// 2. 获取微信客户端
	wechatClient, err := s.getWechatClient(tenant.ID.String())
	if err != nil {
		return nil, nil, false, "", err
	}

	// 3. code 换 openid
	openID, _, err := wechatClient.Code2Session(code)
	if err != nil {
		return nil, nil, false, "", fmt.Errorf("微信登录失败: %w", err)
	}

	// 4. 查找已绑定此 openid 的用户
	var user models.User
	err = s.db.Where("wx_open_id = ? AND tenant_id = ?", openID, tenant.ID).First(&user).Error
	if err == nil {
		// 用户已绑定，检查状态
		if user.Status != "active" {
			return nil, nil, false, "", errors.New("账号已被禁用")
		}
		// 检查租户成员状态（是否已离职）
		var member models.TenantMember
		if err := s.db.Where("user_id = ? AND tenant_id = ?", user.ID, tenant.ID).First(&member).Error; err == nil {
			if member.Status == "resigned" {
				return nil, nil, false, "", errors.New("您已离职，无法登录")
			}
		}

		now := time.Now()
		user.LastLoginAt = &now
		s.db.Save(&user)

		tokens, err := s.authService.generateTokens(user.ID.String(), user.TenantID.String(), user.Role)
		if err != nil {
			return nil, nil, false, "", err
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

		return tokens, userInfo, false, "", nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, false, "", err
	}

	// 未绑定，返回 need_bind
	return nil, nil, true, openID, nil
}

// BindOpenidRequest 绑定请求
type BindOpenidRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Openid     string `json:"openid" binding:"required"`
	TenantCode string `json:"tenant_code" binding:"required"`
}

// BindOpenid 绑定已有账号到微信 openid
func (s *MiniappAuthService) BindOpenid(req *BindOpenidRequest) (*TokenResponse, *UserInfo, error) {
	// 1. 查找租户
	var tenant models.Tenant
	if err := s.db.Where("code = ?", req.TenantCode).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("租户不存在")
		}
		return nil, nil, err
	}

	// 2. 查找用户（租户内按 username）
	var user models.User
	if err := s.db.Where("username = ? AND tenant_id = ?", req.Username, tenant.ID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("用户名或密码错误")
		}
		return nil, nil, err
	}

	// 3. 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, errors.New("用户名或密码错误")
	}

	if user.Status != "active" {
		return nil, nil, errors.New("账号已被禁用")
	}
	// 检查租户成员状态（是否已离职）
	var member models.TenantMember
	if err := s.db.Where("user_id = ? AND tenant_id = ?", user.ID, tenant.ID).First(&member).Error; err == nil {
		if member.Status == "resigned" {
			return nil, nil, errors.New("您已离职，无法登录")
		}
	}

	// 4. 检查 openid 是否已被其他用户绑定
	var existing models.User
	if err := s.db.Where("wx_open_id = ? AND tenant_id = ? AND id != ?", req.Openid, tenant.ID, user.ID).First(&existing).Error; err == nil {
		return nil, nil, errors.New("该微信已绑定其他账号")
	}

	// 5. 绑定 openid
	if err := s.db.Model(&user).Update("wx_open_id", req.Openid).Error; err != nil {
		return nil, nil, err
	}

	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	// 6. 生成 Token
	tokens, err := s.authService.generateTokens(user.ID.String(), user.TenantID.String(), user.Role)
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

// GetMiniAppConfig 获取小程序配置（前端调用）
type MiniAppConfig struct {
	AppID string `json:"app_id"`
}

func (s *MiniappAuthService) GetMiniAppConfig() (*MiniAppConfig, error) {
	var cfg models.SysConfig
	if err := s.db.First(&cfg, "id = ?", 1).Error; err != nil {
		// 无配置时返回空
		return &MiniAppConfig{}, nil
	}

	return &MiniAppConfig{
		AppID: cfg.Config.WechatMiniAppId,
	}, nil
}

// RefreshToken 刷新 Token
func (s *MiniappAuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	return s.authService.RefreshToken(refreshToken)
}

// verifyCaptcha 验证图形验证码
func (s *MiniappAuthService) verifyCaptcha(captchaId, code string) bool {
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
