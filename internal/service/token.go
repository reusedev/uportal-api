package service

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"gorm.io/gorm"
)

// TokenService Token服务
type TokenService struct {
	db *gorm.DB
}

// NewTokenService 创建Token服务实例
func NewTokenService(db *gorm.DB) *TokenService {
	return &TokenService{db: db}
}

// UpdateConsumptionRuleRequest 更新消费规则请求
type UpdateConsumptionRuleRequest struct {
	ID          int64  `json:"id" binding:"required,min=1"`
	FeatureName string `json:"feature_name" binding:"required,max=100"`
	FeatureDesc string `json:"feature_desc" binding:"required,max=255"`
	TokenCost   *int64 `json:"token_cost" binding:"required,min=1"`
	FeatureCode string `json:"feature_code" binding:"required,max=50"`
	Status      *int8  `json:"status" binding:"required,oneof=1 2"`
}

// DeleteConsumptionRule 删除Token消费规则
func (s *TokenService) DeleteConsumptionRule(ctx context.Context, id int) error {
	// 检查规则是否存在
	_, err := model.GetTokenConsumptionRule(s.db, id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "消费规则不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询消费规则失败", err)
	}

	err = model.DeleteTokenConsumptionRule(s.db, id)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "删除消费规则失败", err)
	}

	return nil
}

// ListConsumptionRulesRequest 获取Token消费规则列表请求
type ListConsumptionRulesRequest struct {
	Offset int `json:"offset" binding:"required,min=0"`
	Limit  int `json:"limit" binding:"required,min=1"`
}

type ListUserTokenRecords struct {
	Prev  *string `json:"prev"` //上一条记录 ID
	Limit *int    `json:"limit"`
}

// CreateRechargePlanRequest 创建充值套餐请求
type CreateRechargePlanRequest struct {
	Name        string  `json:"name" binding:"required,max=50"`
	TokenAmount int64   `json:"token_amount" binding:"required,min=1"`
	Price       float64 `json:"price" binding:"required,min=0.01"`
	Discount    float64 `json:"discount" binding:"required,min=0.1,max=1"`
	Description string  `json:"description" binding:"required,max=200"`
}

// CreateRechargePlan 创建充值套餐
func (s *TokenService) CreateRechargePlan(ctx context.Context, req *CreateRechargePlanRequest) (*model.RechargePlan, error) {
	// 将 string 转换为 *string
	description := req.Description
	plan := &model.RechargePlan{
		TokenAmount: int(req.TokenAmount), // 转换为 int
		Price:       req.Price,
		Description: &description,
		Status:      1,
	}

	err := model.CreateRechargePlan(s.db, plan)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "创建充值套餐失败", err)
	}

	return plan, nil
}

// UpdateRechargePlanRequest 更新充值套餐请求
type UpdateRechargePlanRequest struct {
	Name        string  `json:"name" binding:"required,max=50"`
	TokenAmount int64   `json:"token_amount" binding:"required,min=1"`
	Price       float64 `json:"price" binding:"required,min=0.01"`
	Discount    float64 `json:"discount" binding:"required,min=0.1,max=1"`
	Description string  `json:"description" binding:"required,max=200"`
	Status      int     `json:"status" binding:"required,oneof=1 2"`
}

// UpdateRechargePlan 更新充值套餐
func (s *TokenService) UpdateRechargePlan(ctx context.Context, id int64, req *UpdateRechargePlanRequest) error {
	// 检查套餐是否存在
	_, err := model.GetRechargePlan(s.db, id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "充值套餐不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询充值套餐失败", err)
	}

	// 将 string 转换为 *string
	description := req.Description
	updates := map[string]interface{}{
		"token_amount": int(req.TokenAmount), // 转换为 int
		"price":        req.Price,
		"description":  &description,
		"status":       req.Status,
	}

	err = model.UpdateRechargePlan(s.db, id, updates)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "更新充值套餐失败", err)
	}

	return nil
}

// DeleteRechargePlan 删除充值套餐
func (s *TokenService) DeleteRechargePlan(ctx context.Context, id int64) error {
	// 检查套餐是否存在
	_, err := model.GetRechargePlan(s.db, id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "充值套餐不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询充值套餐失败", err)
	}

	err = model.DeleteRechargePlan(s.db, id)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "删除充值套餐失败", err)
	}

	return nil
}

// ListRechargePlans 获取充值套餐列表
func (s *TokenService) ListRechargePlans(ctx context.Context, page, pageSize int) ([]*model.RechargePlan, int64, error) {
	offset := (page - 1) * pageSize
	plans, total, err := model.ListRechargePlans(s.db, offset, pageSize)
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取充值套餐列表失败", err)
	}
	return plans, total, nil
}

// GetUserTokenBalance 获取用户Token余额
func (s *TokenService) GetUserTokenBalance(ctx context.Context, userID int64) (int64, error) {
	balance, err := model.GetUserTokenBalance(s.db, userID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New(errors.ErrCodeNotFound, "用户不存在", nil)
		}
		return 0, errors.New(errors.ErrCodeInternal, "获取Token余额失败", err)
	}
	return balance, nil
}

// GetUserTokenRecords 获取用户的代币记录列表
func (s *TokenService) GetUserTokenRecords(ctx context.Context, userID int64, req ListUserTokenRecords) ([]*model.TokenRecord, error) {
	var start int
	limit := 10
	if req.Prev != nil {
		start, _ = strconv.Atoi(*req.Prev)
	}
	if req.Limit != nil {
		limit = int(*req.Limit)
	}
	return model.GetTokenRecords(s.db, userID, start, limit)
}

// ConsumeToken 消费Token
func (s *TokenService) ConsumeToken(ctx context.Context, userID int64, serviceType string) error {
	// 获取消费规则
	rule, err := model.GetTokenConsumptionRuleByService(s.db, serviceType)
	if err != nil {
		return err
	}

	// 检查规则状态
	if rule.Status != 1 {
		return errors.New(errors.ErrCodeInvalidParams, "该服务已禁用", nil)
	}

	// 消费Token
	desc := ""
	if rule.FeatureDesc != nil {
		desc = *rule.FeatureDesc
	}
	return model.ConsumeToken(s.db, userID, int64(rule.TokenCost), serviceType, desc)
}

// AddToken 增加Token
func (s *TokenService) AddToken(ctx context.Context, userID int64, amount int64, recordType int, orderID string, description string) error {
	err := model.AddToken(s.db, userID, amount, recordType, orderID, description)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "增加Token失败", err)
	}
	return nil
}

// GetRechargeAmount 计算充值金额
func (s *TokenService) GetRechargeAmount(ctx context.Context, planID int64) (float64, error) {
	plan, err := model.GetRechargePlan(s.db, planID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New(errors.ErrCodeNotFound, "充值套餐不存在", nil)
		}
		return 0, errors.New(errors.ErrCodeInternal, "获取充值套餐失败", err)
	}

	if plan.Status != 1 {
		return 0, errors.New(errors.ErrCodeInvalidParams, "充值套餐未启用", nil)
	}

	return plan.Price, nil
}

// GetConsumptionAmount 获取服务消耗的Token数量
func (s *TokenService) GetConsumptionAmount(ctx context.Context, serviceType string) (int64, error) {
	rule, err := model.GetTokenConsumptionRuleByService(s.db, serviceType)
	if err != nil {
		return 0, err
	}
	return int64(rule.TokenCost), nil
}
