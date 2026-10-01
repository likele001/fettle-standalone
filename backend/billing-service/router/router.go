package router

import (
	"net/http"

	"ai-platform/billing-service/handler"
	"ai-platform/billing-service/middleware"
	"ai-platform/billing-service/service"

	"github.com/gin-gonic/gin"
)

func Setup(billingService *service.BillingService, paymentService *service.PaymentService, exportService *service.ExportService, aiBillingService *service.AIBillingService, jwtSecret string) *gin.Engine {
	r := gin.New()
	middleware.MustTrustLocalProxies(r)
	r.Use(gin.Recovery())
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	billingHandler := handler.NewBillingHandler(billingService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	exportHandler := handler.NewExportHandler(exportService)
	aiBillingHandler := handler.NewAIBillingHandler(aiBillingService)

	// gateway proxy 把 /api/v1/billing/xxx 转成 /billing/xxx
	// 所以这里注册 /billing/ 前缀
	api := r.Group("/billing")
	api.Use(middleware.Auth(jwtSecret))
	{
		api.GET("/plans", billingHandler.GetPlans)
		api.GET("/plans/:id", billingHandler.GetPlan)
		api.GET("/subscriptions", billingHandler.GetSubscription)
		api.POST("/subscriptions", billingHandler.Subscribe)
		api.GET("/records", billingHandler.GetBillingRecords)
		api.GET("/usage", billingHandler.GetMonthlyUsage)
		api.GET("/quota", billingHandler.CheckQuota)

		// 数据导出
		api.GET("/export/records", exportHandler.ExportBillingRecords)
		api.GET("/export/usage", exportHandler.ExportUsageSummary)

		// 支付相关
		api.POST("/payment/create", paymentHandler.CreateOrder)
		api.GET("/payment/orders", paymentHandler.GetOrders)
		api.GET("/payment/orders/:id", paymentHandler.GetOrder)

		// 平台 AI 计费相关
		api.GET("/balance", aiBillingHandler.GetBalance)
		api.GET("/packages", aiBillingHandler.ListPackages)
		api.POST("/packages/purchase", aiBillingHandler.PurchasePackage)
		api.GET("/usage-details", aiBillingHandler.GetUsageDetails)
		api.GET("/platform-pricing", aiBillingHandler.GetPlatformPricing)
		api.POST("/ai/deduct", aiBillingHandler.DeductAICost)
	}

	// 支付回调（不需要认证）
	r.POST("/webhook/payment/notify", paymentHandler.Notify)

	// 管理员接口
	admin := r.Group("/admin")
	admin.Use(middleware.AdminAuth(jwtSecret))
	{
		admin.GET("/payment/config", paymentHandler.GetConfig)
		admin.PUT("/payment/config", paymentHandler.SaveConfig)

		// 套餐管理
		admin.GET("/plans", billingHandler.AdminListPlans)
		admin.POST("/plans", billingHandler.AdminCreatePlan)
		admin.PUT("/plans/:id", billingHandler.AdminUpdatePlan)
		admin.DELETE("/plans/:id", billingHandler.AdminDeletePlan)
		admin.PUT("/plans/:id/status", billingHandler.AdminTogglePlanStatus)

		// 订阅管理
		admin.GET("/subscriptions", billingHandler.AdminListSubscriptions)

		// 账单记录管理
		admin.GET("/records", billingHandler.AdminGetBillingRecords)

		// 平台 AI 计费管理
		admin.POST("/ai/recharge", aiBillingHandler.Recharge)
	}

	// 内部服务端点（ai-engine/chat-service 直连，不经 gateway JWT 代理）
	internal := r.Group("/billing/internal")
	internal.Use(middleware.InternalAuth())
	{
		internal.POST("/deduct-ai-cost", aiBillingHandler.DeductAICost)
		internal.GET("/balance", aiBillingHandler.InternalGetBalance)
	}

	return r
}
