package middleware

import "github.com/gin-gonic/gin"

// TenantMiddleware 租户中间件
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("tenant_id"); !exists {
			c.Set("tenant_id", "")
		}
		c.Next()
	}
}
