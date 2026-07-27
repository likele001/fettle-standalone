package handler

import (
	"ai-platform/chat-service/service"
	"ai-platform/shared/middleware"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChannelHandler struct {
	chatService *service.ChatService
}

func NewChannelHandler(chatService *service.ChatService) *ChannelHandler {
	return &ChannelHandler{chatService: chatService}
}

func (h *ChannelHandler) ListChannels(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	channels, err := h.chatService.ListChannels(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"items": channels, "total": len(channels)}})
}

func (h *ChannelHandler) GetChannel(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	channelID := c.Param("id")
	ch, err := h.chatService.GetChannel(tenantID, channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": "channel not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": ch})
}

func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	var req struct {
		Type     string          `json:"type" binding:"required"`
		Name     string          `json:"name" binding:"required"`
		Config   json.RawMessage `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}
	configStr := string(req.Config)
	if configStr == "" {
		configStr = "{}"
	}
	ch, err := h.chatService.CreateChannel(tenantID, req.Type, req.Name, configStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": ch})
}

func (h *ChannelHandler) SaveConfig(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	channelID := c.Param("id")
	var req struct {
		Config json.RawMessage `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}
	configStr := string(req.Config)
	ch, err := h.chatService.SaveChannelConfig(tenantID, channelID, configStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": ch})
}

func (h *ChannelHandler) TestChannel(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	channelID := c.Param("id")
	if err := h.chatService.TestChannel(tenantID, channelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"success": true, "message": "连接测试通过"}})
}
