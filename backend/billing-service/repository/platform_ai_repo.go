package repository

import (
	"errors"
	"time"

	"ai-platform/billing-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlatformAIRepository struct {
	db *gorm.DB
}

func NewPlatformAIRepository(db *gorm.DB) *PlatformAIRepository {
	return &PlatformAIRepository{db: db}
}

// GetTenantBalance 获取租户余额
func (r *PlatformAIRepository) GetTenantBalance(tenantID uuid.UUID) (*models.TenantBalance, error) {
	var balance models.TenantBalance
	err := r.db.Where("tenant_id = ?", tenantID).First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.TenantBalance{TenantID: tenantID, Balance: 0}, nil
		}
		return nil, err
	}
	return &balance, nil
}

// RechargeBalance 充值余额（原子操作）
func (r *PlatformAIRepository) RechargeBalance(tenantID uuid.UUID, amount float64) error {
	result := r.db.Model(&models.TenantBalance{}).
		Where("tenant_id = ?", tenantID).
		Updates(map[string]interface{}{
			"balance":         gorm.Expr("balance + ?", amount),
			"total_recharged": gorm.Expr("total_recharged + ?", amount),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		newBalance := &models.TenantBalance{
			TenantID:       tenantID,
			Balance:        amount,
			TotalRecharged: amount,
		}
		return r.db.Create(newBalance).Error
	}
	return nil
}

// DeductBalance 扣减余额（原子操作，余额不足时返回错误）
func (r *PlatformAIRepository) DeductBalance(tenantID uuid.UUID, cost float64) error {
	result := r.db.Model(&models.TenantBalance{}).
		Where("tenant_id = ? AND balance >= ?", tenantID, cost).
		Updates(map[string]interface{}{
			"balance":        gorm.Expr("balance - ?", cost),
			"total_consumed": gorm.Expr("total_consumed + ?", cost),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("insufficient balance")
	}
	return nil
}

// GetResourcePackages 获取可用资源包列表
func (r *PlatformAIRepository) GetResourcePackages() ([]models.ResourcePackage, error) {
	var packages []models.ResourcePackage
	err := r.db.Where("is_active = ?", true).Order("sort_order ASC").Find(&packages).Error
	return packages, err
}

// GetResourcePackage 获取单个资源包
func (r *PlatformAIRepository) GetResourcePackage(id uuid.UUID) (*models.ResourcePackage, error) {
	var pkg models.ResourcePackage
	err := r.db.First(&pkg, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// PurchasePackage 购买资源包（事务：扣余额 + 创建包记录 + 创建消费记录）
func (r *PlatformAIRepository) PurchasePackage(tenantID uuid.UUID, pkg *models.ResourcePackage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 原子扣减余额
		result := tx.Model(&models.TenantBalance{}).
			Where("tenant_id = ? AND balance >= ?", tenantID, pkg.Price).
			Updates(map[string]interface{}{
				"balance":        gorm.Expr("balance - ?", pkg.Price),
				"total_consumed": gorm.Expr("total_consumed + ?", pkg.Price),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("insufficient balance")
		}

		// 创建租户资源包记录
		tenantPkg := &models.TenantResourcePackage{
			TenantID:       tenantID,
			PackageID:      pkg.ID,
			TokenName:      pkg.Name,
			TokenAmount:    pkg.TokenAmount,
			TokenUsed:      0,
			TokenRemaining: pkg.TokenAmount,
			Price:          pkg.Price,
			Status:         "active",
		}
		if err := tx.Create(tenantPkg).Error; err != nil {
			return err
		}

		// 创建消费记录
		record := &models.TenantRechargeRecord{
			TenantID:      tenantID,
			Amount:        -pkg.Price,
			PaymentMethod: "balance",
			Status:        "completed",
			Remark:        "购买资源包: " + pkg.Name,
		}
		return tx.Create(record).Error
	})
}

// GetTenantPackages 获取租户有效资源包
func (r *PlatformAIRepository) GetTenantPackages(tenantID uuid.UUID) ([]models.TenantResourcePackage, error) {
	var packages []models.TenantResourcePackage
	err := r.db.Where("tenant_id = ? AND status = ?", tenantID, "active").
		Order("created_at ASC").Find(&packages).Error
	return packages, err
}

// UsePackageTokens 从最早的有效资源包中扣减 token（FIFO，原子操作）
func (r *PlatformAIRepository) UsePackageTokens(tenantID uuid.UUID, tokens int64) (*uuid.UUID, int64, error) {
	remaining := tokens
	var usedPackageID *uuid.UUID

	for remaining > 0 {
		var pkg models.TenantResourcePackage
		err := r.db.Where("tenant_id = ? AND status = ? AND token_remaining > 0", tenantID, "active").
			Order("created_at ASC").First(&pkg).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return nil, 0, err
		}

		useFromThis := remaining
		if useFromThis > pkg.TokenRemaining {
			useFromThis = pkg.TokenRemaining
		}

		result := r.db.Model(&models.TenantResourcePackage{}).
			Where("id = ? AND token_remaining >= ?", pkg.ID, useFromThis).
			Updates(map[string]interface{}{
				"token_used":      gorm.Expr("token_used + ?", useFromThis),
				"token_remaining": gorm.Expr("token_remaining - ?", useFromThis),
			})
		if result.Error != nil {
			return nil, 0, result.Error
		}
		if result.RowsAffected == 0 {
			continue
		}

		remaining -= useFromThis
		usedPackageID = &pkg.ID

		// 如果包用完了，标记为 exhausted
		if pkg.TokenRemaining-useFromThis == 0 {
			r.db.Model(&models.TenantResourcePackage{}).
				Where("id = ?", pkg.ID).
				Update("status", "exhausted")
		}
	}

	usedFromPackage := tokens - remaining
	return usedPackageID, usedFromPackage, nil
}

// GetUsageDetails 分页获取用量明细
func (r *PlatformAIRepository) GetUsageDetails(tenantID uuid.UUID, page, pageSize int, startDate, endDate *time.Time) ([]models.TenantUsageDetail, int64, error) {
	var details []models.TenantUsageDetail
	var total int64

	query := r.db.Model(&models.TenantUsageDetail{}).Where("tenant_id = ?", tenantID)
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	query.Count(&total)

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&details).Error
	return details, total, err
}

// AIUsageSummary 月度 AI 用量汇总
type AIUsageSummary struct {
	Month       string  `json:"month"`
	TotalTokens int64   `json:"total_tokens"`
	TotalCost   float64 `json:"total_cost"`
	RecordCount int64   `json:"record_count"`
}

// GetMonthlyUsageSummary 获取月度 AI 用量汇总
func (r *PlatformAIRepository) GetMonthlyUsageSummary(tenantID uuid.UUID, months int) ([]AIUsageSummary, error) {
	since := time.Now().AddDate(0, -months, 0)
	var results []AIUsageSummary
	err := r.db.Model(&models.TenantUsageDetail{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') as month, COALESCE(SUM(total_tokens), 0) as total_tokens, COALESCE(SUM(cost), 0) as total_cost, COUNT(*) as record_count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Group("TO_CHAR(created_at, 'YYYY-MM')").
		Order("month ASC").
		Scan(&results).Error
	return results, err
}

// GetPlatformPricing 获取平台模型定价列表
// JOIN ai_models 取得真实 model_name 与价格（platform_model_pricing 的价格为 0 且 model_name 留空），
// 将 ai_models 中的 per-1k 价格 × 1000 转换为每百万 Token，匹配前端展示约定。
func (r *PlatformAIRepository) GetPlatformPricing() ([]models.ModelPricingView, error) {
	var rows []models.ModelPricingView
	err := r.db.Table("platform_model_pricing p").
		Select("p.id as id, COALESCE(m.model_name, '') as model_name, COALESCE(m.input_price_per1k, 0) * 1000 as input_price, COALESCE(m.output_price_per1k, 0) * 1000 as output_price").
		Joins("LEFT JOIN ai_models m ON m.id = p.model_id").
		Where("p.is_enabled = ?", true).
		Order("m.provider_id, m.model_name").
		Scan(&rows).Error
	return rows, err
}

// GetModelPricing 获取单个模型定价
func (r *PlatformAIRepository) GetModelPricing(modelID string) (*models.PlatformModelPricing, error) {
	var pricing models.PlatformModelPricing
	err := r.db.Where("model_id = ? AND is_enabled = ?", modelID, true).First(&pricing).Error
	if err != nil {
		return nil, err
	}
	return &pricing, nil
}

// CreateUsageDetail 创建用量明细
func (r *PlatformAIRepository) CreateUsageDetail(detail *models.TenantUsageDetail) error {
	return r.db.Create(detail).Error
}

// CreateRechargeRecord 创建充值记录
func (r *PlatformAIRepository) CreateRechargeRecord(record *models.TenantRechargeRecord) error {
	return r.db.Create(record).Error
}
