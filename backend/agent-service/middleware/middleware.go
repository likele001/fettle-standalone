package middleware

import (
	"os"

	"ai-platform/shared/middleware"

	"github.com/gin-gonic/gin"
)

type Claims = middleware.Claims

var JWTAuthMiddleware = middleware.JWTAuthMiddleware
var TenantMiddleware = middleware.TenantMiddleware
var LoggerMiddleware = middleware.LoggerMiddleware
var CORSMiddleware = middleware.CORSMiddleware
var TraceMiddleware = middleware.TraceMiddleware
var MustTrustLocalProxies = middleware.MustTrustLocalProxies

// InternalAuth 内部服务鉴权：优先 X-Internal-Token（与 billing 共用 BILLING_INTERNAL_TOKEN）；未配置时仅放行真正的本机连接
func InternalAuth() gin.HandlerFunc {
	return middleware.InternalAuthFromToken(os.Getenv("BILLING_INTERNAL_TOKEN"))
}
