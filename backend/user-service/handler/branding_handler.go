package handler

import (
	"ai-platform/shared/middleware"
	"ai-platform/user-service/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BrandingHandler struct {
	brandingService *service.BrandingService
}

func NewBrandingHandler(brandingService *service.BrandingService) *BrandingHandler {
	return &BrandingHandler{brandingService: brandingService}
}

// GetBranding 获取当前租户品牌配置
func (h *BrandingHandler) GetBranding(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	branding, err := h.brandingService.GetBranding(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": branding})
}

// UpdateBranding 更新当前租户品牌配置
func (h *BrandingHandler) UpdateBranding(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	var req service.UpdateBrandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}

	branding, err := h.brandingService.UpdateBranding(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": branding})
}

// GetPublicBranding 获取公开品牌配置（无需认证）
func (h *BrandingHandler) GetPublicBranding(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "tenant_id is required"})
		return
	}

	branding, err := h.brandingService.GetPublicBranding(tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": branding})
}
