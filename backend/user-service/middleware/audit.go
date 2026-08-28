package middleware

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AdminAuditLog 平台管理员操作审计日志（对应 admin_audit_log 表）
type AdminAuditLog struct {
	ID         uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AdminID    string          `gorm:"type:uuid" json:"admin_id"`
	Action     string          `gorm:"size:100" json:"action"`
	TargetType string          `gorm:"size:50" json:"target_type"`
	TargetID   string          `gorm:"type:uuid" json:"target_id"`
	Detail     json.RawMessage `gorm:"type:jsonb" json:"detail"`
	IPAddress  string          `gorm:"size:64" json:"ip_address"`
	CreatedAt  time.Time       `json:"created_at"`
}

// TableName 表名
func (AdminAuditLog) TableName() string {
	return "admin_audit_log"
}

// 写操作白名单（仅审计写操作）；GET/HEAD/OPTIONS 不记录
var writeMethods = map[string]bool{"POST": true, "PUT": true, "PATCH": true, "DELETE": true}

// 审计缓冲通道，异步落库，避免阻塞业务请求
var auditBuf = make(chan *AdminAuditLog, 1024)
var auditOnce sync.Once

// startAuditWorker 启动后台落库 worker（幂等）
func startAuditWorker(db *gorm.DB) {
	auditOnce.Do(func() {
		go func() {
			for entry := range auditBuf {
				if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(entry).Error; err != nil {
					// 落库失败不阻塞，仅记录（避免影响主流程）
					continue
				}
			}
		}()
	})
}

// AuditMiddleware 平台管理员操作审计。
// 用于 admin 路由组，在认证之后注册；仅记录写操作。
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	startAuditWorker(db)
	return func(c *gin.Context) {
		if !writeMethods[c.Request.Method] {
			c.Next()
			return
		}

		// 提取管理员身份
		adminID, _ := c.Get("user_id")
		role, _ := c.Get("role")

		// 构造目标类型/ID：/admin/tenants/:id -> target=tentants, target_id=:id
		targetType, targetID := parseTarget(c)

		c.Next()

		// 仅在被拒绝（非 2xx）时不记录；2xx 写操作一律记录
		if c.Writer.Status() >= 300 {
			return
		}

		action := strings.ToLower(c.Request.Method) + ":" + targetType

		detail, _ := json.Marshal(map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"uri":    c.Request.RequestURI,
			"role":   role,
		})

		entry := &AdminAuditLog{
			AdminID:    adminIDStr(adminID),
			Action:     action,
			TargetType: targetType,
			TargetID:   targetID,
			Detail:     detail,
			IPAddress:  c.ClientIP(),
			CreatedAt:  time.Now(),
		}

		// 非阻塞投递，通道满则丢弃（审计尽力而为）
		select {
		case auditBuf <- entry:
		default:
		}
	}
}

func adminIDStr(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}
	return s
}

// parseTarget 从请求路径解析 资源类型 / 资源ID
func parseTarget(c *gin.Context) (string, string) {
	// /admin/tenants/:id
	parts := splitPath(c.FullPath())
	for i, p := range parts {
		if p == ":id" && i > 0 {
			return parts[i-1], c.Param("id")
		}
	}
	// 形如 /admin/tenants 的集合：类型为最后一段，无ID
	if len(parts) >= 2 {
		return parts[len(parts)-1], ""
	}
	return "unknown", ""
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}