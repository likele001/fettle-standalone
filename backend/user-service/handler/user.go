package handler

import (
	"net/http"
	"strconv"

	"ai-platform/shared/middleware"
	"ai-platform/shared/response"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetCurrentUser 获取当前用户
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.Error(c, 1004, "user not found")
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
		"status":     user.Status,
	})
}

// UpdateCurrentUser 更新当前用户
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"phone":      user.Phone,
		"email":      user.Email,
		"name":       user.Name,
		"avatar_url": user.AvatarURL,
	})
}

// ListUsers 用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	users, total, err := h.userService.ListUsers(tenantID, page, pageSize)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"has_more":  int64(page*pageSize) < total,
		"list":      users,
	})
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	user, err := h.userService.CreateUser(tenantID, &req)
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

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	userID := c.Param("id")
	if err := h.userService.DeleteUser(tenantID, userID); err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, nil)
}

// UpdateUserRole 更新用户角色
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	userID := c.Param("id")
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	user, err := h.userService.UpdateUserRole(tenantID, userID, req.Role)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, gin.H{
		"id":   user.ID,
		"name": user.Name,
		"role": user.Role,
	})
}
