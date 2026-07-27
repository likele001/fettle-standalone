package service

import (
	"fmt"
	"time"

	"ai-platform/billing-service/models"
	"ai-platform/billing-service/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIBillingService struct {
	repo *repository.PlatformAIRepository
}

func NewAIBillingService(db *gorm.DB) *AIBillingService {
	return &AIBillingService{
		repo: repository.NewPlatformAIRepository(db),
	}
}

// DeductAICost 扣减 AI 调用费用
// 1. 从 platform_model_pricing 获取定价
// 2. 计算费用
// 3. 优先使用资源包 token（FIFO），不足部分从余额扣减
// 4. 创建用量明细记录
func (s *AIBillingService) DeductAICost(tenantID uuid.UUID, modelID string, inputTokens, outputTokens int64) error {
	// 获取模型定价
	pricing, err := s.repo.GetModelPricing(modelID)
	if err != nil {
		return fmt.Errorf("model pricing not found for %s: %w", modelID, err)
	}

	// 计算费用
	totalTokens := inputTokens + outputTokens
	inputCost := float64(inputTokens) / 1000.0 * pricing.InputPricePer1K
	outputCost := float64(outputTokens) / 1000.0 * pricing.OutputPricePer1K
	totalCost := inputCost + outputCost

	if totalCost <= 0 && totalTokens <= 0 {
		return nil
	}

	// 优先使用资源包
	packageID, usedFromPackage, err := s.repo.UsePackageTokens(tenantID, totalTokens)
	if err != nil {
		return fmt.Errorf("use package tokens failed: %w", err)
	}

	billingMode := "balance"
	if usedFromPackage > 0 && packageID != nil {
		billingMode = "package"
	}

	// 如果资源包不够覆盖全部 token，剩余部分从余额扣减
	remainingTokens := totalTokens - usedFromPackage
	if remainingTokens > 0 && totalCost > 0 {
		// 按剩余 token 比例计算费用
		remainingCost := totalCost * float64(remainingTokens) / float64(totalTokens)
		if err := s.repo.DeductBalance(tenantID, remainingCost); err != nil {
			return fmt.Errorf("deduct balance failed: %w", err)
		}
		billingMode = "mixed"
	}

	// 创建用量明细
	detail := &models.TenantUsageDetail{
		TenantID:     tenantID,
		ModelID:      uuid.MustParse(modelID),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		Cost:         totalCost,
		BillingMode:  billingMode,
		PackageID:    packageID,
	}
	return s.repo.CreateUsageDetail(detail)
}

// GetBalance 获取余额（含资源包信息）
func (s *AIBillingService) GetBalance(tenantID uuid.UUID) (*models.TenantBalance, []models.TenantResourcePackage, error) {
	balance, err := s.repo.GetTenantBalance(tenantID)
	if err != nil {
		return nil, nil, err
	}
	packages, err := s.repo.GetTenantPackages(tenantID)
	if err != nil {
		return nil, nil, err
	}
	return balance, packages, nil
}

// Recharge 管理员充值
func (s *AIBillingService) Recharge(tenantID uuid.UUID, amount float64, paymentMethod, remark, operatorID string) error {
	if err := s.repo.RechargeBalance(tenantID, amount); err != nil {
		return err
	}
	record := &models.TenantRechargeRecord{
		TenantID:      tenantID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Status:        "completed",
		Remark:        remark,
		OperatedBy:    operatorID,
	}
	return s.repo.CreateRechargeRecord(record)
}

// ListPackages 获取可用资源包列表
func (s *AIBillingService) ListPackages() ([]models.ResourcePackage, error) {
	return s.repo.GetResourcePackages()
}

// PurchasePackage 购买资源包
func (s *AIBillingService) PurchasePackage(tenantID, packageID uuid.UUID) error {
	pkg, err := s.repo.GetResourcePackage(packageID)
	if err != nil {
		return err
	}
	if !pkg.IsActive {
		return fmt.Errorf("package is not active")
	}
	return s.repo.PurchasePackage(tenantID, pkg)
}

// GetUsageDetails 获取用量明细
func (s *AIBillingService) GetUsageDetails(tenantID uuid.UUID, page, pageSize int, startDate, endDate *string) ([]models.TenantUsageDetail, int64, error) {
	var sd, ed *time.Time
	if startDate != nil && *startDate != "" {
		t, err := time.Parse("2006-01-02", *startDate)
		if err == nil {
			sd = &t
		}
	}
	if endDate != nil && *endDate != "" {
		t, err := time.Parse("2006-01-02", *endDate)
		if err == nil {
			ed = &t
		}
	}
	return s.repo.GetUsageDetails(tenantID, page, pageSize, sd, ed)
}

// GetPlatformPricing 获取平台模型定价
func (s *AIBillingService) GetPlatformPricing() ([]models.PlatformModelPricing, error) {
	return s.repo.GetPlatformPricing()
}

// GetMonthlyUsageSummary 获取月度用量汇总
func (s *AIBillingService) GetMonthlyUsageSummary(tenantID uuid.UUID, months int) ([]repository.AIUsageSummary, error) {
	return s.repo.GetMonthlyUsageSummary(tenantID, months)
}
