package handler

import (
	"net/http"
	"strconv"

	"ai-platform/skill-service/service"

	"github.com/gin-gonic/gin"
)

// AdminHandler admin API handler for skill-service
type AdminHandler struct {
	skillService *service.SkillService
}

func NewAdminHandler(skillService *service.SkillService) *AdminHandler {
	return &AdminHandler{skillService: skillService}
}

// ListAllSkills lists all skills (admin)
func (h *AdminHandler) ListAllSkills(c *gin.Context) {
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

	skills, total, err := h.skillService.ListAllSkills(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     skills,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetSkillStats returns skill statistics (admin)
func (h *AdminHandler) GetSkillStats(c *gin.Context) {
	stats, err := h.skillService.GetSkillStats()
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
