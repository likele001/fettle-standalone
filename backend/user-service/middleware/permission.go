package middleware

import (
	"net/http"

	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

// NewRequirePermission 创建权限检查中间件
func NewRequirePermission(permSvc *service.PermissionService, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr, _ := role.(string)

		// 超级管理员拥有所有权限
		if roleStr == "super_admin" {
			c.Next()
			return
		}

		tenantID, _ := c.Get("tenant_id")
		tenantIDStr, _ := tenantID.(string)

		// 从数据库查询角色权限
		if tenantIDStr != "" && roleStr != "" {
			hasPerm, err := permSvc.CheckPermission(tenantIDStr, roleStr, permission)
			if err == nil && hasPerm {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "permission denied"})
		c.Abort()
	}
}
