package service

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WechatAPI 微信 API 客户端
type WechatAPI struct {
	appID     string
	appSecret string
	token     string
}

// NewWechatAPI 创建微信 API 客户端
func NewWechatAPI(appID, appSecret, token string) *WechatAPI {
	return &WechatAPI{
		appID:     appID,
		appSecret: appSecret,
		token:     token,
	}
}

// AccessTokenResponse 访问令牌响应
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// WechatMessage 微信消息
type WechatMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
	MsgID        int64    `xml:"MsgId"`
}

// ReplyMessage 回复消息
type ReplyMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
}

// GetAccessToken 获取访问令牌
func (api *WechatAPI) GetAccessToken() (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		api.appID, api.appSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result AccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat api error: %s", result.ErrMsg)
	}

	return result.AccessToken, nil
}

// SendTextMessage 发送文本消息
func (api *WechatAPI) SendTextMessage(toUser, content string) error {
	accessToken, err := api.GetAccessToken()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", accessToken)

	msg := map[string]interface{}{
		"touser":  toUser,
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if errCode, ok := result["errcode"].(float64); ok && errCode != 0 {
		return fmt.Errorf("send message error: %v", result["errmsg"])
	}

	return nil
}

// ParseMessage 解析微信消息
func ParseMessage(xmlData []byte) (*WechatMessage, error) {
	var msg WechatMessage
	if err := xml.Unmarshal(xmlData, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// BuildReplyMessage 构建回复消息
func BuildReplyMessage(toUser, fromUser, content string) []byte {
	reply := ReplyMessage{
		ToUserName:   toUser,
		FromUserName: fromUser,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      content,
	}

	xmlData, _ := xml.Marshal(reply)
	return xmlData
}
