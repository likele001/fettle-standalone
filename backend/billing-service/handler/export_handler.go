package handler

import (
	"ai-platform/billing-service/service"
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

// ExportBillingRecords exports billing records to CSV
func (h *ExportHandler) ExportBillingRecords(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	startDate, endDate := parseDateRange(c)

	filename := fmt.Sprintf("billing_records_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// Add BOM for Excel UTF-8 compatibility
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	if err := h.exportService.ExportBillingRecords(c.Writer, tenantID, startDate, endDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
}

// ExportUsageSummary exports monthly usage summary to CSV
func (h *ExportHandler) ExportUsageSummary(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	months := 12
	if m := c.Query("months"); m != "" {
		if parsed, err := fmt.Sscanf(m, "%d", &months); parsed != 1 || err != nil {
			months = 12
		}
	}

	filename := fmt.Sprintf("usage_summary_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	if err := h.exportService.ExportUsageSummary(c.Writer, tenantID, months); err != nil {
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
