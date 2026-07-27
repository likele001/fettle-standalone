package service

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"ai-platform/billing-service/models"
	"ai-platform/billing-service/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentService struct {
	repo *repository.BillingRepository
	db   *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		repo: repository.NewBillingRepository(db),
		db:   db,
	}
}

// XunhupayCreateRequest 虎皮椒创建支付请求
type XunhupayCreateRequest struct {
	Version      string  `json:"version"`
	AppID        string  `json:"appid"`
	TradeOrderID string  `json:"trade_order_id"`
	TotalFee     float64 `json:"total_fee"`
	Title        string  `json:"title"`
	Time         int64   `json:"time"`
	NotifyURL    string  `json:"notify_url"`
	ReturnURL    string  `json:"return_url"`
	NonceStr     string  `json:"nonce_str"`
	Hash         string  `json:"hash"`
}

// XunhupayCreateResponse 虎皮椒创建支付响应
type XunhupayCreateResponse struct {
	OpenID  string  `json:"openid"` // 历史遗留，实际是 orderid
	URL     string  `json:"url"`
	URLQR   string  `json:"url_qrcode"`
	ErrCode int     `json:"errcode"`
	ErrMsg  string  `json:"errmsg"`
	Hash    string  `json:"hash"`
}

// XunhupayNotify 虎皮椒回调参数
type XunhupayNotify struct {
	TradeOrderID  string `json:"trade_order_id" form:"trade_order_id"`
	TotalFee      string `json:"total_fee" form:"total_fee"`
	TransactionID string `json:"transaction_id" form:"transaction_id"`
	OpenOrderID   string `json:"open_order_id" form:"open_order_id"`
	OrderTitle    string `json:"order_title" form:"order_title"`
	Status        string `json:"status" form:"status"` // OD=已支付, CD=已退款, RD=退款中, UD=退款失败
	Plugins       string `json:"plugins" form:"plugins"`
	Attach        string `json:"attach" form:"attach"`
	AppID         string `json:"appid" form:"appid"`
	Time          string `json:"time" form:"time"`
	NonceStr      string `json:"nonce_str" form:"nonce_str"`
	Hash          string `json:"hash" form:"hash"`
}

// GetPaymentConfig 获取支付配置
func (s *PaymentService) GetPaymentConfig() (*models.PaymentConfig, error) {
	config, err := s.repo.GetPaymentConfig()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.PaymentConfig{}, nil
		}
		return nil, err
	}
	return config, nil
}

// SavePaymentConfig 保存支付配置
func (s *PaymentService) SavePaymentConfig(config *models.PaymentConfig) error {
	config.UpdatedAt = time.Now()
	return s.repo.SavePaymentConfig(config)
}

// GenerateHash 生成虎皮椒签名
func GenerateHash(params map[string]string, secret string) string {
	// 按 key 字典序排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接 key=value&key=value...
	var parts []string
	for _, k := range keys {
		if k == "hash" || params[k] == "" {
			continue
		}
		parts = append(parts, k+"="+params[k])
	}
	str := strings.Join(parts, "&") + secret

	// MD5
	hash := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", hash)
}

// CreatePaymentOrder 创建支付订单
func (s *PaymentService) CreatePaymentOrder(tenantID uuid.UUID, planID uuid.UUID) (*models.PaymentOrder, error) {
	// 获取支付配置
	config, err := s.repo.GetPaymentConfig()
	if err != nil {
		return nil, errors.New("payment not configured")
	}
	if !config.Enabled {
		return nil, errors.New("payment is disabled")
	}

	// 使用配置中的 return_url
	returnURL := config.ReturnURL
	if returnURL == "" {
		returnURL = "https://fettle.cenkor.cn/billing?payment=success"
	}

	// 获取套餐信息
	plan, err := s.repo.GetPlan(planID)
	if err != nil {
		return nil, err
	}
	if plan.PriceMonthly <= 0 {
		return nil, errors.New("plan price must be greater than 0")
	}

	// 生成商户订单号
	tradeOrderID := fmt.Sprintf("FETTLE_%d_%s", time.Now().Unix(), uuid.New().String()[:8])

	// 构建虎皮椒请求参数
	params := map[string]string{
		"version":        "1.1",
		"appid":          config.AppID,
		"trade_order_id": tradeOrderID,
		"total_fee":      fmt.Sprintf("%.2f", plan.PriceMonthly),
		"title":          fmt.Sprintf("AI智能体平台-%s套餐", plan.Name),
		"time":           fmt.Sprintf("%d", time.Now().Unix()),
		"notify_url":     config.NotifyURL,
		"return_url":     returnURL,
		"nonce_str":      fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	// 生成签名
	params["hash"] = GenerateHash(params, config.AppSecret)

	// 调用虎皮椒 API
	resp, err := callXunhupay(config.GatewayURL, params)
	if err != nil {
		return nil, fmt.Errorf("call xunhupay failed: %w", err)
	}
	if resp.ErrCode != 0 {
		return nil, fmt.Errorf("xunhupay error: %s", resp.ErrMsg)
	}

	// 创建订单记录
	order := &models.PaymentOrder{
		TenantID:     tenantID,
		PlanID:       plan.ID.String(),
		PlanName:     plan.Name,
		Amount:       plan.PriceMonthly,
		TradeOrderID: tradeOrderID,
		Status:       "pending",
		PaymentURL:   resp.URL,
	}

	if err := s.repo.CreatePaymentOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

// HandleNotify 处理虎皮椒回调
func (s *PaymentService) HandleNotify(notify *XunhupayNotify) (string, error) {
	// 验证签名
	config, err := s.repo.GetPaymentConfig()
	if err != nil {
		return "", errors.New("payment not configured")
	}

	params := map[string]string{
		"trade_order_id":  notify.TradeOrderID,
		"total_fee":       notify.TotalFee,
		"transaction_id":  notify.TransactionID,
		"open_order_id":   notify.OpenOrderID,
		"order_title":     notify.OrderTitle,
		"status":          notify.Status,
		"appid":           notify.AppID,
		"time":            notify.Time,
		"nonce_str":       notify.NonceStr,
	}
	expectedHash := GenerateHash(params, config.AppSecret)

	if expectedHash != notify.Hash {
		return "", errors.New("invalid sign")
	}

	// 查找订单
	order, err := s.repo.GetPaymentOrderByTradeID(notify.TradeOrderID)
	if err != nil {
		return "", errors.New("order not found")
	}

	// 已处理过
	if order.Status == "paid" {
		return "success", nil
	}

	// 更新订单状态
	now := time.Now()
	order.TransactionID = notify.TransactionID
	order.PaidAt = &now

	if notify.Status == "OD" {
		order.Status = "paid"
	} else if notify.Status == "CD" {
		order.Status = "cancelled"
	} else {
		order.Status = "failed"
	}

	if err := s.repo.UpdatePaymentOrder(order); err != nil {
		return "", err
	}

	// 支付成功，激活订阅
	if notify.Status == "OD" {
		tenantID, _ := uuid.Parse(order.TenantID.String())
		planID, _ := uuid.Parse(order.PlanID)
		if err := s.activateSubscription(tenantID, planID); err != nil {
			return "", fmt.Errorf("activate subscription failed: %w", err)
		}
	}

	return "success", nil
}

// activateSubscription 激活订阅
func (s *PaymentService) activateSubscription(tenantID, planID uuid.UUID) error {
	plan, err := s.repo.GetPlan(planID)
	if err != nil {
		return err
	}

	sub, err := s.repo.GetSubscription(tenantID)
	if err != nil {
		subscription := &models.Subscription{
			TenantID:           tenantID,
			PlanID:             planID,
			Status:             "active",
			StartDate:          time.Now(),
			EndDate:            time.Now().AddDate(0, 1, 0),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
			TokenLimit:         int(plan.MaxMessagesPerMonth),
		}
		return s.repo.CreateSubscription(subscription)
	}

	sub.PlanID = planID
	sub.Status = "active"
	sub.StartDate = time.Now()
	sub.EndDate = time.Now().AddDate(0, 1, 0)
	sub.CurrentPeriodStart = time.Now()
	sub.CurrentPeriodEnd = time.Now().AddDate(0, 1, 0)
	sub.MessagesUsed = 0
	sub.TokenLimit = int(plan.MaxMessagesPerMonth)
	return s.repo.UpdateSubscription(sub)
}

// GetPaymentOrders 获取支付订单列表
func (s *PaymentService) GetPaymentOrders(tenantID uuid.UUID, status string, page, pageSize int) ([]models.PaymentOrder, int64, error) {
	return s.repo.ListPaymentOrders(tenantID, status, page, pageSize)
}

// GetPaymentOrder 获取单个支付订单
func (s *PaymentService) GetPaymentOrder(id uuid.UUID) (*models.PaymentOrder, error) {
	return s.repo.GetPaymentOrder(id)
}

// callXunhupay 调用虎皮椒 API
func callXunhupay(gatewayURL string, params map[string]string) (*XunhupayCreateResponse, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(gatewayURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result XunhupayCreateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %s", string(body))
	}

	return &result, nil
}
