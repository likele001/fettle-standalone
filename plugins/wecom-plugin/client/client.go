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

type WeComClient struct {
	corpID     string
	corpSecret string
	agentID    int
	httpClient *http.Client
	token      string
	tokenExp   time.Time
	mu         sync.RWMutex
}

type TokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type SendMsgRequest struct {
	ToUser  string `json:"touser"`
	MsgType string `json:"msgtype"`
	AgentID int    `json:"agentid"`
	Text    *TextContent `json:"text,omitempty"`
	Markdown *MarkdownContent `json:"markdown,omitempty"`
}

type TextContent struct {
	Content string `json:"content"`
}

type MarkdownContent struct {
	Content string `json:"content"`
}

type CommonResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func NewWeComClient(corpID, corpSecret string, agentID int) *WeComClient {
	return &WeComClient{
		corpID:     corpID,
		corpSecret: corpSecret,
		agentID:    agentID,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *WeComClient) SendText(toUser, content string) error {
	req := SendMsgRequest{
		ToUser:  toUser,
		MsgType: "text",
		AgentID: c.agentID,
		Text:    &TextContent{Content: content},
	}
	return c.sendMessage(req)
}

func (c *WeComClient) SendMarkdown(toUser, content string) error {
	req := SendMsgRequest{
		ToUser:   toUser,
		MsgType:  "markdown",
		AgentID:  c.agentID,
		Markdown: &MarkdownContent{Content: content},
	}
	return c.sendMessage(req)
}

func (c *WeComClient) sendMessage(req SendMsgRequest) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	data, _ := json.Marshal(req)
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
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
	if result.ErrCode != 0 {
		return fmt.Errorf("wecom api error: %d - %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

func (c *WeComClient) getToken() (string, error) {
	c.mu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		t := c.token
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", c.corpID, c.corpSecret)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("get token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result TokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse token response failed: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("get token error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	c.token = result.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.ExpiresIn-200) * time.Second)
	return c.token, nil
}
