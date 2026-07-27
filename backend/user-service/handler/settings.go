package handler

import (
	"ai-platform/shared/response"
	"ai-platform/user-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	db *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	var config models.SysConfig
	if err := h.db.First(&config, "id = ?", 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Success(c, models.PlatformSettings{
				SiteName:          "Fettle AI Platform",
				ContactEmail:      "support@fettle.com",
				SmsProvider:       "",
				SmsConfig:         map[string]string{},
				EmailProvider:     "",
				EmailConfig:       map[string]string{},
				MaxUploadSize:     50,
				AllowedFileTypes:  []string{"jpg", "jpeg", "png", "pdf", "docx"},
				MaintenanceMode:   false,
			})
			return
		}
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, config.Config)
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var settings models.PlatformSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	var config models.SysConfig
	if err := h.db.First(&config, "id = ?", 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			config = models.SysConfig{
				ID:     1,
				Config: settings,
			}
			if err := h.db.Create(&config).Error; err != nil {
				response.Error(c, 5000, err.Error())
				return
			}
		} else {
			response.Error(c, 5000, err.Error())
			return
		}
	} else {
		config.Config = settings
		if err := h.db.Save(&config).Error; err != nil {
			response.Error(c, 5000, err.Error())
			return
		}
	}

	response.Success(c, settings)
}

func (h *SettingsHandler) TestEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	response.Success(c, gin.H{
		"success": true,
		"message": "邮件发送测试成功 (模拟)",
	})
}

func (h *SettingsHandler) TestSms(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	response.Success(c, gin.H{
		"success": true,
		"message": "短信发送测试成功 (模拟)",
	})
}