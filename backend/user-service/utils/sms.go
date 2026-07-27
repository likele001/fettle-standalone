package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SMSConfig 短信服务配置
type SMSConfig struct {
	Provider    string // 短信服务商：aliyun, tencent
	AccessKey   string
	SecretKey   string
	SignName    string // 签名
	TemplateID  string // 模板ID
	CallbackURL string // 回调地址（可选）
}

// SMSClient 短信客户端
type SMSClient struct {
	config *SMSConfig
	client *http.Client
}

// NewSMSClient 创建短信客户端
func NewSMSClient(config *SMSConfig) *SMSClient {
	return &SMSClient{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendSMSRequest 发送短信请求
type SendSMSRequest struct {
	Phone      string            // 手机号
	TemplateID string            // 模板ID
	Params     map[string]string // 模板参数
}

// SendSMSResponse 发送短信响应
type SendSMSResponse struct {
	MessageID string // 消息ID
	Status    string // 发送状态
}

// SendVerificationCode 发送验证码
func (c *SMSClient) SendVerificationCode(phone, code string) error {
	req := &SendSMSRequest{
		Phone:      phone,
		TemplateID: c.config.TemplateID,
		Params: map[string]string{
			"code":      code,
			"expire_at": "5", // 5分钟过期
		},
	}

	_, err := c.SendSMS(req)
	return err
}

// SendSMS 发送短信
func (c *SMSClient) SendSMS(req *SendSMSRequest) (*SendSMSResponse, error) {
	switch c.config.Provider {
	case "aliyun":
		return c.sendAliyunSMS(req)
	case "tencent":
		return c.sendTencentSMS(req)
	default:
		return nil, errors.New("unsupported SMS provider")
	}
}

// sendAliyunSMS 阿里云短信发送
func (c *SMSClient) sendAliyunSMS(req *SendSMSRequest) (*SendSMSResponse, error) {
	// 阿里云短信API调用
	// 实际实现需要参考阿里云SDK文档
	// 这里提供框架代码

	params := map[string]interface{}{
		"PhoneNumbers":  req.Phone,
		"SignName":      c.config.SignName,
		"TemplateCode":  req.TemplateID,
		"TemplateParam": req.Params,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params failed: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://dysmsapi.aliyuncs.com", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessKey))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result struct {
		MessageID string `json:"MessageId"`
		Code      string `json:"Code"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.Code != "OK" {
		return nil, fmt.Errorf("SMS send failed with code: %s", result.Code)
	}

	return &SendSMSResponse{
		MessageID: result.MessageID,
		Status:    "success",
	}, nil
}

// sendTencentSMS 腾讯云短信发送
func (c *SMSClient) sendTencentSMS(req *SendSMSRequest) (*SendSMSResponse, error) {
	// 腾讯云短信API调用
	// 实际实现需要参考腾讯云SDK文档
	// 这里提供框架代码

	params := map[string]interface{}{
		"PhoneNumberSet": []string{req.Phone},
		"SmsSdkAppId":    c.config.AccessKey,
		"SignName":       c.config.SignName,
		"TemplateId":     req.TemplateID,
		"TemplateParamSet": func() []string {
			var result []string
			for _, v := range req.Params {
				result = append(result, v)
			}
			return result
		}(),
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params failed: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://sms.tencentcloudapi.com", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessKey))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result struct {
		SendStatusSet []struct {
			SerialNo string `json:"SerialNo"`
			Code     string `json:"Code"`
		} `json:"SendStatusSet"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if len(result.SendStatusSet) == 0 {
		return nil, errors.New("no send status in response")
	}

	if result.SendStatusSet[0].Code != "Ok" {
		return nil, fmt.Errorf("SMS send failed with code: %s", result.SendStatusSet[0].Code)
	}

	return &SendSMSResponse{
		MessageID: result.SendStatusSet[0].SerialNo,
		Status:    "success",
	}, nil
}

// GenerateVerificationCode 生成验证码
func GenerateVerificationCode(length int) string {
	if length <= 0 {
		length = 6
	}
	// 生成随机数字验证码
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = '0' + byte(time.Now().UnixNano()%10)
		time.Sleep(time.Millisecond) // 确保随机性
	}
	return string(code)
}
