package middleware

import (
	"crypto/hmac"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// localProxyCIDRs 本机回环网段：服务只经本机 nginx / 同机进程访问，故仅信任这些来源的转发头。
var localProxyCIDRs = []string{"127.0.0.0/8", "::1/128"}

// TrustLocalProxies 收紧 gin 的转发头信任范围。
// gin 默认信任所有代理，任何外部请求都能用 X-Forwarded-For 伪装成 127.0.0.1，
// 从而绕过按 IP 的限流和「仅本机可访问」内部端点的兜底判定。
func TrustLocalProxies(r *gin.Engine) error {
	if err := r.SetTrustedProxies(localProxyCIDRs); err != nil {
		return fmt.Errorf("set trusted proxies: %w", err)
	}
	return nil
}

// MustTrustLocalProxies 在启动阶段设置可信代理，失败即 panic（配置错误应当阻断启动而非静默降级）。
func MustTrustLocalProxies(r *gin.Engine) {
	if err := TrustLocalProxies(r); err != nil {
		panic(err)
	}
}

// DirectLoopback 判断连接是否真正来自本机：RemoteAddr 由内核确定，不受 X-Forwarded-For 影响。
func DirectLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// InternalAuthFromToken 内部端点鉴权：优先校验 X-Internal-Token；
// 未配置 token 时退化为仅放行真正的本机连接（判定基于 RemoteAddr，调用方仍需设置可信代理以保护其它依赖 ClientIP 的逻辑）。
func InternalAuthFromToken(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expected != "" {
			if !hmac.Equal([]byte(c.GetHeader("X-Internal-Token")), []byte(expected)) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "invalid internal token"})
				return
			}
			c.Next()
			return
		}
		if !DirectLoopback(c.Request.RemoteAddr) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 1003, "message": "internal endpoint: local only"})
			return
		}
		c.Next()
	}
}
