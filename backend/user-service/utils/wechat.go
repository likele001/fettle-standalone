package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// WechatConfig 微信配置
type WechatConfig struct {
	AppID       string
	AppSecret   string
	RedirectURI string // OAuth回调地址
}

// WechatClient 微信客户端
type WechatClient struct {
	config *WechatConfig
	client *http.Client
}

// NewWechatClient 创建微信客户端
func NewWechatClient(config *WechatConfig) *WechatClient {
	return &WechatClient{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// OAuthUserInfo OAuth用户信息
type OAuthUserInfo struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	UnionID    string `json:"unionid"`
}

// AccessTokenResponse 获取AccessToken响应
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// OAuthTokenResponse OAuth Token响应
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// GetOAuthURL 获取OAuth授权URL
func (c *WechatClient) GetOAuthURL(state string) string {
	params := url.Values{}
	params.Add("appid", c.config.AppID)
	params.Add("redirect_uri", c.config.RedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "snsapi_userinfo")
	params.Add("state", state)

	return fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?%s#wechat_redirect", params.Encode())
}

// GetOAuthToken 通过授权码获取OAuth Token
func (c *WechatClient) GetOAuthToken(code string) (*OAuthTokenResponse, error) {
	params := url.Values{}
	params.Add("appid", c.config.AppID)
	params.Add("secret", c.config.AppSecret)
	params.Add("code", code)
	params.Add("grant_type", "authorization_code")

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?%s", params.Encode())

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result OAuthTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("WeChat API error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// GetUserInfo 获取用户信息
func (c *WechatClient) GetUserInfo(accessToken, openID string) (*OAuthUserInfo, error) {
	params := url.Values{}
	params.Add("access_token", accessToken)
	params.Add("openid", openID)
	params.Add("lang", "zh_CN")

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?%s", params.Encode())

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result OAuthUserInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.OpenID == "" {
		return nil, errors.New("invalid user info response")
	}

	return &result, nil
}

// GetAccessToken 获取小程序AccessToken
func (c *WechatClient) GetAccessToken() (*AccessTokenResponse, error) {
	params := url.Values{}
	params.Add("grant_type", "client_credential")
	params.Add("appid", c.config.AppID)
	params.Add("secret", c.config.AppSecret)

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?%s", params.Encode())

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result AccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("WeChat API error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// Code2Session 小程序登录，通过code换取openid和session_key
func (c *WechatClient) Code2Session(code string) (openID, sessionKey string, err error) {
	params := url.Values{}
	params.Add("appid", c.config.AppID)
	params.Add("secret", c.config.AppSecret)
	params.Add("js_code", code)
	params.Add("grant_type", "authorization_code")

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?%s", params.Encode())

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return "", "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read response failed: %w", err)
	}

	var result struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return "", "", fmt.Errorf("WeChat API error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return result.OpenID, result.SessionKey, nil
}

// DecryptUserInfo 解密用户信息（小程序）
// 参考微信加密数据解密算法: https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/signature.html
func (c *WechatClient) DecryptUserInfo(encryptedData, iv, sessionKey string) (*OAuthUserInfo, error) {
	sessionKeyBytes, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("decode session key failed: %w", err)
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted data failed: %w", err)
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("decode iv failed: %w", err)
	}

	block, err := aes.NewCipher(sessionKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("new cipher failed: %w", err)
	}

	if len(encryptedBytes) < aes.BlockSize {
		return nil, errors.New("encrypted data too short")
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	decrypted := make([]byte, len(encryptedBytes))
	mode.CryptBlocks(decrypted, encryptedBytes)

	// PKCS7 unpadding
	padding := int(decrypted[len(decrypted)-1])
	if padding > aes.BlockSize || padding > len(decrypted) {
		return nil, errors.New("invalid padding")
	}
	decrypted = decrypted[:len(decrypted)-padding]

	var result struct {
		OpenID    string `json:"openId"`
		Nickname  string `json:"nickName"`
		Sex       int    `json:"gender"`
		Province  string `json:"province"`
		City      string `json:"city"`
		Country   string `json:"country"`
		AvatarURL string `json:"avatarUrl"`
		UnionID   string `json:"unionId"`
		Watermark struct {
			AppID string `json:"appid"`
		} `json:"watermark"`
	}

	if err := json.Unmarshal(decrypted, &result); err != nil {
		return nil, fmt.Errorf("unmarshal decrypted data failed: %w", err)
	}

	if result.Watermark.AppID != c.config.AppID {
		return nil, errors.New("watermark appid mismatch")
	}

	return &OAuthUserInfo{
		OpenID:     result.OpenID,
		Nickname:   result.Nickname,
		Sex:        result.Sex,
		Province:   result.Province,
		City:       result.City,
		Country:    result.Country,
		HeadImgURL: result.AvatarURL,
		UnionID:    result.UnionID,
	}, nil
}
