package handler

import (
	"ai-platform/shared/middleware"
	"ai-platform/shared/response"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

// AdminAuthHandler 超级管理员认证处理器
type AdminAuthHandler struct {
	adminAuthService *service.AdminAuthService
}

func NewAdminAuthHandler(adminAuthService *service.AdminAuthService) *AdminAuthHandler {
	return &AdminAuthHandler{adminAuthService: adminAuthService}
}

// Login 超管登录
func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req service.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	tokens, userInfo, err := h.adminAuthService.Login(&req)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}
	response.Success(c, gin.H{
		"tokens": tokens,
		"user":   userInfo,
	})
}

// ChangePassword 超管修改密码
func (h *AdminAuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, 1002, "unauthorized")
		return
	}
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	if err := h.adminAuthService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	response.Success(c, nil)
}

// RefreshToken 刷新Token
func (h *AdminAuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	tokens, err := h.adminAuthService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}
	response.Success(c, tokens)
}

// Logout 登出
func (h *AdminAuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	if err := h.adminAuthService.Logout(req.RefreshToken); err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, nil)
}
