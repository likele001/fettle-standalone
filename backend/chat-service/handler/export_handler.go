package handler

import (
	"ai-platform/chat-service/service"
	"ai-platform/shared/middleware"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	exportService *service.ExportService
}

func NewExportHandler(exportService *service.ExportService) *ExportHandler {
	return &ExportHandler{exportService: exportService}
}

// ExportConversations exports conversations to CSV
func (h *ExportHandler) ExportConversations(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	startDate, endDate := parseDateRange(c)

	filename := fmt.Sprintf("conversations_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// Add BOM for Excel UTF-8 compatibility
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	if err := h.exportService.ExportConversations(c.Writer, tenantID, startDate, endDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
}

// ExportMessages exports messages to CSV
func (h *ExportHandler) ExportMessages(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	startDate, endDate := parseDateRange(c)

	filename := fmt.Sprintf("messages_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	if err := h.exportService.ExportMessages(c.Writer, tenantID, startDate, endDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
}

// ExportAnalytics exports analytics data to CSV
func (h *ExportHandler) ExportAnalytics(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := fmt.Sscanf(d, "%d", &days); parsed != 1 || err != nil {
			days = 30
		}
	}

	filename := fmt.Sprintf("analytics_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	if err := h.exportService.ExportAnalytics(c.Writer, tenantID, days); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
}

// parseDateRange parses start_date and end_date from query params
func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	var startDate, endDate time.Time

	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = t.Add(24*time.Hour - time.Second) // Include the entire end date
		}
	}

	return startDate, endDate
}
