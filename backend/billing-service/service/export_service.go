package service

import (
	"ai-platform/billing-service/models"
	"ai-platform/billing-service/repository"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

type ExportService struct {
	repo *repository.BillingRepository
}

func NewExportService(repo *repository.BillingRepository) *ExportService {
	return &ExportService{repo: repo}
}

// ExportBillingRecords exports billing records to CSV
func (s *ExportService) ExportBillingRecords(w io.Writer, tenantID string, startDate, endDate time.Time) error {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant id: %w", err)
	}

	records, err := s.repo.ListBillingRecordsForExport(tID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch billing records: %w", err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "类型", "金额", "货币", "Token消耗", "输入Token", "输出Token", "描述", "状态", "创建时间"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, record := range records {
		row := []string{
			record.ID.String(),
			record.Type,
			fmt.Sprintf("%.2f", record.Amount),
			record.Currency,
			fmt.Sprintf("%d", record.TokensUsed),
			fmt.Sprintf("%d", record.InputTokens),
			fmt.Sprintf("%d", record.OutputTokens),
			record.Description,
			record.Status,
			record.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportUsageSummary exports monthly usage summary to CSV
func (s *ExportService) ExportUsageSummary(w io.Writer, tenantID string, months int) error {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant id: %w", err)
	}

	summary, err := s.repo.GetMonthlyUsageSummary(tID, months)
	if err != nil {
		return fmt.Errorf("failed to fetch usage summary: %w", err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"年月", "总消耗Token", "总费用", "记录数"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, s := range summary {
		row := []string{
			s.Month,
			fmt.Sprintf("%d", s.TotalTokens),
			fmt.Sprintf("%.2f", s.TotalAmount),
			fmt.Sprintf("%d", s.RecordCount),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// MonthlyUsageSummary represents monthly usage aggregation
type MonthlyUsageSummary struct {
	Month       string
	TotalTokens int64
	TotalAmount float64
	RecordCount int64
}

// Helper to convert to models if needed
func (s *ExportService) GetBillingRecordsForExport(tenantID string, startDate, endDate time.Time) ([]models.BillingRecord, error) {
	tID, _ := uuid.Parse(tenantID)
	return s.repo.ListBillingRecordsForExport(tID, startDate, endDate)
}
