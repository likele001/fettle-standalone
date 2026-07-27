package handler

import (
	"net/http"

	"ai-platform/shared/middleware"
	"ai-platform/shared/response"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

// MiniappAuthHandler 小程序认证处理器
type MiniappAuthHandler struct {
	miniappService *service.MiniappAuthService
	userService    *service.UserService
}

// NewMiniappAuthHandler 创建小程序认证处理器
func NewMiniappAuthHandler(miniappService *service.MiniappAuthService, userService *service.UserService) *MiniappAuthHandler {
	return &MiniappAuthHandler{miniappService: miniappService, userService: userService}
}

// Login 小程序微信登录
// GET /miniapp/auth/login?code=xxx&tenant_code=xxx
func (h *MiniappAuthHandler) Login(c *gin.Context) {
	code := c.Query("code")
	tenantCode := c.Query("tenant_code")

	if code == "" {
		response.Error(c, 1001, "缺少微信授权码 code")
		return
	}
	if tenantCode == "" {
		response.Error(c, 1001, "缺少租户编码 tenant_code")
		return
	}

	tokens, userInfo, needBind, openid, err := h.miniappService.MiniappLogin(code, tenantCode)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}

	if needBind {
		response.Success(c, gin.H{
			"need_bind": true,
			"openid":    openid,
		})
		return
	}

	response.Success(c, gin.H{
		"token": tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"user":          userInfo,
	})
}

// BindOpenid 绑定已有账号到微信 openid
// POST /miniapp/auth/bind-openid
func (h *MiniappAuthHandler) BindOpenid(c *gin.Context) {
	var req service.BindOpenidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	tokens, userInfo, err := h.miniappService.BindOpenid(&req)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"user":          userInfo,
	})
}

// GetConfig 获取小程序配置
// GET /public/miniapp/config
func (h *MiniappAuthHandler) GetConfig(c *gin.Context) {
	config, err := h.miniappService.GetMiniAppConfig()
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, config)
}

// RefreshToken 刷新 Token
func (h *MiniappAuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	tokens, err := h.miniappService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}
	response.Success(c, tokens)
}

// GetUserInfo 获取当前用户信息
func (h *MiniappAuthHandler) GetUserInfo(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "未登录"})
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.Error(c, 1004, "用户不存在")
		return
	}

	response.Success(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"phone":      user.Phone,
		"email":      user.Email,
		"name":       user.Name,
		"role":       user.Role,
		"tenant_id":  user.TenantID,
		"avatar_url": user.AvatarURL,
	})
}
