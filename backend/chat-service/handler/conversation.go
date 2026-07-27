package handler

import (
	"ai-platform/chat-service/service"
	"ai-platform/shared/middleware"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ConversationHandler struct {
	chatService *service.ChatService
}

func NewConversationHandler(chatService *service.ChatService) *ConversationHandler {
	return &ConversationHandler{chatService: chatService}
}

func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	var req service.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}

	conv, err := h.chatService.CreateConversation(tenantID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": conv})
}

func (h *ConversationHandler) GetConversation(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	convID := c.Param("id")

	conv, err := h.chatService.GetConversation(tenantID, convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": conv})
}

func (h *ConversationHandler) ListConversations(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	status := c.Query("status")
	page := atoi(c.DefaultQuery("page", "1"), 1)
	pageSize := atoi(c.DefaultQuery("page_size", "20"), 20)

	convs, total, err := h.chatService.ListConversations(tenantID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     convs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *ConversationHandler) CloseConversation(c *gin.Context) {
	convID := c.Param("id")

	if err := h.chatService.CloseConversation(convID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

func (h *ConversationHandler) SendMessage(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	convID := c.Param("id")

	var req service.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}

	msg, err := h.chatService.SendMessage(tenantID, convID, userID, "user", "", &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": msg})
}

type StreamChatRequest struct {
	Content      string `json:"content" binding:"required"`
	ContentType  string `json:"content_type"`
	AgentID      string `json:"agent_id"`
	SystemPrompt string `json:"system_prompt"`
	Model        string `json:"model"`
}

func (h *ConversationHandler) StreamChat(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	convID := c.Param("id")

	var req StreamChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}

	if req.ContentType == "" {
		req.ContentType = "text"
	}

	sendReq := &service.SendMessageRequest{
		ContentType: req.ContentType,
		Content:     req.Content,
	}
	_, err := h.chatService.SendMessage(tenantID, convID, userID, "user", "", sendReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	streamCh, _, err := h.chatService.GenerateAIMessage(
		ctx, tenantID, convID, userID,
		req.Content, req.SystemPrompt, req.AgentID, req.Model,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "streaming not supported"})
		return
	}

	c.Stream(func(w io.Writer) bool {
		token, more := <-streamCh
		if !more {
			return false
		}

		data := map[string]interface{}{
			"content":    token.Content,
			"is_final":   token.IsFinal,
			"message_id": token.MessageID,
		}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		flusher.Flush()
		return true
	})
}

func (h *ConversationHandler) GetMessages(c *gin.Context) {
	convID := c.Param("id")
	limit := atoi64(c.DefaultQuery("limit", "50"), 50)

	messages, err := h.chatService.GetMessages(convID, int(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": messages})
}

func (h *ConversationHandler) SearchMessages(c *gin.Context) {
	convID := c.Param("id")
	keyword := c.Query("keyword")
	limit := atoi64(c.DefaultQuery("limit", "20"), 20)

	messages, err := h.chatService.SearchMessages(convID, keyword, int(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": messages})
}

// GetQuota returns the tenant's current message quota usage
func (h *ConversationHandler) GetQuota(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	used, limit, planType, err := h.chatService.GetQuotaInfo(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"plan_type":  planType,
			"used":       used,
			"limit":      limit,
			"remaining":  limit - used,
			"percentage": float64(used) / float64(limit) * 100,
		},
	})
}

func atoi(s string, def int) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return def
		}
	}
	if n <= 0 {
		return def
	}
	return n
}

func atoi64(s string, def int64) int64 {
	var n int64 = 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		} else {
			return def
		}
	}
	if n <= 0 {
		return def
	}
	return n
}

// GetUnreadCount returns the count of active conversations
func (h *ConversationHandler) GetUnreadCount(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid tenant"})
		return
	}
	count, err := h.chatService.GetUnreadCount(tID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"count": count}})
}
