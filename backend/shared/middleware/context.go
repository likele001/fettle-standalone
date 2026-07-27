package middleware

import "github.com/gin-gonic/gin"

// GetTenantID 安全获取租户ID
func GetTenantID(c *gin.Context) (string, bool) {
	val, exists := c.Get("tenant_id")
	if !exists {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// GetUserID 安全获取用户ID
func GetUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// GetRole 安全获取用户角色
func GetRole(c *gin.Context) (string, bool) {
	val, exists := c.Get("role")
	if !exists {
		return "", false
	}
 role, ok := val.(string)
	return role, ok
}
