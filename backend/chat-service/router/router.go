package router

import (
	"ai-platform/chat-service/aiengine"
	"ai-platform/chat-service/handler"
	"ai-platform/chat-service/middleware"
	"ai-platform/chat-service/repository"
	"ai-platform/chat-service/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, jwtSecret string, aiEngineURL string) (*gin.Engine, *service.ChatService) {
	r := gin.New()
	middleware.MustTrustLocalProxies(r)
	r.Use(gin.Recovery())
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "chat-service",
			"timestamp": time.Now().Unix(),
		})
	})

	convRepo := repository.NewConversationRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	aiClient := aiengine.NewHTTPClient(aiEngineURL)
	workflowClient := aiengine.NewWorkflowClient(aiEngineURL)
	billingClient := service.NewAIBillingClient("http://localhost:9600")
	chatService := service.NewChatService(convRepo, msgRepo, channelRepo, aiClient, workflowClient, billingClient)
	exportService := service.NewExportService(convRepo, msgRepo)

	wsManager := handler.NewWSManager(chatService)
	convHandler := handler.NewConversationHandler(chatService)
	channelHandler := handler.NewChannelHandler(chatService)
	analyticsHandler := handler.NewAnalyticsHandler(chatService)
	exportHandler := handler.NewExportHandler(exportService)
	wsHandler := handler.NewWSHandler(wsManager)
	adminHandler := handler.NewAdminHandler(adminRepo)
	adminChannelHandler := handler.NewAdminChannelHandler(chatService)

	auth := r.Group("")
	auth.Use(middleware.EitherAuthMiddleware(jwtSecret))
	{
		auth.GET("/chats/conversations", convHandler.ListConversations)
		auth.POST("/chats/conversations", convHandler.CreateConversation)
		auth.GET("/chats/conversations/:id", convHandler.GetConversation)
		auth.POST("/chats/conversations/:id/close", convHandler.CloseConversation)
		auth.GET("/chats/conversations/:id/messages", convHandler.GetMessages)
		auth.POST("/chats/conversations/:id/messages", convHandler.SendMessage)
		auth.POST("/chats/conversations/:id/stream", convHandler.StreamChat)
		auth.GET("/chats/conversations/:id/messages/search", convHandler.SearchMessages)

		auth.GET("/chats/quota", convHandler.GetQuota)
		auth.GET("/chats/unread-count", convHandler.GetUnreadCount)

		auth.GET("/chats/channels", channelHandler.ListChannels)
		auth.POST("/chats/channels", channelHandler.CreateChannel)
		auth.GET("/chats/channels/:id", channelHandler.GetChannel)
		auth.PUT("/chats/channels/:id/config", channelHandler.SaveConfig)
		auth.POST("/chats/channels/:id/test", channelHandler.TestChannel)

		auth.GET("/chats/analytics/dashboard", analyticsHandler.Dashboard)
		auth.GET("/chats/analytics/trend", analyticsHandler.Trend)
		auth.GET("/chats/analytics/agent-usage", analyticsHandler.AgentUsage)
		auth.GET("/chats/analytics/channel-distribution", analyticsHandler.ChannelDistribution)
		auth.GET("/chats/analytics/recent-conversations", analyticsHandler.RecentConversations)

		// 数据导出
		auth.GET("/chats/export/conversations", exportHandler.ExportConversations)
		auth.GET("/chats/export/messages", exportHandler.ExportMessages)
		auth.GET("/chats/export/analytics", exportHandler.ExportAnalytics)

		// WebSocket
		auth.GET("/chats/conversations/:id/ws", wsHandler.Connect)
		auth.POST("/chats/conversations/:id/ws/send", wsHandler.SendMessage)
	}

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(middleware.AdminAuth(jwtSecret))
	{
		admin.GET("/analytics/overview", adminHandler.PlatformOverview)
		admin.GET("/channels", adminChannelHandler.ListAllChannels)
		admin.GET("/channels/stats", adminChannelHandler.GetChannelStats)
	}

	return r, chatService
}
