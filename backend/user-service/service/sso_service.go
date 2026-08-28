package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"ai-platform/user-service/models"

	"gorm.io/gorm"
)

// SSOConfig OIDC SSO 配置（环境变量）
type SSOConfig struct {
	Enabled      bool
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scope        string
}

// LoadSSOConfig 从环境变量加载 SSO 配置
func LoadSSOConfig() SSOConfig {
	return SSOConfig{
		Enabled:      os.Getenv("SSO_ENABLED") == "1" || os.Getenv("SSO_ENABLED") == "true",
		Issuer:       os.Getenv("SSO_ISSUER"),
		ClientID:     os.Getenv("SSO_CLIENT_ID"),
		ClientSecret: os.Getenv("SSO_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("SSO_REDIRECT_URI"),
		Scope:        firstNonEmpty(os.Getenv("SSO_SCOPE"), "openid email profile"),
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// SSOService OIDC 授权码模式 SSO（纯 net/http 实现，无外部依赖）
type SSOService struct {
	db         *gorm.DB
	auth       *AuthService
	config     SSOConfig
	httpClient *http.Client
}

func NewSSOService(db *gorm.DB, auth *AuthService) *SSOService {
	return &SSOService{
		db:         db,
		auth:       auth,
		config:     LoadSSOConfig(),
		httpClient: &http.Client{Timeout: 12 * time.Second},
	}
}

func (s *SSOService) Enabled() bool { return s.config.Enabled }

type oidcDiscovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

func (s *SSOService) discovery() (*oidcDiscovery, error) {
	wellKnown := strings.TrimRight(s.config.Issuer, "/") + "/.well-known/openid-configuration"
	resp, err := s.httpClient.Get(wellKnown)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc discovery failed: %d", resp.StatusCode)
	}
	var d oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	if d.AuthorizationEndpoint == "" || d.TokenEndpoint == "" {
		return nil, fmt.Errorf("oidc discovery: missing endpoints")
	}
	return &d, nil
}

// AuthorizeURL 构造 IdP 授权跳转 URL
func (s *SSOService) AuthorizeURL(state string) (string, error) {
	d, err := s.discovery()
	if err != nil {
		return "", err
	}
	u, _ := url.Parse(d.AuthorizationEndpoint)
	q := u.Query()
	q.Set("client_id", s.config.ClientID)
	q.Set("redirect_uri", s.config.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", s.config.Scope)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ExchangeAndLogin 用授权码换 token → 拉取用户信息 → 按 email 匹配平台账号 → 签发 JWT
func (s *SSOService) ExchangeAndLogin(code string) (*TokenResponse, *models.User, error) {
	d, err := s.discovery()
	if err != nil {
		return nil, nil, err
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", s.config.RedirectURI)
	form.Set("client_id", s.config.ClientID)
	form.Set("client_secret", s.config.ClientSecret)
	resp, err := s.httpClient.PostForm(d.TokenEndpoint, form)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("token exchange failed: %d %s", resp.StatusCode, truncate(string(body), 200))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, nil, err
	}
	if tok.AccessToken == "" {
		return nil, nil, fmt.Errorf("token exchange: empty access_token")
	}

	// 拉取用户信息
	req, _ := http.NewRequest(http.MethodGet, d.UserinfoEndpoint, nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	uresp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer uresp.Body.Close()
	ubody, _ := io.ReadAll(uresp.Body)
	if uresp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("userinfo failed: %d", uresp.StatusCode)
	}
	var ui struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(ubody, &ui); err != nil {
		return nil, nil, err
	}
	if ui.Email == "" {
		return nil, nil, fmt.Errorf("SSO 用户无 email，无法匹配平台账号")
	}

	var user models.User
	if err := s.db.Where("email = ?", ui.Email).First(&user).Error; err != nil {
		return nil, nil, fmt.Errorf("SSO 邮箱 %s 未绑定平台账号", ui.Email)
	}
	if user.Status != "active" {
		return nil, nil, fmt.Errorf("账号已停用")
	}

	tokens, err := s.auth.generateTokens(user.ID.String(), user.TenantID.String(), user.Role)
	if err != nil {
		return nil, nil, err
	}
	return tokens, &user, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
