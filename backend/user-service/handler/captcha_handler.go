package handler

import (
	"ai-platform/shared/response"
	"ai-platform/user-service/service"

	"github.com/gin-gonic/gin"
)

// CaptchaHandler 验证码处理器
type CaptchaHandler struct {
	captchaService *service.CaptchaService
}

// NewCaptchaHandler 创建验证码处理器
func NewCaptchaHandler(cs *service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{captchaService: cs}
}

// GenerateCaptcha 生成图形验证码
func (h *CaptchaHandler) GenerateCaptcha(c *gin.Context) {
	id, b64, err := h.captchaService.Generate()
	if err != nil {
		response.Error(c, 5000, "生成验证码失败")
		return
	}
	response.Success(c, gin.H{
		"captcha_id":    id,
		"captcha_image": b64,
	})
}
