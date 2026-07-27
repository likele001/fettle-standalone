package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type FeishuClient struct {
	appID     string
	appSecret string
	httpClient *http.Client
	token     string
	tokenExp  time.Time
	mu        sync.RWMutex
}

type TenantTokenResponse struct {
	Code              int    `json:"code"`
	Msg              string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire           int    `json:"expire"`
}

type SendMsgRequest struct {
	ReceiveID string      `json:"receive_id"`
	MsgType   string      `json:"msg_type"`
	Content   string      `json:"content"`
}

type CommonResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewFeishuClient(appID, appSecret string) *FeishuClient {
	return &FeishuClient{
		appID:      appID,
		appSecret:  appSecret,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *FeishuClient) SendText(openID, text string) error {
	content, _ := json.Marshal(map[string]string{"text": text})
	return c.sendMessage(openID, "text", string(content))
}

func (c *FeishuClient) SendInteractive(openID string, card map[string]interface{}) error {
	content, _ := json.Marshal(card)
	return c.sendMessage(openID, "interactive", string(content))
}

func (c *FeishuClient) sendMessage(receiveID, msgType, content string) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	req := SendMsgRequest{
		ReceiveID: receiveID,
		MsgType:   msgType,
		Content:   content,
	}
	data, _ := json.Marshal(req)

	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id")
	httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send message request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result CommonResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response failed: %w", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("feishu api error: %d - %s", result.Code, result.Msg)
	}
	return nil
}

func (c *FeishuClient) getToken() (string, error) {
	c.mu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		t := c.token
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]string{"app_id": c.appID, "app_secret": c.appSecret}
	data, _ := json.Marshal(body)

	resp, err := c.httpClient.Post(
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		"application/json", bytes.NewReader(data),
	)
	if err != nil {
		return "", fmt.Errorf("get token request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result TenantTokenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse token response failed: %w", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("get token error: %d - %s", result.Code, result.Msg)
	}

	c.token = result.TenantAccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.Expire-200) * time.Second)
	return c.token, nil
}
