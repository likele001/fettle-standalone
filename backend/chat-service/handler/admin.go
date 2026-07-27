package handler

import (
	"ai-platform/chat-service/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminHandler admin API handler for chat-service
type AdminHandler struct {
	adminRepo *repository.AdminRepository
}

func NewAdminHandler(adminRepo *repository.AdminRepository) *AdminHandler {
	return &AdminHandler{adminRepo: adminRepo}
}

// PlatformOverview returns platform-wide analytics
func (h *AdminHandler) PlatformOverview(c *gin.Context) {
	stats, err := h.adminRepo.GetPlatformStats()
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
