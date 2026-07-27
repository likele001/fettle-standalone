package handler

import (
	"ai-platform/agent-service/service"
	"ai-platform/shared/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BundleHandler struct {
	bundleService *service.BundleService
}

func NewBundleHandler(bundleService *service.BundleService) *BundleHandler {
	return &BundleHandler{bundleService: bundleService}
}

// ListBundles 获取所有行业套餐
func (h *BundleHandler) ListBundles(c *gin.Context) {
	bundles, err := h.bundleService.ListBundles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": bundles})
}

// GetBundle 获取单个行业套餐详情
func (h *BundleHandler) GetBundle(c *gin.Context) {
	industry := c.Param("industry")
	bundle, err := h.bundleService.GetBundle(industry)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "行业套餐不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": bundle})
}

// ApplyBundle 应用行业套餐到当前租户
func (h *BundleHandler) ApplyBundle(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	var req service.ApplyBundleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}

	result, err := h.bundleService.ApplyBundle(tenantID, req.Industry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}
