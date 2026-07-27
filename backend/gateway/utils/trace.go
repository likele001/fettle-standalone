package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceIDKey 上下文中的trace ID键
const TraceIDKey = "trace_id"

// GenerateTraceID 生成追踪ID
func GenerateTraceID() string {
	return uuid.New().String()
}

// GetTraceID 从上下文获取追踪ID
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		return traceID.(string)
	}
	return ""
}

// SetTraceID 设置追踪ID到上下文
func SetTraceID(c *gin.Context, traceID string) {
	c.Set(TraceIDKey, traceID)
}

// EnsureTraceID 确保上下文中有追踪ID
func EnsureTraceID(c *gin.Context) string {
	traceID := GetTraceID(c)
	if traceID == "" {
		traceID = GenerateTraceID()
		SetTraceID(c, traceID)
	}
	return traceID
}
