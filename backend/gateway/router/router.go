package router

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"ai-platform/gateway/middleware"
	"ai-platform/shared/cache"
	"ai-platform/shared/config"

	"github.com/gin-gonic/gin"
)

// serviceRoutes 服务路由配置
var serviceRoutes = map[string]string{
	"users":   config.GetEnv("USER_SERVICE_ADDR", "http://localhost:9200"),
	"agents":  config.GetEnv("AGENT_SERVICE_ADDR", "http://localhost:9300"),
	"chats":   config.GetEnv("CHAT_SERVICE_ADDR", "http://localhost:9400"),
	"skills":  config.GetEnv("SKILL_SERVICE_ADDR", "http://localhost:9500"),
	"billing": config.GetEnv("BILLING_SERVICE_ADDR", "http://localhost:9600"),
	"ai":      config.GetEnv("AI_ENGINE_ADDR", "http://localhost:9700"),
}

// NewRouter 创建路由
func NewRouter(redisClient *cache.RedisClient, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.RedirectTrailingSlash = false

	// 全局中间件
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "gateway",
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1 路由
	api := r.Group("/api/v1")

	// 公开接口（登录、注册等）
	public := api.Group("")
	public.Use(middleware.RateLimitMiddleware(redisClient, 100, time.Minute))
	{
		public.Any("/auth/*path", proxyTo("users"))
		public.Any("/admin/auth/*path", proxyTo("users"))
		public.Any("/public/*path", proxyTo("users"))
		public.Any("/miniapp/*path", proxyTo("users"))
		public.Any("/webhook/payment/notify", proxyTo("billing"))
		public.Any("/webhook/trigger/*path", proxyTo("ai"))
	}

	// 需要认证的接口（租户 JWT）
	auth := api.Group("")
	auth.Use(middleware.EitherAuthMiddleware(jwtSecret))
	auth.Use(middleware.TenantMiddleware())
	auth.Use(middleware.RateLimitMiddleware(redisClient, 1000, time.Minute))
	{
		auth.Any("/users/*path", proxyTo("users"))
		auth.Any("/agents", proxyTo("agents"))
		auth.Any("/agents/*path", proxyTo("agents"))
		auth.Any("/knowledge", proxyTo("agents"))
		auth.Any("/knowledge/*path", proxyTo("agents"))
		auth.Any("/skills", proxyTo("skills"))
		auth.Any("/skills/*path", proxyTo("skills"))
		auth.GET("/my-permissions", proxyTo("users"))
		auth.Any("/roles", proxyTo("users"))
		auth.Any("/roles/*path", proxyTo("users"))
		auth.Any("/permissions", proxyTo("users"))
		auth.Any("/permissions/*path", proxyTo("users"))
		auth.Any("/ai/*path", proxyTo("ai"))
		auth.Any("/workflows", proxyTo("ai"))
		auth.Any("/workflows/*path", proxyTo("ai"))
		auth.Any("/mcp-servers", proxyTo("ai"))
		auth.Any("/mcp-servers/*path", proxyTo("ai"))
	}

	// chats 接口：同时支持管理员 JWT 和租户 JWT
	api.Any("/chats", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), proxyTo("chats"))
	api.Any("/chats/*path", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), proxyTo("chats"))

	// billing 接口：同时支持管理员 JWT 和租户 JWT
	api.Any("/billing", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), proxyTo("billing"))
	api.Any("/billing/*path", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), proxyTo("billing"))

	// 管理员接口（需要认证，走 user-service）
	// 使用独立的 AdminJWTAuthMiddleware，不要求 tenant_id
	adminPaths := []string{"tenants", "admins", "settings", "monitor", "revenue", "dashboard"}

	// /admin/xxx 路径
	admin := api.Group("/admin")
	admin.Use(middleware.AdminJWTAuthMiddleware(jwtSecret))
	admin.Use(middleware.RateLimitMiddleware(redisClient, 500, time.Minute))
	{
		for _, p := range adminPaths {
			admin.Any("/"+p, proxyTo("users"))
			admin.Any("/"+p+"/*path", proxyTo("users"))
		}
		// 管理员用户管理接口（需要独立注册，因为 /users/*path 被租户接口占用）
		admin.Any("/users", proxyTo("users"))
		admin.Any("/users/*path", proxyTo("users"))
		// payment 和 plans 接口代理到 billing 服务
		admin.Any("/payment", proxyTo("billing"))
		admin.Any("/payment/*path", proxyTo("billing"))
		admin.Any("/plans", proxyTo("billing"))
		admin.Any("/plans/*path", proxyTo("billing"))
		// 订阅和账单记录接口代理到 billing 服务
		admin.Any("/subscriptions", proxyTo("billing"))
		admin.Any("/records", proxyTo("billing"))
		// AI 配置管理接口代理到 user-service
		admin.Any("/ai", proxyTo("users"))
		admin.Any("/ai/*path", proxyTo("users"))
		// 智能体管理
		admin.Any("/agents", proxyTo("agents"))
		admin.Any("/agents/*path", proxyTo("agents"))
		// 技能管理
		admin.Any("/skills", proxyTo("skills"))
		admin.Any("/skills/*path", proxyTo("skills"))
		// 渠道管理
		admin.Any("/channels", proxyTo("chats"))
		admin.Any("/channels/*path", proxyTo("chats"))
		// 数据分析
		admin.Any("/analytics", proxyTo("chats"))
		admin.Any("/analytics/*path", proxyTo("chats"))
	}

	// 同时支持不带 /admin/ 前缀的路径（前端直接调用）
	for _, p := range adminPaths {
		api.Any("/"+p, middleware.AdminJWTAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyToAdmin("users"))
		api.Any("/"+p+"/*path", middleware.AdminJWTAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyToAdmin("users"))
	}

	// plans 接口不带 /admin/ 前缀也代理到 billing
	api.Any("/plans", middleware.AdminJWTAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyTo("billing"))
	api.Any("/plans/*path", middleware.AdminJWTAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyTo("billing"))

	// 租户接口（需要认证，走 user-service）
	tenant := api.Group("/tenant")
	tenant.Use(middleware.JWTAuthMiddleware(jwtSecret))
	tenant.Use(middleware.TenantMiddleware())
	tenant.Use(middleware.RateLimitMiddleware(redisClient, 500, time.Minute))
	{
		tenant.Any("", proxyTo("users"))
		tenant.Any("/*path", proxyTo("users"))
	}

	return r
}

// proxyTo 反向代理到指定服务
func proxyTo(service string) gin.HandlerFunc {
	target := serviceRoutes[service]
	if target == "" {
		return func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": "service not found"})
		}
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "invalid service target"})
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		// 去掉 /api/v1/{service}/ 前缀
		path := req.URL.Path
		parts := strings.SplitN(path, "/", 4)
		if len(parts) >= 4 {
			req.URL.Path = "/" + parts[3]
		}
		if targetURL.Path != "" {
			req.URL.Path = targetURL.Path + req.URL.Path
		}
		req.Host = targetURL.Host
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
		if service == "ai" {
			req.Header.Set("X-Internal-Token", config.GetEnv("AI_ENGINE_INTERNAL_TOKEN", ""))
		}
	}

	return func(c *gin.Context) {
		if traceID, exists := c.Get("trace_id"); exists {
			c.Request.Header.Set("X-Trace-ID", traceID.(string))
		}
		if tid, exists := c.Get("tenant_id"); exists {
			c.Request.Header.Set("X-Tenant-ID", tid.(string))
		}
		if uid, exists := c.Get("user_id"); exists {
			c.Request.Header.Set("X-User-ID", uid.(string))
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// proxyToAdmin 反向代理到指定服务，自动加 /admin/ 前缀
func proxyToAdmin(service string) gin.HandlerFunc {
	target := serviceRoutes[service]
	if target == "" {
		return func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": "service not found"})
		}
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "invalid service target"})
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		// 去掉 /api/v1/ 前缀，加上 /admin/ 前缀
		path := req.URL.Path
		parts := strings.SplitN(path, "/", 4)
		if len(parts) >= 4 {
			req.URL.Path = "/admin/" + parts[3]
		} else {
			req.URL.Path = "/admin/" + path
		}
		if targetURL.Path != "" {
			req.URL.Path = targetURL.Path + req.URL.Path
		}
		req.Host = targetURL.Host
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}

	return func(c *gin.Context) {
		if traceID, exists := c.Get("trace_id"); exists {
			c.Request.Header.Set("X-Trace-ID", traceID.(string))
		}
		if tid, exists := c.Get("tenant_id"); exists {
			c.Request.Header.Set("X-Tenant-ID", tid.(string))
		}
		if uid, exists := c.Get("user_id"); exists {
			c.Request.Header.Set("X-User-ID", uid.(string))
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
