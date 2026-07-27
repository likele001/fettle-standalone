package handler

import (
	"ai-platform/chat-service/service"
	"ai-platform/shared/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	chatService *service.ChatService
}

func NewAnalyticsHandler(chatService *service.ChatService) *AnalyticsHandler {
	return &AnalyticsHandler{chatService: chatService}
}

func (h *AnalyticsHandler) Dashboard(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	stats := h.chatService.GetDashboardStats(tenantID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": stats})
}

func (h *AnalyticsHandler) Trend(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	trend, err := h.chatService.GetTrend(tenantID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": trend})
}

func (h *AnalyticsHandler) AgentUsage(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	usage, err := h.chatService.GetAgentUsage(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": usage})
}

func (h *AnalyticsHandler) ChannelDistribution(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	data, err := h.chatService.GetChannelDistribution(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": data})
}

func (h *AnalyticsHandler) RecentConversations(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	data, err := h.chatService.GetRecentConversations(tenantID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": data})
}
