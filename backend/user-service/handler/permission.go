package handler

import (
	"net/http"

	"ai-platform/shared/middleware"
	"ai-platform/shared/response"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permService *service.PermissionService
}

func NewPermissionHandler(permService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permService: permService}
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

func (h *PermissionHandler) ListAll(c *gin.Context) {
	permissions, err := h.permService.GetAllPermissions()
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, permissions)
}

func (h *PermissionHandler) ListByModule(c *gin.Context) {
	module := c.Param("module")
	permissions, err := h.permService.GetPermissionsByModule(module)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, permissions)
}

func (h *PermissionHandler) ListRoles(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	roles, err := h.permService.GetRolesByTenant(tenantID)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, roles)
}

func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	role, err := h.permService.CreateRole(tenantID, req.Name, req.Code, req.Description)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, role)
}

func (h *PermissionHandler) AssignPermissions(c *gin.Context) {
	roleID := c.Param("id")
	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	if err := h.permService.AssignPermissionsToRole(roleID, req.PermissionIDs); err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	role, err := h.permService.GetRoleByID(roleID)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	if role.TenantID.String() != tenantID {
		response.Error(c, 1003, "permission denied")
		return
	}

	permissions, err := h.permService.GetPermissionsByRole(tenantID, role.Code)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, permissions)
}
