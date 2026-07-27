package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceMiddleware 追踪中间件
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set("trace_id", traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)
		c.Next()
	}
}
