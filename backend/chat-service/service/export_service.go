package service

import (
	"ai-platform/chat-service/models"
	"ai-platform/chat-service/repository"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

type ExportService struct {
	convRepo *repository.ConversationRepository
	msgRepo  *repository.MessageRepository
}

func NewExportService(convRepo *repository.ConversationRepository, msgRepo *repository.MessageRepository) *ExportService {
	return &ExportService{convRepo: convRepo, msgRepo: msgRepo}
}

// ExportConversations exports conversations to CSV
func (s *ExportService) ExportConversations(w io.Writer, tenantID string, startDate, endDate time.Time) error {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant id: %w", err)
	}

	conversations, err := s.convRepo.ListByTenantForExport(tID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch conversations: %w", err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "标题", "客户名称", "状态", "消息数", "创建时间", "最后消息时间"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, conv := range conversations {
		lastMsgAt := ""
		if conv.LastMessageAt != nil {
			lastMsgAt = conv.LastMessageAt.Format("2006-01-02 15:04:05")
		}
		row := []string{
			conv.ID.String(),
			conv.Title,
			conv.CustomerName,
			conv.Status,
			fmt.Sprintf("%d", conv.MessageCount),
			conv.CreatedAt.Format("2006-01-02 15:04:05"),
			lastMsgAt,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportMessages exports messages to CSV
func (s *ExportService) ExportMessages(w io.Writer, tenantID string, startDate, endDate time.Time) error {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant id: %w", err)
	}

	messages, err := s.msgRepo.ListByTenantForExport(tID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "会话ID", "发送者类型", "发送者名称", "内容类型", "内容", "状态", "创建时间"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, msg := range messages {
		// Truncate content if too long
		content := msg.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		row := []string{
			msg.ID.String(),
			msg.ConversationID.String(),
			msg.SenderType,
			msg.SenderName,
			msg.ContentType,
			content,
			msg.Status,
			msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportAnalytics exports analytics data to CSV
func (s *ExportService) ExportAnalytics(w io.Writer, tenantID string, days int) error {
	trend, err := s.convRepo.TrendByDay(tenantID, days)
	if err != nil {
		return fmt.Errorf("failed to fetch trend data: %w", err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"日期", "会话数"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, point := range trend {
		row := []string{point.Date, fmt.Sprintf("%d", point.Value)}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// GetConversationsForExport returns conversations for export
func (s *ExportService) GetConversationsForExport(tenantID string, startDate, endDate time.Time) ([]models.Conversation, error) {
	tID, _ := uuid.Parse(tenantID)
	return s.convRepo.ListByTenantForExport(tID, startDate, endDate)
}
