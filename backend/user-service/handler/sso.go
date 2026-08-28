package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

// SSOHandler OIDC SSO 登录
type SSOHandler struct {
	service *service.SSOService
}

func NewSSOHandler(svc *service.SSOService) *SSOHandler {
	return &SSOHandler{service: svc}
}

// Authorize 跳转 IdP 授权页（302）
func (h *SSOHandler) Authorize(c *gin.Context) {
	if !h.service.Enabled() {
		c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": "SSO not enabled"})
		return
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := hex.EncodeToString(b)
	c.SetCookie("sso_state", state, 600, "/", "", false, true)
	u, err := h.service.AuthorizeURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, u)
}

// Callback 处理 IdP 回调：换 token → 匹配账号 → 返回 JWT
func (h *SSOHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "missing code"})
		return
	}
	if ck, err := c.Cookie("sso_state"); err == nil && ck != "" && ck != state {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "invalid state"})
		return
	}
	tokens, user, err := h.service.ExchangeAndLogin(code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"tokens": tokens,
			"user":   user,
		},
	})
}
