package handler

import (
	"net/http"
	"strconv"

	"ai-platform/billing-service/models"
	"ai-platform/billing-service/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	service *service.PaymentService
}

func NewPaymentHandler(s *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: s}
}

// GetConfig 获取支付配置（管理员）
func (h *PaymentHandler) GetConfig(c *gin.Context) {
	config, err := h.service.GetPaymentConfig()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": &models.PaymentConfig{}})
		return
	}
	// 隐藏 secret
	config.AppSecret = "****"
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": config})
}

// SaveConfig 保存支付配置（管理员）
func (h *PaymentHandler) SaveConfig(c *gin.Context) {
	var req models.PaymentConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	// 获取现有配置（保留 ID）
	existing, _ := h.service.GetPaymentConfig()
	if existing != nil && existing.ID != uuid.Nil {
		req.ID = existing.ID
	}

	if err := h.service.SavePaymentConfig(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// CreateOrder 创建支付订单
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	var req struct {
		PlanID string `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "plan_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "message": "unauthorized"})
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid tenant_id"})
		return
	}

	planUUID, err := uuid.Parse(req.PlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid plan_id"})
		return
	}

	order, err := h.service.CreatePaymentOrder(tenantUUID, planUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": order})
}

// Notify 虎皮椒回调通知
func (h *PaymentHandler) Notify(c *gin.Context) {
	var notify service.XunhupayNotify
	if err := c.ShouldBind(&notify); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	result, err := h.service.HandleNotify(&notify)
	if err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, result)
}

// GetOrders 获取订单列表
func (h *PaymentHandler) GetOrders(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	var tenantUUID uuid.UUID
	if tenantID != "" {
		tenantUUID, _ = uuid.Parse(tenantID)
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	orders, total, err := h.service.GetPaymentOrders(tenantUUID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"data":  orders,
		"total": total,
	})
}

// GetOrder 获取单个订单
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}

	order, err := h.service.GetPaymentOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": order})
}
