package handler

import (
	"net/http"
	"strconv"

	"ai-platform/billing-service/models"
	"ai-platform/billing-service/service"
	"ai-platform/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AIBillingHandler struct {
	aiService *service.AIBillingService
}

func NewAIBillingHandler(aiService *service.AIBillingService) *AIBillingHandler {
	return &AIBillingHandler{aiService: aiService}
}

// getTenantUUID 从 context 中解析 tenant_id
func (h *AIBillingHandler) getTenantUUID(c *gin.Context) (uuid.UUID, bool) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return uuid.Nil, false
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid tenant id"})
		return uuid.Nil, false
	}
	return tenantUUID, true
}

// GetBalance 获取租户余额（含资源包信息）
func (h *AIBillingHandler) GetBalance(c *gin.Context) {
	tenantUUID, ok := h.getTenantUUID(c)
	if !ok {
		return
	}

	balance, packages, err := h.aiService.GetBalance(tenantUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	var totalTokens, remainingTokens int64
	for _, p := range packages {
		totalTokens += p.TokenAmount
		remainingTokens += p.TokenRemaining
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"balance":          balance,
			"packages":         packages,
			"total_tokens":     totalTokens,
			"remaining_tokens": remainingTokens,
		},
	})
}

// Recharge 管理员充值
func (h *AIBillingHandler) Recharge(c *gin.Context) {
	var req struct {
		TenantID      string  `json:"tenant_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		PaymentMethod string  `json:"payment_method"`
		Remark        string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	tenantUUID, err := uuid.Parse(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid tenant_id"})
		return
	}

	operatorID, _ := c.Get("user_id")
	operatorStr, _ := operatorID.(string)

	if err := h.aiService.Recharge(tenantUUID, req.Amount, req.PaymentMethod, req.Remark, operatorStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "充值成功"})
}

// ListPackages 获取可用资源包列表
func (h *AIBillingHandler) ListPackages(c *gin.Context) {
	packages, err := h.aiService.ListPackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": packages})
}

// PurchasePackage 购买资源包
func (h *AIBillingHandler) PurchasePackage(c *gin.Context) {
	tenantUUID, ok := h.getTenantUUID(c)
	if !ok {
		return
	}

	var req struct {
		PackageID string `json:"package_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	packageUUID, err := uuid.Parse(req.PackageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid package_id"})
		return
	}

	if err := h.aiService.PurchasePackage(tenantUUID, packageUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "购买成功"})
}

// GetUsageDetails 获取用量明细
func (h *AIBillingHandler) GetUsageDetails(c *gin.Context) {
	tenantUUID, ok := h.getTenantUUID(c)
	if !ok {
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

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	details, total, err := h.aiService.GetUsageDetails(tenantUUID, page, pageSize, &startDate, &endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"items":     details,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetPlatformPricing 获取平台模型定价
func (h *AIBillingHandler) GetPlatformPricing(c *gin.Context) {
	pricing, err := h.aiService.GetPlatformPricing()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": pricing})
}

// DeductAICost 扣减 AI 费用（供内部服务调用）
func (h *AIBillingHandler) DeductAICost(c *gin.Context) {
	var req struct {
		TenantID     string `json:"tenant_id" binding:"required"`
		ModelID      string `json:"model_id" binding:"required"`
		InputTokens  int64  `json:"input_tokens" binding:"required"`
		OutputTokens int64  `json:"output_tokens" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	tenantUUID, err := uuid.Parse(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid tenant_id"})
		return
	}

	if err := h.aiService.DeductAICost(tenantUUID, req.ModelID, req.InputTokens, req.OutputTokens); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// ensure models import is used
var _ = models.TenantResourcePackage{}
