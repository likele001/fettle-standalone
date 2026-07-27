package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health 健康检查
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "gateway",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

// Ready 就绪检查
func (h *HealthHandler) Ready(c *gin.Context) {
	// 可以添加依赖检查逻辑，如数据库连接、Redis连接等
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"service":   "gateway",
		"timestamp": time.Now().Unix(),
	})
}
