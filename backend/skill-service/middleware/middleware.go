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
var CORS = middleware.CORSMiddleware
var TraceMiddleware = middleware.TraceMiddleware
var MustTrustLocalProxies = middleware.MustTrustLocalProxies

// Auth 别名，兼容 billing/skill 服务的调用方式
var Auth = middleware.JWTAuthMiddleware


// InternalAuth 内部服务鉴权：优先 X-Internal-Token（与 billing 共用 BILLING_INTERNAL_TOKEN）；未配置时仅放行真正的本机连接
func InternalAuth() gin.HandlerFunc {
	return middleware.InternalAuthFromToken(os.Getenv("BILLING_INTERNAL_TOKEN"))
}
