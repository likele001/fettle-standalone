package repository

import (
	"ai-platform/billing-service/models"

	"github.com/google/uuid"
)

// GetPaymentConfig 获取支付配置
func (r *BillingRepository) GetPaymentConfig() (*models.PaymentConfig, error) {
	var config models.PaymentConfig
	err := r.db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// SavePaymentConfig 保存支付配置
func (r *BillingRepository) SavePaymentConfig(config *models.PaymentConfig) error {
	if config.ID == uuid.Nil {
		return r.db.Create(config).Error
	}
	return r.db.Save(config).Error
}

// CreatePaymentOrder 创建支付订单
func (r *BillingRepository) CreatePaymentOrder(order *models.PaymentOrder) error {
	return r.db.Create(order).Error
}

// GetPaymentOrderByTradeID 通过商户订单号获取订单
func (r *BillingRepository) GetPaymentOrderByTradeID(tradeOrderID string) (*models.PaymentOrder, error) {
	var order models.PaymentOrder
	err := r.db.Where("trade_order_id = ?", tradeOrderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetPaymentOrder 获取订单
func (r *BillingRepository) GetPaymentOrder(id uuid.UUID) (*models.PaymentOrder, error) {
	var order models.PaymentOrder
	err := r.db.First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdatePaymentOrder 更新订单
func (r *BillingRepository) UpdatePaymentOrder(order *models.PaymentOrder) error {
	return r.db.Save(order).Error
}

// ListPaymentOrders 列出订单
func (r *BillingRepository) ListPaymentOrders(tenantID uuid.UUID, status string, page, pageSize int) ([]models.PaymentOrder, int64, error) {
	var orders []models.PaymentOrder
	var total int64

	query := r.db.Model(&models.PaymentOrder{})
	if tenantID != uuid.Nil {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count
	countQuery := r.db.Model(&models.PaymentOrder{})
	if tenantID != uuid.Nil {
		countQuery = countQuery.Where("tenant_id = ?", tenantID)
	}
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	countQuery.Count(&total)

	// Find
	findQuery := r.db.Model(&models.PaymentOrder{})
	if tenantID != uuid.Nil {
		findQuery = findQuery.Where("tenant_id = ?", tenantID)
	}
	if status != "" {
		findQuery = findQuery.Where("status = ?", status)
	}
	err := findQuery.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}
