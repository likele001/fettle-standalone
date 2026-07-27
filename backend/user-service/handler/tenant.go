package handler

import (
	"net/http"
	"strconv"

	"ai-platform/shared/middleware"
	"ai-platform/shared/response"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

// TenantHandler 租户处理器
type TenantHandler struct {
	tenantService *service.TenantService
}

// NewTenantHandler 创建租户处理器
func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// GetCurrentTenant 获取当前租户
func (h *TenantHandler) GetCurrentTenant(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	tenant, err := h.tenantService.GetByID(tenantID)
	if err != nil {
		response.Error(c, 1004, "tenant not found")
		return
	}

	response.Success(c, gin.H{
		"id":              tenant.ID,
		"name":            tenant.Name,
		"code":            tenant.Code,
		"plan_type":       tenant.PlanType,
		"plan_expires_at": tenant.PlanExpiresAt,
		"status":          tenant.Status,
		"config":          tenant.Config,
		"created_at":      tenant.CreatedAt,
	})
}

// UpdateCurrentTenant 更新当前租户
func (h *TenantHandler) UpdateCurrentTenant(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	tenant, err := h.tenantService.Update(tenantID, &req)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":        tenant.ID,
		"name":      tenant.Name,
		"code":      tenant.Code,
		"plan_type": tenant.PlanType,
		"status":    tenant.Status,
	})
}

// ========== 管理员接口 ==========

// ListTenants 管理员：获取所有租户列表
func (h *TenantHandler) ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	req := &service.ListTenantsRequest{
		Page:        page,
		PageSize:    pageSize,
		Search:      c.Query("search"),
		Status:      c.Query("status"),
		AuditStatus: c.Query("audit_status"),
	}

	result, err := h.tenantService.ListTenants(req)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, result)
}

// GetTenantDetail 管理员：获取租户详情
func (h *TenantHandler) GetTenantDetail(c *gin.Context) {
	tenantID := c.Param("id")

	tenant, err := h.tenantService.GetByID(tenantID)
	if err != nil {
		response.Error(c, 1004, "tenant not found")
		return
	}

	response.Success(c, tenant)
}

// UpdateTenant 管理员：更新租户信息
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantID := c.Param("id")

	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	tenant, err := h.tenantService.Update(tenantID, &req)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, tenant)
}

// AuditTenant 管理员：审核租户
func (h *TenantHandler) AuditTenant(c *gin.Context) {
	tenantID := c.Param("id")
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, 1002, "unauthorized")
		return
	}

	var req service.AuditTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		response.Error(c, 1001, "status must be 'approved' or 'rejected'")
		return
	}

	tenant, err := h.tenantService.AuditTenant(tenantID, adminID, &req)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, tenant)
}

// BanTenant 管理员：封禁/解封租户
func (h *TenantHandler) BanTenant(c *gin.Context) {
	tenantID := c.Param("id")
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, 1002, "unauthorized")
		return
	}

	var req service.BanTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	tenant, err := h.tenantService.BanTenant(tenantID, adminID, &req)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, tenant)
}
