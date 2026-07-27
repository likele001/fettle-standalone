package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type SysConfig struct {
	ID        uint            `gorm:"primary_key" json:"id"`
	Key       string          `gorm:"size:100;unique" json:"key"`
	Config    PlatformSettings `gorm:"type:jsonb" json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type PlatformSettings struct {
	SiteName        string            `json:"site_name"`
	SiteLogo        string            `json:"site_logo"`
	ContactEmail    string            `json:"contact_email"`
	SmsProvider     string            `json:"sms_provider"`
	SmsConfig       map[string]string `json:"sms_config"`
	EmailProvider   string            `json:"email_provider"`
	EmailConfig     map[string]string `json:"email_config"`
	MaxUploadSize   int               `json:"max_upload_size"`
	AllowedFileTypes []string         `json:"allowed_file_types"`
	MaintenanceMode bool              `json:"maintenance_mode"`
	// 微信小程序配置
	WechatMiniAppId     string `json:"wechat_miniapp_id"`
	WechatMiniAppSecret string `json:"wechat_miniapp_secret"`
}

// Value 实现 driver.Valuer 接口，序列化为 JSONB
func (p PlatformSettings) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口，从 JSONB 反序列化
func (p *PlatformSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return fmt.Errorf("unsupported type for PlatformSettings: %T", value)
	}
}

func (SysConfig) TableName() string {
	return "sys_config"
}
