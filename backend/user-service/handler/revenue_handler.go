package handler

import (
	"net/http"
	"time"

	"ai-platform/user-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RevenueHandler 收入统计
type RevenueHandler struct {
	db *gorm.DB
}

func NewRevenueHandler(db *gorm.DB) *RevenueHandler {
	return &RevenueHandler{db: db}
}

// GetStats 获取收入统计
func (h *RevenueHandler) GetStats(c *gin.Context) {
	// 收入数据暂时返回 0（等 billing 表有了再接入）
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total_revenue":   0,
			"monthly_revenue": 0,
			"pending_amount":  0,
			"paid_count":      0,
			"pending_count":   0,
		},
	})
}

// GetTrend 获取收入趋势
func (h *RevenueHandler) GetTrend(c *gin.Context) {
	// 收入趋势暂时返回空数组（等 billing 表有了再接入）
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    []gin.H{},
	})
}

// GetActiveTenantStats 获取活跃租户统计
func (h *RevenueHandler) GetActiveTenantStats(c *gin.Context) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var todayActive, weekActive, monthActive, totalTenants int64

	// 总租户数
	h.db.Model(&models.Tenant{}).Where("status = ?", "active").Count(&totalTenants)

	// 今日活跃（有用户在今天登录过的租户）
	h.db.Model(&models.User{}).
		Where("last_login_at >= ?", todayStart).
		Distinct("tenant_id").
		Count(&todayActive)

	// 本周活跃
	h.db.Model(&models.User{}).
		Where("last_login_at >= ?", weekStart).
		Distinct("tenant_id").
		Count(&weekActive)

	// 本月活跃
	h.db.Model(&models.User{}).
		Where("last_login_at >= ?", monthStart).
		Distinct("tenant_id").
		Count(&monthActive)

	activeRate := 0
	if totalTenants > 0 {
		activeRate = int(monthActive * 100 / totalTenants)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"today_active":  todayActive,
			"week_active":   weekActive,
			"month_active":  monthActive,
			"active_rate":   activeRate,
			"total_tenants": totalTenants,
		},
	})
}
