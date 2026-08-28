package repository

import (
	"ai-platform/chat-service/models"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// GetDB returns the underlying gorm.DB for direct queries
func (r *ConversationRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *ConversationRepository) Create(conv *models.Conversation) error {
	return r.db.Create(conv).Error
}

func (r *ConversationRepository) GetByID(tenantID, convID string) (*models.Conversation, error) {
	tID, _ := uuid.Parse(tenantID)
	cID, _ := uuid.Parse(convID)

	var conv models.Conversation
	err := r.db.Where("id = ? AND tenant_id = ?", cID, tID).First(&conv).Error
	return &conv, err
}


// GetByChannel 按渠道+客户标识查会话（渠道入站幂等）
func (r *ConversationRepository) GetByChannel(tenantID uuid.UUID, channel, channelID string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Where("tenant_id = ? AND channel = ? AND channel_id = ?", tenantID, channel, channelID).
		Order("updated_at DESC").First(&conv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepository) ListByTenant(tenantID string, status string, page, pageSize int) ([]models.Conversation, int64, error) {
	tID, _ := uuid.Parse(tenantID)

	var convs []models.Conversation
	var total int64

	// Count
	countQuery := r.db.Model(&models.Conversation{}).Where("tenant_id = ?", tID)
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	countQuery.Count(&total)

	// Find
	offset := (page - 1) * pageSize
	findQuery := r.db.Model(&models.Conversation{}).Where("tenant_id = ?", tID)
	if status != "" {
		findQuery = findQuery.Where("status = ?", status)
	}
	err := findQuery.Offset(offset).Limit(pageSize).Order("last_message_at DESC NULLS LAST, created_at DESC").Find(&convs).Error
	return convs, total, err
}

func (r *ConversationRepository) UpdateStatus(convID string, status string) error {
	cID, _ := uuid.Parse(convID)
	return r.db.Model(&models.Conversation{}).Where("id = ?", cID).Update("status", status).Error
}

func (r *ConversationRepository) UpdateLastMessage(convID string, messageCount int64) error {
	cID, _ := uuid.Parse(convID)
	now := time.Now()
	return r.db.Model(&models.Conversation{}).Where("id = ?", cID).
		Updates(map[string]interface{}{
			"last_message_at": now,
			"message_count":   messageCount,
		}).Error
}

func (r *ConversationRepository) IncrementMessageCount(convID string) error {
	cID, _ := uuid.Parse(convID)
	now := time.Now()
	return r.db.Model(&models.Conversation{}).Where("id = ?", cID).
		Updates(map[string]interface{}{
			"last_message_at": now,
			"message_count":   gorm.Expr("message_count + 1"),
		}).Error
}

func (r *ConversationRepository) AssignToAgent(convID string, agentID *uuid.UUID) error {
	cID, _ := uuid.Parse(convID)
	return r.db.Model(&models.Conversation{}).Where("id = ?", cID).Update("assigned_to", agentID).Error
}

func (r *ConversationRepository) CountByTenant(tenantID uuid.UUID, count *int64) error {
	return r.db.Model(&models.Conversation{}).Where("tenant_id = ?", tenantID).Count(count).Error
}

func (r *ConversationRepository) SumMessages(tenantID uuid.UUID, sum *int64) error {
	return r.db.Model(&models.Conversation{}).Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(message_count), 0)").Scan(sum).Error
}

func (r *ConversationRepository) TrendByDay(tenantID string, days int) ([]struct {
	Date  string
	Value int64
}, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	since := time.Now().AddDate(0, 0, -days)
	var results []struct {
		Date  string
		Value int64
	}
	err = r.db.Model(&models.Conversation{}).
		Select("DATE(created_at) as date, COUNT(*) as value").
		Where("tenant_id = ? AND created_at >= ?", tID, since).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	return results, err
}

func (r *ConversationRepository) AgentUsageStats(tenantID string) ([]struct {
	AgentID           string
	AgentName         string
	MessageCount      int64
	ConversationCount int64
}, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	var results []struct {
		AgentID           string
		AgentName         string
		MessageCount      int64
		ConversationCount int64
	}
	err = r.db.Model(&models.Conversation{}).
		Select("agent_id, customer_name as agent_name, SUM(message_count) as message_count, COUNT(*) as conversation_count").
		Where("tenant_id = ? AND agent_id IS NOT NULL", tID).
		Group("agent_id, customer_name").
		Scan(&results).Error
	return results, err
}

func (r *ConversationRepository) ChannelDistribution(tenantID string) ([]struct {
	Channel string
	Count   int64
}, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	var results []struct {
		Channel string
		Count   int64
	}
	err = r.db.Model(&models.Conversation{}).
		Select("channel, COUNT(*) as count").
		Where("tenant_id = ? AND channel IS NOT NULL", tID).
		Group("channel").
		Scan(&results).Error
	return results, err
}

func (r *ConversationRepository) RecentConversations(tenantID string, limit int) ([]models.Conversation, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	var convs []models.Conversation
	err = r.db.Where("tenant_id = ?", tID).
		Order("last_message_at DESC NULLS LAST, created_at DESC").
		Limit(limit).
		Find(&convs).Error
	return convs, err
}

// CountMonthlyMessagesByTenant counts user messages sent by a tenant this calendar month.
// Only counts messages where sender_type = "user" to reflect actual usage.
func (r *ConversationRepository) CountMonthlyMessagesByTenant(tenantID uuid.UUID) (int64, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var count int64
	err := r.db.Model(&models.Message{}).
		Where("tenant_id = ? AND sender_type = ? AND created_at >= ?", tenantID, "user", monthStart).
		Count(&count).Error
	return count, err
}

// GetTenantPlanType queries the tenant's plan type from the tenants table.
// Returns "free" as default if not found.
func (r *ConversationRepository) GetTenantPlanType(tenantID uuid.UUID) string {
	var planType string
	err := r.db.Table("tenants").Select("plan_type").Where("id = ?", tenantID).Scan(&planType).Error
	if err != nil || planType == "" {
		return "free"
	}
	return planType
}

// GetAgentWorkflowID 返回指定智能体绑定的工作流ID。
// 若智能体不存在或未配置工作流，返回空字符串。
func (r *ConversationRepository) GetAgentWorkflowID(tenantID, agentID string) string {
	if agentID == "" {
		return ""
	}
	var workflowID string
	err := r.db.Table("agents").
		Select("workflow_id").
		Where("id = ? AND tenant_id = ?", agentID, tenantID).
		Scan(&workflowID).Error
	if err != nil || workflowID == "" {
		return ""
	}
	return workflowID
}

// GetAgentRAGConfig 返回指定智能体绑定的知识库 ID 列表与人设提示词（system_prompt）。
// 从 agents 表读 knowledge_base_ids（jsonb）与 personality_config（jsonb）。
func (r *ConversationRepository) GetAgentRAGConfig(tenantID, agentID string) (kbIDs []string, persona string) {
	if agentID == "" {
		return nil, ""
	}
	var kbJSON, personaJSON string
	err := r.db.Table("agents").
		Select("knowledge_base_ids").
		Where("id = ? AND tenant_id = ?", agentID, tenantID).
		Scan(&kbJSON).Error
	if err != nil {
	}
	if err == nil && kbJSON != "" {
		_ = json.Unmarshal([]byte(kbJSON), &kbIDs)
	}
	err = r.db.Table("agents").
		Select("personality_config").
		Where("id = ? AND tenant_id = ?", agentID, tenantID).
		Scan(&personaJSON).Error
	if err != nil {
	}
	if err == nil && personaJSON != "" {
		persona = buildPersonaPrompt(personaJSON)
	}
	return kbIDs, persona
}

// buildPersonaPrompt 把 personality_config JSON 转成中文人设提示词
func buildPersonaPrompt(raw string) string {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return ""
	}
	parts := make([]string, 0, 3)
	if role, _ := cfg["role"].(string); role != "" {
		parts = append(parts, fmt.Sprintf("你是%s", role))
	}
	if tone, _ := cfg["tone"].(string); tone != "" {
		parts = append(parts, fmt.Sprintf("语气%s", tone))
	}
	if lang, _ := cfg["language"].(string); lang != "" {
		parts = append(parts, fmt.Sprintf("使用%s回答", lang))
	}
	return strings.Join(parts, "，")
}

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(msg *models.Message) error {
	return r.db.Create(msg).Error
}

func (r *MessageRepository) GetByConversation(convID string, limit int) ([]models.Message, error) {
	cID, _ := uuid.Parse(convID)
	var messages []models.Message
	err := r.db.Where("conversation_id = ?", cID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, err
}

func (r *MessageRepository) GetRecentMessages(convID string, limit int) ([]models.Message, error) {
	cID, _ := uuid.Parse(convID)
	var messages []models.Message
	err := r.db.Where("conversation_id = ?", cID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) Search(convID, keyword string, limit int) ([]models.Message, error) {
	cID, _ := uuid.Parse(convID)
	var messages []models.Message
	err := r.db.Where("conversation_id = ? AND content ILIKE ?", cID, "%"+keyword+"%").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) UpdateMessage(msgID string, updates map[string]interface{}) error {
	mID, _ := uuid.Parse(msgID)
	return r.db.Model(&models.Message{}).Where("id = ?", mID).Updates(updates).Error
}

// ListByTenantForExport returns all messages for a tenant within a date range (for CSV export)
func (r *MessageRepository) ListByTenantForExport(tenantID uuid.UUID, startDate, endDate time.Time) ([]models.Message, error) {
	var messages []models.Message
	query := r.db.Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	err := query.Order("created_at ASC").Find(&messages).Error
	return messages, err
}

// ListByTenantForExport returns all conversations for a tenant within a date range (for CSV export)
func (r *ConversationRepository) ListByTenantForExport(tenantID uuid.UUID, startDate, endDate time.Time) ([]models.Conversation, error) {
	var convs []models.Conversation
	query := r.db.Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	err := query.Order("created_at ASC").Find(&convs).Error
	return convs, err
}





// CountTodayConversations counts today's conversations for a tenant
func (r *ConversationRepository) CountTodayConversations(tenantID uuid.UUID) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.Model(&models.Conversation{}).
		Where("tenant_id = ? AND DATE(created_at) = ?", tenantID, today).
		Count(&count).Error
	return count, err
}

// SumTodayMessages sums today's messages for a tenant
func (r *MessageRepository) SumTodayMessages(tenantID uuid.UUID) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.Model(&models.Message{}).
		Where("tenant_id = ? AND sender_type = ? AND DATE(created_at) = ?", tenantID, "user", today).
		Count(&count).Error
	return count, err
}

// CountActiveConversations counts active (unread/pending) conversations for a tenant
func (r *ConversationRepository) CountActiveConversations(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Conversation{}).
		Where("tenant_id = ? AND status = ?", tenantID, "active").
		Count(&count).Error
	return count, err
}
