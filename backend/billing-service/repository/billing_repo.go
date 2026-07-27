package repository

import (
	"ai-platform/billing-service/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillingRepository struct {
	db *gorm.DB
}

func NewBillingRepository(db *gorm.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

// GetPlan 获取套餐
func (r *BillingRepository) GetPlan(id uuid.UUID) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.First(&plan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// ListPlans 列出所有套餐
func (r *BillingRepository) ListPlans() ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.Where("is_active = ?", true).Order("sort_order ASC").Find(&plans).Error
	return plans, err
}

// ListAllPlans 列出所有套餐（含下架）
func (r *BillingRepository) ListAllPlans() ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.Order("sort_order ASC").Find(&plans).Error
	return plans, err
}

// CreatePlan 创建套餐
func (r *BillingRepository) CreatePlan(plan *models.Plan) error {
	return r.db.Create(plan).Error
}

// UpdatePlan 更新套餐
func (r *BillingRepository) UpdatePlan(plan *models.Plan) error {
	return r.db.Save(plan).Error
}

// DeletePlan 删除套餐
func (r *BillingRepository) DeletePlan(id uuid.UUID) error {
	return r.db.Delete(&models.Plan{}, "id = ?", id).Error
}

// GetSubscription 获取订阅
func (r *BillingRepository) GetSubscription(tenantID uuid.UUID) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Where("tenant_id = ?", tenantID).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// CreateSubscription 创建订阅
func (r *BillingRepository) CreateSubscription(subscription *models.Subscription) error {
	return r.db.Create(subscription).Error
}

// UpdateSubscription 更新订阅
func (r *BillingRepository) UpdateSubscription(subscription *models.Subscription) error {
	return r.db.Save(subscription).Error
}

// IncrementMessagesUsed 增加已用消息数
func (r *BillingRepository) IncrementMessagesUsed(tenantID uuid.UUID, count int64) error {
	return r.db.Model(&models.Subscription{}).Where("tenant_id = ?", tenantID).
		UpdateColumn("messages_used", gorm.Expr("messages_used + ?", count)).Error
}

// CreateBillingRecord 创建计费记录
func (r *BillingRepository) CreateBillingRecord(record *models.BillingRecord) error {
	return r.db.Create(record).Error
}

// ListBillingRecords 列出计费记录
func (r *BillingRepository) ListBillingRecords(tenantID uuid.UUID, limit int) ([]models.BillingRecord, error) {
	var records []models.BillingRecord
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// GetMonthlyUsage 获取月度使用量
func (r *BillingRepository) GetMonthlyUsage(tenantID uuid.UUID, year, month int) (int64, error) {
	var total int64
	err := r.db.Model(&models.BillingRecord{}).
		Where("tenant_id = ? AND EXTRACT(YEAR FROM record_date) = ? AND EXTRACT(MONTH FROM record_date) = ?",
			tenantID, year, month).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// ListSubscriptions 列出所有订阅（分页）
func (r *BillingRepository) ListSubscriptions(page, pageSize int) ([]models.Subscription, int64, error) {
	var subscriptions []models.Subscription
	var total int64
	r.db.Model(&models.Subscription{}).Count(&total)
	err := r.db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&subscriptions).Error
	return subscriptions, total, err
}

// ListAllBillingRecords 列出所有账单记录（分页、带状态筛选）
func (r *BillingRepository) ListAllBillingRecords(page, pageSize int, status string) ([]models.BillingRecord, int64, error) {
	var records []models.BillingRecord
	var total int64
	// Count
	countQuery := r.db.Model(&models.BillingRecord{})
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	countQuery.Count(&total)

	// Find
	findQuery := r.db.Model(&models.BillingRecord{})
	if status != "" {
		findQuery = findQuery.Where("status = ?", status)
	}
	err := findQuery.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&records).Error
	return records, total, err
}

// ListBillingRecordsForExport returns billing records for CSV export
func (r *BillingRepository) ListBillingRecordsForExport(tenantID uuid.UUID, startDate, endDate time.Time) ([]models.BillingRecord, error) {
	var records []models.BillingRecord
	query := r.db.Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	err := query.Order("created_at ASC").Find(&records).Error
	return records, err
}

// MonthlyUsageSummary represents monthly usage aggregation
type MonthlyUsageSummary struct {
	Month       string
	TotalTokens int64
	TotalAmount float64
	RecordCount int64
}

// GetMonthlyUsageSummary returns aggregated monthly usage for the past N months
func (r *BillingRepository) GetMonthlyUsageSummary(tenantID uuid.UUID, months int) ([]MonthlyUsageSummary, error) {
	since := time.Now().AddDate(0, -months, 0)
	var results []MonthlyUsageSummary
	err := r.db.Model(&models.BillingRecord{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') as month, COALESCE(SUM(tokens_used), 0) as total_tokens, COALESCE(SUM(amount), 0) as total_amount, COUNT(*) as record_count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Group("TO_CHAR(created_at, 'YYYY-MM')").
		Order("month ASC").
		Scan(&results).Error
	return results, err
}
