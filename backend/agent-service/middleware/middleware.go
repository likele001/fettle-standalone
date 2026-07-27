package middleware

import "ai-platform/shared/middleware"

type Claims = middleware.Claims

var JWTAuthMiddleware = middleware.JWTAuthMiddleware
var TenantMiddleware = middleware.TenantMiddleware
var LoggerMiddleware = middleware.LoggerMiddleware
var CORSMiddleware = middleware.CORSMiddleware
var TraceMiddleware = middleware.TraceMiddleware
