package handler

import (
	"ai-platform/agent-service/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminHandler admin API handler
type AdminHandler struct {
	agentService *service.AgentService
}

func NewAdminHandler(agentService *service.AgentService) *AdminHandler {
	return &AdminHandler{agentService: agentService}
}

// ListAllAgents lists all agents across all tenants
func (h *AdminHandler) ListAllAgents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	agents, total, err := h.agentService.ListAllAgents(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     agents,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetAgentStats returns agent statistics
func (h *AdminHandler) GetAgentStats(c *gin.Context) {
	stats, err := h.agentService.GetAgentStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}
