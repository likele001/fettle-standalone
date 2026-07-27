package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type DouyinClient struct {
	appID     string
	appSecret string
	httpClient *http.Client
	token     string
	tokenExp  time.Time
	mu        sync.RWMutex
}

type TokenResponse struct {
	ErrNo  int    `json:"err_no"`
	ErrMsg string `json:"err_msg"`
	Data   struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	} `json:"data"`
}

type SendMsgRequest struct {
	OpenID    string `json:"open_id"`
	MsgType   string `json:"msg_type"`
	Content   string `json:"content"`
}

type CommonResponse struct {
	ErrNo  int    `json:"err_no"`
	ErrMsg string `json:"err_msg"`
}

func NewDouyinClient(appID, appSecret string) *DouyinClient {
	return &DouyinClient{
		appID:      appID,
		appSecret:  appSecret,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *DouyinClient) VerifySignature(timestamp, nonce, body, signature string) bool {
	params := []string{c.appSecret, timestamp, nonce}
	sort.Strings(params)
	str := strings.Join(params, "") + body
	mac := hmac.New(sha256.New, []byte(c.appSecret))
	mac.Write([]byte(str))
	expected := hex.EncodeToString(mac.Sum(nil))
	return expected == signature
}

func (c *DouyinClient) SendText(openID, text string) error {
	content, _ := json.Marshal(map[string]string{"text": text})
	return c.sendMessage(openID, "text", string(content))
}

func (c *DouyinClient) sendMessage(openID, msgType, content string) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	req := SendMsgRequest{OpenID: openID, MsgType: msgType, Content: content}
	data, _ := json.Marshal(req)

	url := fmt.Sprintf("https://open.douyin.com/api/im/message/send/?access_token=%s", token)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("send message request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result CommonResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response failed: %w", err)
	}
	if result.ErrNo != 0 {
		return fmt.Errorf("douyin api error: %d - %s", result.ErrNo, result.ErrMsg)
	}
	return nil
}

func (c *DouyinClient) getToken() (string, error) {
	c.mu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		t := c.token
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]string{"app_id": c.appID, "grant_type": "client_credential", "secret": c.appSecret}
	data, _ := json.Marshal(body)

	resp, err := c.httpClient.Post(
		"https://open.douyin.com/oauth/client_token/",
		"application/json", bytes.NewReader(data),
	)
	if err != nil {
		return "", fmt.Errorf("get token request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result TokenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse token response failed: %w", err)
	}
	if result.ErrNo != 0 {
		return "", fmt.Errorf("get token error: %d - %s", result.ErrNo, result.ErrMsg)
	}

	c.token = result.Data.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.Data.ExpiresIn-200) * time.Second)
	return c.token, nil
}
