package handler

import (
	"ai-platform/chat-service/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminChannelHandler struct {
	chatService *service.ChatService
}

func NewAdminChannelHandler(chatService *service.ChatService) *AdminChannelHandler {
	return &AdminChannelHandler{chatService: chatService}
}

func (h *AdminChannelHandler) ListAllChannels(c *gin.Context) {
	channels, err := h.chatService.ListAllChannels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": channels,
			"total": len(channels),
		},
	})
}

func (h *AdminChannelHandler) GetChannelStats(c *gin.Context) {
	stats, err := h.chatService.GetChannelStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	var total int64
	for _, s := range stats {
		total += s.Count
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"by_type": stats,
			"total":   total,
		},
	})
}
