package service

import (
	"ai-platform/chat-service/aiengine"
	"ai-platform/chat-service/models"
	"ai-platform/chat-service/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Message quota limits per plan type (monthly)
var messageQuotaByPlan = map[string]int64{
	"free":       100,
	"pro":        5000,
	"enterprise": 50000,
}

type ChatService struct {
	convRepo      *repository.ConversationRepository
	msgRepo       *repository.MessageRepository
	channelRepo   *repository.ChannelRepository
	aiClient      *aiengine.HTTPClient
	billingClient *AIBillingClient
}

func NewChatService(
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	channelRepo *repository.ChannelRepository,
	aiClient *aiengine.HTTPClient,
	billingClient *AIBillingClient,
) *ChatService {
	return &ChatService{
		convRepo:      convRepo,
		msgRepo:       msgRepo,
		channelRepo:   channelRepo,
		aiClient:      aiClient,
		billingClient: billingClient,
	}
}

type CreateConversationRequest struct {
	Channel        string `json:"channel" binding:"required"`
	ChannelID      string `json:"channel_id"`
	CustomerID     string `json:"customer_id"`
	CustomerName   string `json:"customer_name"`
	CustomerAvatar string `json:"customer_avatar"`
	AgentID        string `json:"agent_id"`
}

type SendMessageRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	Content     string `json:"content" binding:"required"`
}


// HasTenantOwnAPIKey checks whether the tenant has configured their own API key
// in the tenant_api_keys table. If they have, they pay for AI usage themselves
// and the platform should NOT bill them.
func (s *ChatService) HasTenantOwnAPIKey(tenantID uuid.UUID) bool {
	var count int64
	err := s.convRepo.GetDB().Table("tenant_api_keys").
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	if err != nil {
		log.Printf("Failed to check tenant_api_keys for tenant %s: %v", tenantID, err)
		return false // fail-open: treat as platform-managed
	}
	return count > 0
}

// checkMessageQuota checks if the tenant has exceeded their monthly message limit.
// Returns an error with a user-friendly message if the limit is exceeded.
func (s *ChatService) checkMessageQuota(tenantID uuid.UUID) error {
	planType := s.convRepo.GetTenantPlanType(tenantID)
	limit, ok := messageQuotaByPlan[planType]
	if !ok {
		limit = 100 // default to free tier
	}

	used, err := s.convRepo.CountMonthlyMessagesByTenant(tenantID)
	if err != nil {
		log.Printf("Failed to count monthly messages for tenant %s: %v", tenantID, err)
		return nil // allow message on counting error (fail-open)
	}

	if used >= limit {
		return fmt.Errorf("本月消息额度已用完（已使用 %d/%d 条），请升级套餐以获取更多额度", used, limit)
	}
	return nil
}

func (s *ChatService) CreateConversation(tenantID, userID string, req *CreateConversationRequest) (*models.Conversation, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("invalid tenant id")
	}
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	conv := &models.Conversation{
		TenantID:       tID,
		UserID:         uID,
		Channel:        req.Channel,
		ChannelID:      req.ChannelID,
		CustomerID:     req.CustomerID,
		CustomerName:   req.CustomerName,
		CustomerAvatar: req.CustomerAvatar,
		Status:         "active",
		Title:          "新会话",
	}

	if req.AgentID != "" {
		aID, err := uuid.Parse(req.AgentID)
		if err == nil {
			conv.AgentID = aID
		}
	}

	if conv.CustomerName == "" {
		conv.CustomerName = "访客"
	}

	if err := s.convRepo.Create(conv); err != nil {
		return nil, err
	}
	return conv, nil
}

func (s *ChatService) GetConversation(tenantID, convID string) (*models.Conversation, error) {
	conv, err := s.convRepo.GetByID(tenantID, convID)
	if err != nil {
		return nil, errors.New("conversation not found")
	}
	return conv, nil
}

func (s *ChatService) ListConversations(tenantID string, status string, page, pageSize int) ([]models.Conversation, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.convRepo.ListByTenant(tenantID, status, page, pageSize)
}

func (s *ChatService) CloseConversation(convID string) error {
	return s.convRepo.UpdateStatus(convID, "closed")
}

func (s *ChatService) SendMessage(tenantID, convID, senderID, senderType, senderName string, req *SendMessageRequest) (*models.Message, error) {
	tID, _ := uuid.Parse(tenantID)
	cID, _ := uuid.Parse(convID)

	// Check message quota before creating (only for user messages)
	if senderType == "user" {
		if err := s.checkMessageQuota(tID); err != nil {
			return nil, err
		}
	}

	msg := &models.Message{
		ConversationID: cID,
		TenantID:       tID,
		SenderID:       senderID,
		SenderType:     senderType,
		SenderName:     senderName,
		ContentType:    req.ContentType,
		Content:        req.Content,
		Status:         "sent",
		CreatedAt:      time.Now(),
	}

	if err := s.msgRepo.Create(msg); err != nil {
		return nil, err
	}

	s.convRepo.IncrementMessageCount(convID)

	return msg, nil
}

func (s *ChatService) GenerateAIMessage(ctx context.Context, tenantID, convID, userID string, userContent string, systemPrompt string, agentID string, model string) (<-chan StreamToken, *models.Message, error) {
	tID, _ := uuid.Parse(tenantID)
	cID, _ := uuid.Parse(convID)

	// Check message quota for the user's message that triggered AI
	if err := s.checkMessageQuota(tID); err != nil {
		return nil, nil, err
	}

	historyMsgs, err := s.msgRepo.GetRecentMessages(convID, 20)
	if err != nil {
		return nil, nil, fmt.Errorf("get history failed: %w", err)
	}

	history := make([]aiengine.ChatMessage, 0, len(historyMsgs))
	for i := len(historyMsgs) - 1; i >= 0; i-- {
		msg := historyMsgs[i]
		role := "user"
		if msg.SenderType == "agent" || msg.SenderType == "assistant" {
			role = "assistant"
		} else if msg.SenderType == "system" {
			role = "system"
		}
		history = append(history, aiengine.ChatMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	aiReq := aiengine.GenerateReplyRequest{
		TenantID:       tenantID,
		UserID:         userID,
		Input:          userContent,
		AgentID:        agentID,
		ConversationID: convID,
		History:        history,
		SystemPrompt:   systemPrompt,
		Model:          model,
		Context: map[string]string{
			"tenant_id": tenantID,
		},
	}

	streamCh := make(chan StreamToken, 128)

	aiMsg := &models.Message{
		ConversationID: cID,
		TenantID:       tID,
		SenderType:     "agent",
		SenderName:     "AI助手",
		ContentType:    "text",
		Content:        "",
		Status:         "sending",
		CreatedAt:      time.Now(),
	}
	if err := s.msgRepo.Create(aiMsg); err != nil {
		close(streamCh)
		return nil, nil, fmt.Errorf("create ai message failed: %w", err)
	}
	s.convRepo.IncrementMessageCount(convID)

	go func() {
		defer close(streamCh)

		aiStream, err := s.aiClient.StreamReply(ctx, aiReq)
		if err != nil {
			log.Printf("AI stream error: %v", err)
			aiMsg.Content = "抱歉，AI服务暂时不可用，请稍后再试。"
			aiMsg.Status = "failed"
			s.msgRepo.UpdateMessage(aiMsg.ID.String(), map[string]interface{}{
				"content": aiMsg.Content,
				"status":  aiMsg.Status,
			})
			streamCh <- StreamToken{
				Content: "抱歉，AI服务暂时不可用，请稍后再试。",
				IsFinal: true,
			}
			return
		}

		var fullContent strings.Builder
		var modelUsed string
		var tokensUsed int

		for chunk := range aiStream {
			if chunk.IsFinal {
				break
			}
			fullContent.WriteString(chunk.Chunk)
			if chunk.ModelUsed != "" {
				modelUsed = chunk.ModelUsed
			}
			if chunk.TokensUsed > 0 {
				tokensUsed = chunk.TokensUsed
			}
			streamCh <- StreamToken{
				Content:   chunk.Chunk,
				IsFinal:   false,
				MessageID: aiMsg.ID.String(),
			}
		}

		aiMsg.Content = fullContent.String()
		aiMsg.Status = "sent"
		aiMsg.ModelUsed = modelUsed
		aiMsg.TokensUsed = tokensUsed

		s.msgRepo.UpdateMessage(aiMsg.ID.String(), map[string]interface{}{
			"content":    aiMsg.Content,
			"status":     aiMsg.Status,
			"model_used": aiMsg.ModelUsed,
			"tokens_used": aiMsg.TokensUsed,
		})

		// Platform-managed billing: if tenant has no own API key, deduct AI cost
		if s.billingClient != nil && !s.HasTenantOwnAPIKey(tID) {
			// Estimate input/output split: 70% input, 30% output if not tracked separately
			inputTokens := int(float64(tokensUsed) * 0.7)
			outputTokens := tokensUsed - inputTokens
			if err := s.billingClient.DeductAICost(tenantID, modelUsed, inputTokens, outputTokens); err != nil {
				log.Printf("WARNING: AI billing deduct failed for tenant %s: %v", tenantID, err)
				// If insufficient balance, notify via stream (non-fatal for already-generated response)
				if strings.Contains(err.Error(), "insufficient balance") {
					streamCh <- StreamToken{
						Content:   "[余额不足] AI usage quota exceeded, please top up.",
						IsFinal:   false,
						MessageID: aiMsg.ID.String(),
					}
				}
			}
		}

		streamCh <- StreamToken{
			Content:   "",
			IsFinal:   true,
			MessageID: aiMsg.ID.String(),
		}
	}()

	return streamCh, aiMsg, nil
}

type StreamToken struct {
	Content   string
	IsFinal   bool
	MessageID string
	Error     string
}

func (s *ChatService) GetMessages(convID string, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.msgRepo.GetByConversation(convID, limit)
}

func (s *ChatService) SearchMessages(convID, keyword string, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.msgRepo.Search(convID, keyword, limit)
}

func (s *ChatService) ListChannels(tenantID string) ([]models.Channel, error) {
	return s.channelRepo.ListByTenant(tenantID)
}

func (s *ChatService) GetChannel(tenantID, channelID string) (*models.Channel, error) {
	return s.channelRepo.GetByID(tenantID, channelID)
}


// checkChannelQuota checks if the tenant's plan allows the requested channel type.
// Returns an error if the channel type is not in the plan's allowed_channels list.
func (s *ChatService) checkChannelQuota(tenantID string, channelType string) error {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return errors.New("invalid tenant id")
	}

	// Get tenant's plan_type from tenants table
	var planType string
	err = s.convRepo.GetDB().Table("tenants").Select("plan_type").Where("id = ?", tID).Scan(&planType).Error
	if err != nil || planType == "" {
		planType = "free"
	}

	// Get allowed_channels from the plans table
	var allowedChannelsJSON string
	err = s.convRepo.GetDB().Table("plans").Select("allowed_channels").
		Where("type = ?", planType).Scan(&allowedChannelsJSON).Error
	if err != nil {
		log.Printf("Failed to query plan allowed_channels for plan_type=%s: %v", planType, err)
		return nil // fail-open: allow if plan not found
	}

	if allowedChannelsJSON == "" {
		return nil // no restriction defined
	}

	var allowedChannels []string
	if err := json.Unmarshal([]byte(allowedChannelsJSON), &allowedChannels); err != nil {
		log.Printf("Failed to parse allowed_channels JSON for plan_type=%s: %v", planType, err)
		return nil // fail-open on parse error
	}

	for _, ch := range allowedChannels {
		if ch == channelType {
			return nil
		}
	}

	return fmt.Errorf("您的套餐不支持开通此渠道，请升级套餐")
}

func (s *ChatService) CreateChannel(tenantID, channelType, channelName, config string) (*models.Channel, error) {
	// Check channel quota against tenant's plan
	if err := s.checkChannelQuota(tenantID, channelType); err != nil {
		return nil, err
	}

	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	ch := &models.Channel{
		TenantID: tID,
		Type:     channelType,
		Name:     channelName,
		Config:   config,
		Status:   "active",
	}
	if err := s.channelRepo.Upsert(ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *ChatService) SaveChannelConfig(tenantID, channelID, config string) (*models.Channel, error) {
	ch, err := s.channelRepo.GetByID(tenantID, channelID)
	if err != nil {
		return nil, err
	}
	ch.Config = config
	if err := s.channelRepo.Update(ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *ChatService) TestChannel(tenantID, channelID string) error {
	_, err := s.channelRepo.GetByID(tenantID, channelID)
	return err
}

type TrendDataPoint struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

type AgentUsage struct {
	AgentID           string `json:"agent_id"`
	AgentName         string `json:"agent_name"`
	MessageCount      int64  `json:"message_count"`
	ConversationCount int64  `json:"conversation_count"`
}

type DashboardStats struct {
	TotalAgents         int64 `json:"total_agents"`
	TotalConversations  int64 `json:"total_conversations"`
	TotalMessages       int64 `json:"total_messages"`
	TotalKnowledgeBases int64 `json:"total_knowledge_bases"`
	TodayConversations  int64 `json:"today_conversations"`
	TodayMessages       int64 `json:"today_messages"`
}

func (s *ChatService) GetDashboardStats(tenantID string) *DashboardStats {
	var totalConvs, totalMsgs, todayConvs, todayMsgs int64
	tID, err := uuid.Parse(tenantID)
	if err == nil {
		s.convRepo.CountByTenant(tID, &totalConvs)
		s.convRepo.SumMessages(tID, &totalMsgs)
		todayConvs, _ = s.convRepo.CountTodayConversations(tID)
		todayMsgs, _ = s.msgRepo.SumTodayMessages(tID)
	}
	return &DashboardStats{
		TotalConversations:  totalConvs,
		TotalMessages:       totalMsgs,
		TotalAgents:         0,
		TotalKnowledgeBases: 0,
		TodayConversations:  todayConvs,
		TodayMessages:       todayMsgs,
	}
}

func (s *ChatService) GetTrend(tenantID string, days int) ([]TrendDataPoint, error) {
	results, err := s.convRepo.TrendByDay(tenantID, days)
	if err != nil {
		return nil, err
	}
	var points []TrendDataPoint
	for _, r := range results {
		points = append(points, TrendDataPoint{Date: r.Date, Value: r.Value})
	}
	return points, nil
}

func (s *ChatService) GetAgentUsage(tenantID string) ([]AgentUsage, error) {
	results, err := s.convRepo.AgentUsageStats(tenantID)
	if err != nil {
		return nil, err
	}
	var usage []AgentUsage
	for _, r := range results {
		usage = append(usage, AgentUsage{
			AgentID:           r.AgentID,
			AgentName:         r.AgentName,
			ConversationCount: r.ConversationCount,
			MessageCount:      r.MessageCount,
		})
	}
	return usage, nil
}

type ChannelDistribution struct {
	Channel string `json:"channel"`
	Count   int64  `json:"count"`
}

func (s *ChatService) GetChannelDistribution(tenantID string) ([]ChannelDistribution, error) {
	results, err := s.convRepo.ChannelDistribution(tenantID)
	if err != nil {
		return nil, err
	}
	var dist []ChannelDistribution
	channelNames := map[string]string{
		"wechat":  "微信公众号",
		"wecom":   "企业微信",
		"feishu":  "飞书",
		"dingtalk": "钉钉",
		"douyin":  "抖音",
	}
	for _, r := range results {
		name := channelNames[r.Channel]
		if name == "" {
			name = r.Channel
		}
		dist = append(dist, ChannelDistribution{
			Channel: name,
			Count:   r.Count,
		})
	}
	return dist, nil
}

func (s *ChatService) GetRecentConversations(tenantID string, limit int) ([]models.Conversation, error) {
	return s.convRepo.RecentConversations(tenantID, limit)
}

// GetQuotaInfo returns the tenant's current message quota usage.
func (s *ChatService) GetQuotaInfo(tenantID string) (used, limit int64, planType string, err error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return 0, 0, "", errors.New("invalid tenant id")
	}
	planType = s.convRepo.GetTenantPlanType(tID)
	limit = messageQuotaByPlan[planType]
	if limit == 0 {
		limit = 100
	}
	used, err = s.convRepo.CountMonthlyMessagesByTenant(tID)
	if err != nil {
		return 0, 0, planType, err
	}
	return used, limit, planType, nil
}

// ListAllChannels returns all channels across all tenants (admin)
func (s *ChatService) ListAllChannels() ([]models.Channel, error) {
	return s.channelRepo.ListAllChannels()
}

// ChannelStatItem represents a channel type group count
type ChannelStatItem struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// GetChannelStats returns channel counts grouped by type (admin)
func (s *ChatService) GetChannelStats() ([]ChannelStatItem, error) {
	stats, err := s.channelRepo.GetChannelStats()
	if err != nil {
		return nil, err
	}
	var items []ChannelStatItem
	for _, st := range stats {
		items = append(items, ChannelStatItem{
			Type:  st.Type,
			Count: st.Count,
		})
	}
	return items, nil
}

// GetUnreadCount returns active conversation count for notifications
func (s *ChatService) GetUnreadCount(tenantID uuid.UUID) (int64, error) {
	return s.convRepo.CountActiveConversations(tenantID)
}
