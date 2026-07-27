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

// AIProviderHandler 平台 AI 厂商管理 Handler（超管后台）
type AIProviderHandler struct {
	aiConfigService *service.AIConfigService
}

func NewAIProviderHandler(aiConfigService *service.AIConfigService) *AIProviderHandler {
	return &AIProviderHandler{aiConfigService: aiConfigService}
}

// List 获取所有厂商列表
// GET /admin/ai/providers
func (h *AIProviderHandler) List(c *gin.Context) {
	status := c.Query("status")
	
	providers, err := h.aiConfigService.ListProviders(c.Request.Context(), status)
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

// Get 获取厂商详情
// GET /admin/ai/providers/:id
func (h *AIProviderHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的厂商 ID"})
		return
	}
	
	provider, err := h.aiConfigService.GetProvider(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": "厂商不存在"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    provider,
	})
}

// Create 创建厂商
// POST /admin/ai/providers
func (h *AIProviderHandler) Create(c *gin.Context) {
	var req models.AIProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "请求参数错误: " + err.Error()})
		return
	}
	
	if req.Code == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "厂商代码和名称不能为空"})
		return
	}
	
	if err := h.aiConfigService.CreateProvider(c.Request.Context(), &req); err != nil {
		logger.Error("create provider failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "创建厂商失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// Update 更新厂商
// PUT /admin/ai/providers/:id
func (h *AIProviderHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的厂商 ID"})
		return
	}
	
	var req models.AIProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "请求参数错误: " + err.Error()})
		return
	}
	
	req.ID = id
	if err := h.aiConfigService.UpdateProvider(c.Request.Context(), &req); err != nil {
		logger.Error("update provider failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "更新厂商失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// Delete 删除厂商
// DELETE /admin/ai/providers/:id
func (h *AIProviderHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的厂商 ID"})
		return
	}
	
	if err := h.aiConfigService.DeleteProvider(c.Request.Context(), id); err != nil {
		logger.Error("delete provider failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "删除厂商失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// AIModelHandler 平台 AI 模型管理 Handler（超管后台）
type AIModelHandler struct {
	aiConfigService *service.AIConfigService
}

func NewAIModelHandler(aiConfigService *service.AIConfigService) *AIModelHandler {
	return &AIModelHandler{aiConfigService: aiConfigService}
}

// List 获取所有模型列表
// GET /admin/ai/models
func (h *AIModelHandler) List(c *gin.Context) {
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
	status := c.Query("status")
	
	modelsList, err := h.aiConfigService.ListModels(c.Request.Context(), providerID, modelType, status)
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

// Get 获取模型详情
// GET /admin/ai/models/:id
func (h *AIModelHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的模型 ID"})
		return
	}
	
	model, err := h.aiConfigService.GetModel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": "模型不存在"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    model,
	})
}

// Create 创建模型
// POST /admin/ai/models
func (h *AIModelHandler) Create(c *gin.Context) {
	var req models.AIModel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "请求参数错误: " + err.Error()})
		return
	}
	
	if req.ProviderID == uuid.Nil || req.ModelCode == "" || req.ModelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "厂商 ID、模型代码和名称不能为空"})
		return
	}
	
	if err := h.aiConfigService.CreateModel(c.Request.Context(), &req); err != nil {
		logger.Error("create model failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "创建模型失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// Update 更新模型
// PUT /admin/ai/models/:id
func (h *AIModelHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的模型 ID"})
		return
	}
	
	var req models.AIModel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "请求参数错误: " + err.Error()})
		return
	}
	
	req.ID = id
	if err := h.aiConfigService.UpdateModel(c.Request.Context(), &req); err != nil {
		logger.Error("update model failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "更新模型失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// Delete 删除模型
// DELETE /admin/ai/models/:id
func (h *AIModelHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "无效的模型 ID"})
		return
	}

	if err := h.aiConfigService.DeleteModel(c.Request.Context(), id); err != nil {
		logger.Error("delete model failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "删除模型失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// SeedProviders 预置常用 AI 厂商
// POST /admin/ai/providers/seed
func (h *AIProviderHandler) SeedProviders(c *gin.Context) {
	if err := h.aiConfigService.SeedProviders(c.Request.Context()); err != nil {
		logger.Error("seed providers failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "预置厂商失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "预置成功",
	})
}