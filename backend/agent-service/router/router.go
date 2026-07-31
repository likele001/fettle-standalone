package router

import (
	"ai-platform/agent-service/grpc_client"
	"ai-platform/agent-service/handler"
	"ai-platform/agent-service/middleware"
	"ai-platform/agent-service/service"
	"ai-platform/agent-service/repository"
	"ai-platform/agent-service/storage"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, jwtSecret string, aiEngineAddr string, aiClient *grpc_client.AIEngineClient, minioClient *storage.MinioClient) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "agent-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// 初始化依赖
	agentRepo := repository.NewAgentRepository(db)
	kbRepo := repository.NewKnowledgeBaseRepository(db)
	bundleRepo := repository.NewBundleRepository(db)
	agentService := service.NewAgentService(agentRepo, aiClient)
	kbService := service.NewKnowledgeService(kbRepo, agentRepo, aiClient)
	bundleService := service.NewBundleService(bundleRepo)

	agentHandler := handler.NewAgentHandler(agentService)
	kbHandler := handler.NewKnowledgeHandler(kbService, minioClient)
	bundleHandler := handler.NewBundleHandler(bundleService)
	adminHandler := handler.NewAdminHandler(agentService)

	// API 路由
	api := r.Group("")
	api.Use(middleware.JWTAuthMiddleware(jwtSecret))
	api.Use(middleware.TenantMiddleware())
	{
		// 智能体管理
		agents := api.Group("/agents")
		{
			agents.GET("", agentHandler.ListAgents)
			agents.POST("", agentHandler.CreateAgent)
			agents.POST("/test", agentHandler.TestChat)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.PUT("/:id", agentHandler.UpdateAgent)
			agents.DELETE("/:id", agentHandler.DeleteAgent)
		}

		// 知识库管理
		knowledge := api.Group("/knowledge")
		{
			knowledge.GET("", kbHandler.ListKB)
			knowledge.POST("", kbHandler.CreateKB)
			knowledge.GET("/:id", kbHandler.GetKB)
			knowledge.PUT("/:id", kbHandler.UpdateKB)
			knowledge.DELETE("/:id", kbHandler.DeleteKB)
			knowledge.POST("/:id/documents", kbHandler.UploadDocument)
			knowledge.GET("/:id/documents", kbHandler.ListDocuments)
		}

		// 行业套餐 (under /agents group so gateway /agents/*path proxy works)
		agents.GET("/bundles", bundleHandler.ListBundles)
		agents.GET("/bundles/:industry", bundleHandler.GetBundle)
		agents.POST("/bundles/apply", bundleHandler.ApplyBundle)
	}

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(middleware.AdminAuth(jwtSecret))
	{
		admin.GET("/agents", adminHandler.ListAllAgents)
		admin.GET("/agents/stats", adminHandler.GetAgentStats)
	}

	return r
}
