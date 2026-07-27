package handler

import (
	"ai-platform/shared/middleware"
	"ai-platform/shared/response"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	tokens, userInfo, err := h.authService.Register(&req)
	if err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{
		"tokens": tokens,
		"user":   userInfo,
	})
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	tokens, userInfo, err := h.authService.Login(&req)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}
	response.Success(c, gin.H{
		"tokens": tokens,
		"user":   userInfo,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, 1002, "unauthorized")
		return
	}
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	if err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	response.Success(c, nil)
}

// RefreshToken 刷新Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}
	response.Success(c, tokens)
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	if err := h.authService.Logout(req.RefreshToken); err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, nil)
}
