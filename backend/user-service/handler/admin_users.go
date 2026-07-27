package handler

import (
	"strconv"

	"ai-platform/shared/response"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct {
	adminAuthService *service.AdminAuthService
}

func NewAdminUserHandler(service *service.AdminAuthService) *AdminUserHandler {
	return &AdminUserHandler{adminAuthService: service}
}

func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.adminAuthService.ListAdminUsers(page, pageSize)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, gin.H{
		"items":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AdminUserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.adminAuthService.GetAdminUserByID(userID)
	if err != nil {
		response.Error(c, 1004, "user not found")
		return
	}

	response.Success(c, user)
}

func (h *AdminUserHandler) CreateUser(c *gin.Context) {
	var req service.CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	user, err := h.adminAuthService.CreateAdminUser(&req)
	if err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":    user.ID,
		"phone": user.Phone,
		"name":  user.Name,
		"role":  user.Role,
	})
}

func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var req service.UpdateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	user, err := h.adminAuthService.UpdateAdminUser(userID, &req)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":    user.ID,
		"phone": user.Phone,
		"name":  user.Name,
		"role":  user.Role,
	})
}

func (h *AdminUserHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	if err := h.adminAuthService.UpdateAdminUserStatus(userID, req.Status); err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	if err := h.adminAuthService.ResetAdminPassword(userID, req.Password); err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, nil)
}