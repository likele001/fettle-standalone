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
	"users":   config.GetEnv("USER_SERVICE_ADDR", "http://localhost:20002"),
	"agents":  config.GetEnv("AGENT_SERVICE_ADDR", "http://localhost:20003"),
	"chats":   config.GetEnv("CHAT_SERVICE_ADDR", "http://localhost:20004"),
	"skills":  config.GetEnv("SKILL_SERVICE_ADDR", "http://localhost:20005"),
	"billing": config.GetEnv("BILLING_SERVICE_ADDR", "http://localhost:20006"),
	"ai":      config.GetEnv("AI_ENGINE_ADDR", "http://localhost:20007"),
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
		public.Any("/public/*path", proxyTo("users"))
		public.Any("/miniapp/*path", proxyTo("users"))
		public.Any("/webhook/trigger/*path", proxyTo("ai"))
	}

	// 需要认证的接口
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
		auth.Any("/ai/*path", proxyTo("ai"))
		auth.Any("/workflows", proxyTo("ai"))
		auth.Any("/workflows/*path", proxyTo("ai"))
		auth.Any("/mcp-servers", proxyTo("ai"))
		auth.Any("/mcp-servers/*path", proxyTo("ai"))
		auth.Any("/channels", proxyTo("chats"))
		auth.Any("/channels/*path", proxyTo("chats"))
		auth.Any("/analytics", proxyTo("chats"))
		auth.Any("/analytics/*path", proxyTo("chats"))
		auth.Any("/tenant", proxyTo("users"))
		auth.Any("/tenant/*path", proxyTo("users"))
	}

	// chats 接口
	api.Any("/chats", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), proxyTo("chats"))
	api.Any("/chats/*path", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), proxyTo("chats"))

	// billing 接口
	api.Any("/billing", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), proxyTo("billing"))
	api.Any("/billing/*path", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 1000, time.Minute), proxyTo("billing"))

	// settings / dashboard 等管理接口（普通用户即可访问，user-service 自己做角色校验）
	api.Any("/settings", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyTo("users"))
	api.Any("/settings/*path", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyTo("users"))
	api.Any("/dashboard", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyTo("users"))
	api.Any("/dashboard/*path", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyTo("users"))
	api.Any("/plans", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyTo("billing"))
	api.Any("/plans/*path", middleware.EitherAuthMiddleware(jwtSecret), middleware.RateLimitMiddleware(redisClient, 500, time.Minute), proxyTo("billing"))

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
