package handler

import (
	"ai-platform/shared/logger"
	"ai-platform/user-service/models"
	"ai-platform/user-service/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantAIConfigHandler 租户 AI 配置管理 Handler（租户后台）
type TenantAIConfigHandler struct {
	aiConfigService *service.AIConfigService
}

func NewTenantAIConfigHandler(aiConfigService *service.AIConfigService) *TenantAIConfigHandler {
	return &TenantAIConfigHandler{aiConfigService: aiConfigService}
}

// GetConfig 获取租户 AI 配置
// GET /tenant/ai/config
func (h *TenantAIConfigHandler) GetConfig(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的租户 ID"})
		return
	}
	
	config, err := h.aiConfigService.GetTenantAIConfig(c.Request.Context(), tenantID)
	if err != nil {
		logger.Error("get tenant ai config failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "获取租户 AI 配置失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    config,
	})
}

// UpdateConfig 更新租户 AI 配置
// PUT /tenant/ai/config
func (h *TenantAIConfigHandler) UpdateConfig(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的租户 ID"})
		return
	}
	
	var req models.TenantAIConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "请求参数错误: " + err.Error()})
		return
	}
	
	if err := h.aiConfigService.UpdateTenantAIConfig(c.Request.Context(), tenantID, &req); err != nil {
		logger.Error("update tenant ai config failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "更新租户 AI 配置失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// ListProviders 获取可选的厂商列表（供租户配置时选择）
// GET /tenant/ai/providers
func (h *TenantAIConfigHandler) ListProviders(c *gin.Context) {
	providers, err := h.aiConfigService.ListProviders(c.Request.Context(), "active")
	if err != nil {
		logger.Error("list providers failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "获取厂商列表失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    providers,
	})
}

// ListModels 获取可选的模型列表（供租户配置时选择）
// GET /tenant/ai/models
func (h *TenantAIConfigHandler) ListModels(c *gin.Context) {
	var providerID *uuid.UUID
	providerIDStr := c.Query("provider_id")
	if providerIDStr != "" {
		pid, err := uuid.Parse(providerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的厂商 ID"})
			return
		}
		providerID = &pid
	}
	
	modelType := c.Query("model_type")
	
	modelsList, err := h.aiConfigService.ListModels(c.Request.Context(), providerID, modelType, "active")
	if err != nil {
		logger.Error("list models failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "获取模型列表失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    modelsList,
	})
}

// TenantAPIKeyHandler 租户 API Key 管理 Handler（租户后台）
type TenantAPIKeyHandler struct {
	aiConfigService *service.AIConfigService
}

func NewTenantAPIKeyHandler(aiConfigService *service.AIConfigService) *TenantAPIKeyHandler {
	return &TenantAPIKeyHandler{aiConfigService: aiConfigService}
}

// List 获取租户所有 API Key
// GET /tenant/ai/api-keys
func (h *TenantAPIKeyHandler) List(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的租户 ID"})
		return
	}
	
	keys, err := h.aiConfigService.ListTenantAPIKeys(c.Request.Context(), tenantID)
	if err != nil {
		logger.Error("list tenant api keys failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "获取 API Key 列表失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    keys,
	})
}

// Create 创建租户 API Key
// POST /tenant/ai/api-keys
func (h *TenantAPIKeyHandler) Create(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的租户 ID"})
		return
	}
	
	var req models.TenantAPIKey
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("bind json failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "请求参数错误: " + err.Error()})
		return
	}
	
	logger.Info("create api key request", 
		zap.String("provider_id", req.ProviderID.String()),
		zap.String("api_key_name", req.APIKeyName),
		zap.String("api_key_value", req.APIKeyValue))
	
	if req.ProviderID == uuid.Nil || req.APIKeyValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "厂商 ID 和 API Key 值不能为空"})
		return
	}
	
	if err := h.aiConfigService.CreateTenantAPIKey(c.Request.Context(), tenantID, &req); err != nil {
		logger.Error("create tenant api key failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "创建 API Key 失败"})
		return
	}
	
	// 返回时隐藏 Key 值
	req.APIKeyValue = ""
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// Update 更新租户 API Key
// PUT /tenant/ai/api-keys/:id
func (h *TenantAPIKeyHandler) Update(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的租户 ID"})
		return
	}
	
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的 API Key ID"})
		return
	}
	
	var req models.TenantAPIKey
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "请求参数错误: " + err.Error()})
		return
	}
	
	req.ID = id
	if err := h.aiConfigService.UpdateTenantAPIKey(c.Request.Context(), tenantID, &req); err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "无权操作该 API Key"})
			return
		}
		logger.Error("update tenant api key failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "更新 API Key 失败"})
		return
	}
	
	req.APIKeyValue = ""
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// Delete 删除租户 API Key
// DELETE /tenant/ai/api-keys/:id
func (h *TenantAPIKeyHandler) Delete(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的租户 ID"})
		return
	}
	
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的 API Key ID"})
		return
	}
	
	if err := h.aiConfigService.DeleteTenantAPIKey(c.Request.Context(), tenantID, id); err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "无权操作该 API Key"})
			return
		}
		logger.Error("delete tenant api key failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "删除 API Key 失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// AIUsageHandler AI 使用统计 Handler
type AIUsageHandler struct {
	aiConfigService *service.AIConfigService
}

func NewAIUsageHandler(aiConfigService *service.AIConfigService) *AIUsageHandler {
	return &AIUsageHandler{aiConfigService: aiConfigService}
}

// GetUsageSummary 获取租户使用统计
// GET /tenant/ai/usage/summary
func (h *AIUsageHandler) GetUsageSummary(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的租户 ID"})
		return
	}

	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "请提供开始和结束日期"})
		return
	}

	summary, err := h.aiConfigService.GetTenantUsageSummary(c.Request.Context(), tenantID, startDate, endDate)
	if err != nil {
		logger.Error("get usage summary failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "获取使用统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    summary,
	})
}

// TestAPIKey 测试 API Key
// POST /tenant/ai/api-keys/:id/test
func (h *TenantAPIKeyHandler) TestAPIKey(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的租户 ID"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的 API Key ID"})
		return
	}

	valid, err := h.aiConfigService.TestAPIKey(c.Request.Context(), id, tenantID)
	if err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "无权操作该 API Key"})
			return
		}
		logger.Error("test api key failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "测试 API Key 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "测试成功",
		"data":    gin.H{"valid": valid},
	})
}