package handler

import (
	"net/http"
	"strconv"
	"time"

	"ai-platform/billing-service/models"
	"ai-platform/billing-service/service"
	"ai-platform/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BillingHandler struct {
	service *service.BillingService
}

func NewBillingHandler(service *service.BillingService) *BillingHandler {
	return &BillingHandler{service: service}
}

// GetPlans 获取所有套餐
func (h *BillingHandler) GetPlans(c *gin.Context) {
	plans, err := h.service.GetPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": plans})
}

// GetPlan 获取套餐详情
func (h *BillingHandler) GetPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}

	plan, err := h.service.GetPlan(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// Subscribe 订阅套餐
func (h *BillingHandler) Subscribe(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	var req struct {
		PlanID string `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}

	if err := h.service.Subscribe(tenantUUID, planID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscribed successfully"})
}

// GetSubscription 获取当前订阅
func (h *BillingHandler) GetSubscription(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	sub, err := h.service.GetSubscription(tenantUUID)
	if err != nil {
		// 没有订阅记录，返回默认免费版
		c.JSON(http.StatusOK, gin.H{
			"tenant_id":  tenantID,
			"plan_id":    "free",
			"plan_name":  "免费版",
			"status":     "active",
			"started_at": nil,
			"expires_at": nil,
		})
		return
	}

	// 获取套餐名称
	planName := ""
	plan, planErr := h.service.GetPlan(sub.PlanID)
	if planErr == nil {
		planName = plan.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id":  sub.TenantID,
		"plan_id":    sub.PlanID,
		"plan_name":  planName,
		"status":     sub.Status,
		"started_at": sub.StartDate,
		"expires_at": sub.EndDate,
	})
}

// GetBillingRecords 获取计费记录
func (h *BillingHandler) GetBillingRecords(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	records, err := h.service.GetBillingRecords(tenantUUID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": records})
}

// GetMonthlyUsage 获取月度使用量
func (h *BillingHandler) GetMonthlyUsage(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	if year < 2020 || year > 2099 {
		year = time.Now().Year()
	}
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}

	usage, err := h.service.GetMonthlyUsage(tenantUUID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"usage": usage, "year": year, "month": month})
}

// CheckQuota 检查配额
func (h *BillingHandler) CheckQuota(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	used, limit, err := h.service.CheckQuota(tenantUUID)
	if err != nil {
		// 没有配额记录，返回默认值
		c.JSON(http.StatusOK, gin.H{
			"used":      0,
			"limit":     0,
			"remaining": 0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"used":  used,
		"limit": limit,
		"remaining": limit - used,
	})
}

// AdminListPlans 管理员获取所有套餐
func (h *BillingHandler) AdminListPlans(c *gin.Context) {
	plans, err := h.service.AdminListPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": plans})
}

// AdminCreatePlan 管理员创建套餐
func (h *BillingHandler) AdminCreatePlan(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.AdminCreatePlan(&plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plan)
}

// AdminUpdatePlan 管理员更新套餐
func (h *BillingHandler) AdminUpdatePlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}

	var req models.Plan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := h.service.GetPlan(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	plan.Name = req.Name
	plan.Type = req.Type
	plan.Description = req.Description
	plan.PriceMonthly = req.PriceMonthly
	plan.PriceYearly = req.PriceYearly
	plan.MaxAgents = req.MaxAgents
	plan.MaxKnowledgeBases = req.MaxKnowledgeBases
	plan.MaxDocumentsPerKB = req.MaxDocumentsPerKB
	plan.MaxMessagesPerMonth = req.MaxMessagesPerMonth
	plan.MaxConcurrentSessions = req.MaxConcurrentSessions
	plan.Features = req.Features
	plan.SortOrder = req.SortOrder

	if err := h.service.AdminUpdatePlan(plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plan)
}

// AdminDeletePlan 管理员删除套餐
func (h *BillingHandler) AdminDeletePlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}
	if err := h.service.AdminDeletePlan(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// AdminTogglePlanStatus 管理员切换套餐状态
func (h *BillingHandler) AdminTogglePlanStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AdminTogglePlanStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// AdminListSubscriptions 管理员获取所有订阅
func (h *BillingHandler) AdminListSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	subscriptions, total, err := h.service.AdminGetSubscriptions(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": subscriptions, "total": total, "page": page, "page_size": pageSize})
}

// AdminGetBillingRecords 管理员获取所有账单记录
func (h *BillingHandler) AdminGetBillingRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := h.service.AdminGetBillingRecords(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": records, "total": total, "page": page, "page_size": pageSize})
}
