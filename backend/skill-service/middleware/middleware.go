package middleware

import "ai-platform/shared/middleware"

type Claims = middleware.Claims

var JWTAuthMiddleware = middleware.JWTAuthMiddleware
var TenantMiddleware = middleware.TenantMiddleware
var LoggerMiddleware = middleware.LoggerMiddleware
var CORSMiddleware = middleware.CORSMiddleware
var CORS = middleware.CORSMiddleware
var TraceMiddleware = middleware.TraceMiddleware

// Auth 别名，兼容 billing/skill 服务的调用方式
var Auth = middleware.JWTAuthMiddleware
