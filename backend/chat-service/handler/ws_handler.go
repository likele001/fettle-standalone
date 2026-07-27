package handler

import (
	"ai-platform/chat-service/models"
	"ai-platform/chat-service/service"
	"ai-platform/shared/middleware"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSMessage struct {
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
	ConvID string      `json:"conv_id,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type WSConnection struct {
	conn           *websocket.Conn
	tenantID       string
	conversationID string
	stop           chan struct{}
	mu             sync.Mutex
}

type WSManager struct {
	connections sync.Map
	chatService *service.ChatService
}

func NewWSManager(chatService *service.ChatService) *WSManager {
	return &WSManager{
		chatService: chatService,
	}
}

func (m *WSManager) AddConnection(conn *WSConnection) {
	key := conn.tenantID + ":" + conn.conversationID
	m.connections.Store(key, conn)
}

func (m *WSManager) RemoveConnection(conn *WSConnection) {
	key := conn.tenantID + ":" + conn.conversationID
	m.connections.Delete(key)
}

func (m *WSManager) Broadcast(tenantID, convID string, message *models.Message) {
	key := tenantID + ":" + convID
	if conn, ok := m.connections.Load(key); ok {
		wsConn := conn.(*WSConnection)
		wsConn.SendMessage(&WSMessage{
			Type:   "new_message",
			ConvID: convID,
			Data:   message,
		})
	}
}

func (m *WSManager) SendStatusUpdate(tenantID, convID string, status string) {
	key := tenantID + ":" + convID
	if conn, ok := m.connections.Load(key); ok {
		wsConn := conn.(*WSConnection)
		wsConn.SendMessage(&WSMessage{
			Type:   "status_update",
			ConvID: convID,
			Data:   gin.H{"status": status},
		})
	}
}

func (c *WSConnection) SendMessage(msg *WSMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_ = c.conn.WriteJSON(msg)
	}
}

func (c *WSConnection) Close() {
	close(c.stop)
	c.conn.Close()
}

type WSHandler struct {
	manager *WSManager
}

func NewWSHandler(manager *WSManager) *WSHandler {
	return &WSHandler{manager: manager}
}

func (h *WSHandler) Connect(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	convID := c.Param("id")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "conversation id required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	wsConn := &WSConnection{
		conn:           conn,
		tenantID:       tenantID,
		conversationID: convID,
		stop:           make(chan struct{}),
	}

	h.manager.AddConnection(wsConn)

	go func() {
		defer func() {
			h.manager.RemoveConnection(wsConn)
			wsConn.Close()
		}()

		for {
			select {
			case <-wsConn.stop:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(60 * time.Second))
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}
	}()

	wsConn.SendMessage(&WSMessage{
		Type:   "connected",
		ConvID: convID,
		Data:   gin.H{"message": "WebSocket connected"},
	})
}

func (h *WSHandler) NotifyNewMessage(tenantID, convID string, message *models.Message) {
	h.manager.Broadcast(tenantID, convID, message)
}

func (h *WSHandler) NotifyStatusUpdate(tenantID, convID, status string) {
	h.manager.SendStatusUpdate(tenantID, convID, status)
}

type WSSendMessageRequest struct {
	Content string `json:"content"`
}

func (h *WSHandler) SendMessage(c *gin.Context) {
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

	var req WSSendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}

	sendReq := &service.SendMessageRequest{
		ContentType: "text",
		Content:     req.Content,
	}

	msg, err := h.manager.chatService.SendMessage(tenantID, convID, userID, "user", "", sendReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	h.manager.Broadcast(tenantID, convID, msg)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": msg})
}
