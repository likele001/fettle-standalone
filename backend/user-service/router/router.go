package router

import (
	"net/http"
	"time"

	"ai-platform/shared/cache"
	"ai-platform/user-service/handler"
	"ai-platform/user-service/middleware"
	"ai-platform/user-service/repository"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter 创建路由
func NewRouter(db *gorm.DB, redisClient *cache.RedisClient, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.RedirectTrailingSlash = false

	r.Use(gin.Recovery())
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 依赖注入
	authService := service.NewAuthService(db, redisClient, jwtSecret)
	adminAuthService := service.NewAdminAuthService(db, redisClient, jwtSecret)
	userService := service.NewUserService(db)
	tenantService := service.NewTenantService(db)
	permService := service.NewPermissionService(db)
	captchaService := service.NewCaptchaService(redisClient)
	brandingService := service.NewBrandingService(db)
	teamService := service.NewTeamService(db)
	miniappAuthService := service.NewMiniappAuthService(db, redisClient, authService)

	// AI 配置服务
	providerRepo := repository.NewAIProviderRepository(db)
	modelRepo := repository.NewAIModelRepository(db)
	tenantConfigRepo := repository.NewTenantAIConfigRepository(db)
	tenantKeyRepo := repository.NewTenantAPIKeyRepository(db)
	usageLogRepo := repository.NewAIUsageLogRepository(db)
	aiConfigService := service.NewAIConfigService(providerRepo, modelRepo, tenantConfigRepo, tenantKeyRepo, usageLogRepo)

	authHandler := handler.NewAuthHandler(authService)
	adminAuthHandler := handler.NewAdminAuthHandler(adminAuthService)
	userHandler := handler.NewUserHandler(userService)
	tenantHandler := handler.NewTenantHandler(tenantService)
	permHandler := handler.NewPermissionHandler(permService)
	captchaHandler := handler.NewCaptchaHandler(captchaService)
	brandingHandler := handler.NewBrandingHandler(brandingService)
	teamHandler := handler.NewTeamHandler(teamService)
	miniappHandler := handler.NewMiniappAuthHandler(miniappAuthService, userService)

	providerHandler := handler.NewAIProviderHandler(aiConfigService)
	modelHandler := handler.NewAIModelHandler(aiConfigService)
	tenantConfigHandler := handler.NewTenantAIConfigHandler(aiConfigService)
	tenantAPIKeyHandler := handler.NewTenantAPIKeyHandler(aiConfigService)
	usageHandler := handler.NewAIUsageHandler(aiConfigService)

	monitorHandler := handler.NewMonitorHandler()
	revenueHandler := handler.NewRevenueHandler(db)
	adminUserHandler := handler.NewAdminUserHandler(adminAuthService)
	settingsHandler := handler.NewSettingsHandler(db)

	requirePerm := func(p string) gin.HandlerFunc {
		return middleware.NewRequirePermission(permService, p)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "user-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// 公开路由 - 验证码
	r.GET("/public/captcha", captchaHandler.GenerateCaptcha)

	// 公开路由 - 品牌配置（无需认证，前端初始化时获取）
	r.GET("/public/branding/:tenant_id", brandingHandler.GetPublicBranding)

	// 公开路由 - 邀请信息（无需认证，接受邀请前查看）
	r.GET("/public/invitation", teamHandler.GetInvitationInfo)

	// 公开路由 - 小程序配置（无需认证，前端初始化时获取）
	r.GET("/public/miniapp/config", miniappHandler.GetConfig)

	// ===== 小程序认证路由 =====
	miniapp := r.Group("/miniapp/auth")
	{
		miniapp.GET("/login", miniappHandler.Login)
		miniapp.POST("/bind-openid", miniappHandler.BindOpenid)
		miniapp.POST("/refresh", miniappHandler.RefreshToken)
	}

	// 公开路由 - 租户用户认证
	public := r.Group("/auth")
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
		public.POST("/refresh", authHandler.RefreshToken)
		public.POST("/logout", authHandler.Logout)
	}

	// 公开路由 - 超级管理员认证
	adminPublic := r.Group("/admin/auth")
	{
		adminPublic.POST("/login", adminAuthHandler.Login)
		adminPublic.POST("/refresh", adminAuthHandler.RefreshToken)
		adminPublic.POST("/logout", adminAuthHandler.Logout)
	}

	// 需要认证的路由
	api := r.Group("")
	api.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		api.GET("/users/me", userHandler.GetCurrentUser)
		api.PUT("/users/me", userHandler.UpdateCurrentUser)
		api.POST("/users/me/change-password", authHandler.ChangePassword)
		api.GET("/users", requirePerm("user:read"), userHandler.ListUsers)
		api.POST("/users", requirePerm("user:write"), userHandler.CreateUser)
		api.DELETE("/users/:id", requirePerm("user:write"), userHandler.DeleteUser)
		api.PUT("/users/:id/role", requirePerm("user:write"), userHandler.UpdateUserRole)

		// 小程序用户信息
		api.GET("/miniapp/user/info", miniappHandler.GetUserInfo)

		api.GET("/tenants/current", tenantHandler.GetCurrentTenant)
		api.PUT("/tenants/current", requirePerm("user:write"), tenantHandler.UpdateCurrentTenant)

		// 品牌定制
		api.GET("/branding", brandingHandler.GetBranding)
		api.PUT("/branding", requirePerm("user:write"), brandingHandler.UpdateBranding)

		// 团队管理
		team := api.Group("/team")
		{
			team.GET("/members", teamHandler.ListMembers)
			team.POST("/invite", teamHandler.CreateInvitation)
			team.GET("/invitations", teamHandler.ListInvitations)
			team.DELETE("/invitations/:id", teamHandler.CancelInvitation)
			team.PUT("/members/:id/role", teamHandler.UpdateMemberRole)
			team.PUT("/members/:id/resign", teamHandler.ResignMember)
			team.POST("/accept-invite", teamHandler.AcceptInvitation)
		}

		api.GET("/my-permissions", permHandler.GetMyPermissions)

	perm := api.Group("/permissions")
		perm.Use(requirePerm("user:write"))
		{
			perm.GET("", permHandler.ListAll)
			perm.GET("/:module", permHandler.ListByModule)
		}

		roles := api.Group("/roles")
		roles.Use(requirePerm("user:write"))
		{
			roles.GET("", permHandler.ListRoles)
			roles.POST("", permHandler.CreateRole)
			roles.POST("/:id/permissions", permHandler.AssignPermissions)
			roles.GET("/:id/permissions", permHandler.GetRolePermissions)
		}

		tenantAI := api.Group("/tenant/ai")
		{
			tenantAI.GET("/config", tenantConfigHandler.GetConfig)
			tenantAI.PUT("/config", tenantConfigHandler.UpdateConfig)
			tenantAI.GET("/providers", tenantConfigHandler.ListProviders)
			tenantAI.GET("/models", tenantConfigHandler.ListModels)
			tenantAI.GET("/api-keys", tenantAPIKeyHandler.List)
			tenantAI.POST("/api-keys", tenantAPIKeyHandler.Create)
			tenantAI.PUT("/api-keys/:id", tenantAPIKeyHandler.Update)
			tenantAI.DELETE("/api-keys/:id", tenantAPIKeyHandler.Delete)
			tenantAI.POST("/api-keys/:id/test", tenantAPIKeyHandler.TestAPIKey)
			tenantAI.GET("/usage/summary", usageHandler.GetUsageSummary)
		}
	}

	adminAuth := func(c *gin.Context) {
		role, ok := middleware.GetRole(c)
		if !ok || role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}

	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware(jwtSecret))
	admin.Use(adminAuth)
	{
		admin.GET("/tenants", tenantHandler.ListTenants)
		admin.GET("/tenants/:id", tenantHandler.GetTenantDetail)
		admin.PUT("/tenants/:id", tenantHandler.UpdateTenant)
		admin.PUT("/tenants/:id/audit", tenantHandler.AuditTenant)
		admin.PUT("/tenants/:id/ban", tenantHandler.BanTenant)

		admin.GET("/ai/providers", providerHandler.List)
		admin.GET("/ai/providers/:id", providerHandler.Get)
		admin.POST("/ai/providers", providerHandler.Create)
		admin.POST("/ai/providers/seed", providerHandler.SeedProviders)
		admin.PUT("/ai/providers/:id", providerHandler.Update)
		admin.DELETE("/ai/providers/:id", providerHandler.Delete)

		admin.GET("/ai/models", modelHandler.List)
		admin.GET("/ai/models/:id", modelHandler.Get)
		admin.POST("/ai/models", modelHandler.Create)
		admin.PUT("/ai/models/:id", modelHandler.Update)
		admin.DELETE("/ai/models/:id", modelHandler.Delete)

		admin.GET("/monitor/health", monitorHandler.GetHealth)
		admin.GET("/monitor/stats", monitorHandler.GetStats)
		admin.GET("/monitor/api-stats", monitorHandler.GetApiStats)
		admin.POST("/monitor/restart/:service", monitorHandler.RestartService)

		admin.GET("/revenue/stats", revenueHandler.GetStats)
		admin.GET("/revenue/trend", revenueHandler.GetTrend)
		admin.GET("/revenue/active-tenants", revenueHandler.GetActiveTenantStats)

		admin.GET("/users", adminUserHandler.ListUsers)
		admin.GET("/users/:id", adminUserHandler.GetUser)
		admin.POST("/users", adminUserHandler.CreateUser)
		admin.PUT("/users/:id", adminUserHandler.UpdateUser)
		admin.PUT("/users/:id/status", adminUserHandler.UpdateUserStatus)
		admin.POST("/users/:id/reset-password", adminUserHandler.ResetPassword)

		// 超管修改密码
		admin.POST("/auth/change-password", adminAuthHandler.ChangePassword)

		admin.GET("/settings", settingsHandler.GetSettings)
		admin.PUT("/settings", settingsHandler.UpdateSettings)
		admin.POST("/settings/test-email", settingsHandler.TestEmail)
		admin.POST("/settings/test-sms", settingsHandler.TestSms)
	}

	return r
}
