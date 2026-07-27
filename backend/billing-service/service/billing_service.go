package service

import (
	"errors"
	"time"

	"ai-platform/billing-service/models"
	"ai-platform/billing-service/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillingService struct {
	repo *repository.BillingRepository
}

func NewBillingService(db *gorm.DB) *BillingService {
	return &BillingService{
		repo: repository.NewBillingRepository(db),
	}
}

// GetPlans 获取所有套餐
func (s *BillingService) GetPlans() ([]models.Plan, error) {
	return s.repo.ListPlans()
}

// GetPlan 获取套餐详情
func (s *BillingService) GetPlan(id uuid.UUID) (*models.Plan, error) {
	return s.repo.GetPlan(id)
}

// Subscribe 订阅套餐
func (s *BillingService) Subscribe(tenantID, planID uuid.UUID) error {
	// 检查是否已有订阅
	if sub, err := s.repo.GetSubscription(tenantID); err == nil && sub.Status == "active" {
		return errors.New("already subscribed")
	}

	// 获取套餐信息（验证套餐存在）
	_, err := s.repo.GetPlan(planID)
	if err != nil {
		return err
	}

	plan, err := s.repo.GetPlan(planID)
	if err != nil {
		return err
	}

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

// GetSubscription 获取当前订阅
func (s *BillingService) GetSubscription(tenantID uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetSubscription(tenantID)
}

// RecordMessageUsage 记录消息使用量
func (s *BillingService) RecordMessageUsage(tenantID uuid.UUID, count int64) error {
	// 获取订阅
	sub, err := s.repo.GetSubscription(tenantID)
	if err != nil {
		return err
	}

	// 获取套餐
	plan, err := s.repo.GetPlan(sub.PlanID)
	if err != nil {
		return err
	}

	// 检查是否超出限额
	if sub.MessagesUsed+count > plan.MaxMessagesPerMonth {
		return errors.New("message limit exceeded")
	}

	// 更新使用量
	if err := s.repo.IncrementMessagesUsed(tenantID, count); err != nil {
		return err
	}

	record := &models.BillingRecord{
		TenantID:       tenantID,
		SubscriptionID: sub.ID,
		Type:           "message",
		Amount:         0,
		TokensUsed:     int(count),
		Description:    "套餐内消息使用",
		Status:         "completed",
	}

	return s.repo.CreateBillingRecord(record)
}

// GetBillingRecords 获取计费记录
func (s *BillingService) GetBillingRecords(tenantID uuid.UUID, limit int) ([]models.BillingRecord, error) {
	return s.repo.ListBillingRecords(tenantID, limit)
}

// GetMonthlyUsage 获取月度使用量
func (s *BillingService) GetMonthlyUsage(tenantID uuid.UUID, year, month int) (int64, error) {
	return s.repo.GetMonthlyUsage(tenantID, year, month)
}

// CheckQuota 检查配额
func (s *BillingService) CheckQuota(tenantID uuid.UUID) (used, limit int64, err error) {
	sub, err := s.repo.GetSubscription(tenantID)
	if err != nil {
		return 0, 0, err
	}

	plan, err := s.repo.GetPlan(sub.PlanID)
	if err != nil {
		return 0, 0, err
	}

	return sub.MessagesUsed, plan.MaxMessagesPerMonth, nil
}

// AdminListPlans 管理员获取所有套餐（含下架）
func (s *BillingService) AdminListPlans() ([]models.Plan, error) {
	return s.repo.ListAllPlans()
}

// AdminCreatePlan 管理员创建套餐
func (s *BillingService) AdminCreatePlan(plan *models.Plan) error {
	return s.repo.CreatePlan(plan)
}

// AdminUpdatePlan 管理员更新套餐
func (s *BillingService) AdminUpdatePlan(plan *models.Plan) error {
	return s.repo.UpdatePlan(plan)
}

// AdminDeletePlan 管理员删除套餐
func (s *BillingService) AdminDeletePlan(id uuid.UUID) error {
	return s.repo.DeletePlan(id)
}

// AdminTogglePlanStatus 管理员切换套餐状态
func (s *BillingService) AdminTogglePlanStatus(id uuid.UUID, status string) error {
	plan, err := s.repo.GetPlan(id)
	if err != nil {
		return err
	}
	plan.IsActive = status == "active"
	return s.repo.UpdatePlan(plan)
}

// AdminGetSubscriptions 管理员获取所有订阅
func (s *BillingService) AdminGetSubscriptions(page, pageSize int) ([]models.Subscription, int64, error) {
	return s.repo.ListSubscriptions(page, pageSize)
}

// AdminGetSubscription 管理员获取单个订阅
func (s *BillingService) AdminGetSubscription(tenantID uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetSubscription(tenantID)
}

// AdminGetBillingRecords 管理员获取所有账单记录
func (s *BillingService) AdminGetBillingRecords(page, pageSize int, status string) ([]models.BillingRecord, int64, error) {
	return s.repo.ListAllBillingRecords(page, pageSize, status)
}
