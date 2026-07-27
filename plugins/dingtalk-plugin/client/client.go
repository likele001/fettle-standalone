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

type DingTalkClient struct {
	appKey     string
	appSecret  string
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
	AgentID  int64  `json:"agent_id"`
	UserIDList string `json:"userid_list"`
	Msg      MsgContent `json:"msg"`
}

type MsgContent struct {
	MsgType string `json:"msgtype"`
	Text    *TextContent `json:"text,omitempty"`
	Markdown *MarkdownContent `json:"markdown,omitempty"`
}

type TextContent struct {
	Content string `json:"content"`
}

type MarkdownContent struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type CommonResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func NewDingTalkClient(appKey, appSecret string) *DingTalkClient {
	return &DingTalkClient{
		appKey:     appKey,
		appSecret:  appSecret,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *DingTalkClient) SendText(agentID int64, userIDList, text string) error {
	req := SendMsgRequest{
		AgentID:    agentID,
		UserIDList: userIDList,
		Msg: MsgContent{
			MsgType: "text",
			Text:    &TextContent{Content: text},
		},
	}
	return c.sendMessage(req)
}

func (c *DingTalkClient) sendMessage(req SendMsgRequest) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	data, _ := json.Marshal(req)
	url := fmt.Sprintf("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2?access_token=%s", token)
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
		return fmt.Errorf("dingtalk api error: %d - %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

func (c *DingTalkClient) getToken() (string, error) {
	c.mu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		t := c.token
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]string{"appkey": c.appKey, "appsecret": c.appSecret}
	data, _ := json.Marshal(body)

	resp, err := c.httpClient.Post(
		"https://oapi.dingtalk.com/gettoken",
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
	if result.ErrCode != 0 {
		return "", fmt.Errorf("get token error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	c.token = result.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.ExpiresIn-200) * time.Second)
	return c.token, nil
}
