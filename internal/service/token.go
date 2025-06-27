package service

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/reusedev/uportal-api/pkg/logs"
	"go.uber.org/zap"

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
	Class       string `json:"class" binding:"required,max=100"`
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
	Prev  *int `json:"prev"` //上一条记录 ID
	Limit *int `json:"limit"`
}

type ListUserTokenResp struct {
	Id           int64     `json:"id"`
	Source       string    `json:"source"`
	Points       int       `json:"points"`
	CreatedAt    time.Time `json:"created_at"`
	Remark       string    `json:"remark"`
	BalanceAfter int       `json:"balance_after"`
}

// CreateRechargePlanRequest 创建充值套餐请求
type CreateRechargePlanRequest struct {
	TokenAmount int     `json:"token_amount" binding:"required,min=1"`
	Price       float64 `json:"price" binding:"required"`
	Currency    string  `json:"currency" binding:"required"`
	Description string  `json:"description" binding:"required,max=200"`
	Status      *int8   `json:"status" binding:"required,oneof=1 2"` // 1:启用, 2:禁用
	Name        string  `json:"name" binding:"required,max=100"`
	Tag         string  `json:"tag" binding:"required,max=100"`
	IsRecommend *int8   `json:"is_recommend" binding:"required"` // 是否推荐 0:推荐, 1:不推荐
}

type ListRechargePlanRequest struct {
	Status *int8 `json:"status"`
}

type TokenIsBuyRequest struct {
	UserId      string `json:"user_id" binding:"required"`
	FeatureCode string `json:"feature_code" binding:"required"`
	Num         int    `json:"num" binding:"required,min=1"`
}

type TokenBuyRequest struct {
	UserId      string `json:"user_id" binding:"required"`
	FeatureCode string `json:"feature_code" binding:"required"`
	Num         int    `json:"num" binding:"required,min=1"`
	Type        int    `json:"type" binding:"required,oneof=1 2"`
}

// CreateRechargePlan 创建充值套餐
func (s *TokenService) CreateRechargePlan(ctx context.Context, req *CreateRechargePlanRequest) (*model.RechargePlan, error) {
	// 将 string 转换为 *string
	description := req.Description
	plan := &model.RechargePlan{
		TokenAmount: req.TokenAmount, // 转换为 int
		Price:       float64(req.Price),
		Description: &description,
		Status:      *req.Status,
		Currency:    req.Currency,
		Name:        req.Name,
		Tag:         req.Tag,
		IsRecommend: *req.IsRecommend,
	}

	err := model.CreateRechargePlan(s.db, plan)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "创建充值套餐失败", err)
	}

	return plan, nil
}

// UpdateRechargePlanRequest 更新充值套餐请求
type UpdateRechargePlanRequest struct {
	ID          int64    `json:"plan_id" binding:"required,min=1"`
	Currency    string   `json:"currency" binding:"required"`
	TokenAmount int      `json:"token_amount" binding:"required"`
	Price       *float64 `json:"price"`
	Description *string  `json:"description"`
	Status      *int     `json:"status"`
	Name        string   `json:"name" binding:"required,max=100"`
	Tag         string   `json:"tag" binding:"required,max=100"`
	IsRecommend *int8    `json:"is_recommend" binding:"required"` // 是否推荐 0:推荐, 1:不推荐
}

// UpdateRechargePlan 更新充值套餐
func (s *TokenService) UpdateRechargePlan(ctx context.Context, req *UpdateRechargePlanRequest) error {
	// 检查套餐是否存在
	_, err := model.GetRechargePlan(s.db, req.ID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "充值套餐不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询充值套餐失败", err)
	}

	updates := map[string]interface{}{
		"token_amount": req.TokenAmount,
		"currency":     req.Currency,
		"name":         req.Name,
		"tag":          req.Tag,
		"is_recommend": *req.IsRecommend,
	}
	if req.Price != nil {
		updates["price"] = float64(*req.Price)
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	err = model.UpdateRechargePlan(s.db, req.ID, updates)
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
func (s *TokenService) ListRechargePlans(ctx context.Context, status *int8) ([]*model.RechargePlan, error) {
	plans, err := model.ListRechargePlans(s.db, status)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "获取充值套餐列表失败", err)
	}
	return plans, nil
}

// GetUserTokenBalance 获取用户Token余额
func (s *TokenService) GetUserTokenBalance(ctx context.Context, userID string) (int64, error) {
	balance, err := model.GetUserTokenBalance(s.db, userID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New(errors.ErrCodeNotFound, "用户不存在", nil)
		}
		return 0, errors.New(errors.ErrCodeInternal, "获取Token余额失败", err)
	}
	return balance, nil
}

// TokenIsBuy 获取用户Token余额
func (s *TokenService) TokenIsBuy(ctx context.Context, userID, FeatureCode string, num int) (int, error) {

	isBuy, err := model.GetUserTokenIsBuy(s.db, userID, FeatureCode, num)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New(errors.ErrCodeNotFound, "用户不存在", nil)
		}
		return 0, errors.New(errors.ErrCodeInternal, "获取Token余额失败", err)
	}
	return isBuy, nil
}

// TokenBuy 用户金币消耗
func (s *TokenService) TokenBuy(ctx context.Context, userID, FeatureCode string, num int) (int, error) {

	isBuy, err := model.GetUserTokenIsBuy(s.db, userID, FeatureCode, num)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New(errors.ErrCodeNotFound, "用户不存在", nil)
		}
		return 0, errors.New(errors.ErrCodeInternal, "获取Token余额失败", err)
	}
	return isBuy, nil
}

// GetUserTokenRecords 获取用户的代币记录列表
func (s *TokenService) GetUserTokenRecords(ctx context.Context, userID string, req ListUserTokenRecords) ([]*model.TokenRecord, error) {
	var start int
	limit := 10
	if req.Prev != nil {
		start = *req.Prev
	}
	if req.Limit != nil {
		limit = *req.Limit
	}
	return model.GetTokenRecords(s.db, userID, start, limit)
}

// ConsumeToken 消费Token
func (s *TokenService) ConsumeToken(ctx context.Context, userID, featureCode, descSuffix string, num int) (int64, error) {
	// 获取消费规则
	rule, err := model.GetTokenConsumptionRuleByService(s.db, featureCode)
	if err != nil {
		return 0, err
	}

	// 检查规则状态
	if rule.Status != 1 {
		return 0, errors.New(errors.ErrCodeInvalidParams, "该服务已禁用", nil)
	}

	// 消费Token
	desc := descSuffix
	if rule.FeatureDesc != nil {
		desc = *rule.FeatureDesc + descSuffix
	}
	cost := int64(rule.TokenCost * num)
	err = model.ConsumeToken(s.db, userID, cost, featureCode, desc)
	return cost, err
}

// AddToken 增加Token
func (s *TokenService) AddToken(ctx context.Context, userID string, amount int64, recordType int, orderID string, description string) error {
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

// ProcessPointsReward 处理代币奖励
func (s *TokenService) ProcessPointsReward(ctx context.Context, userID string, rewardType string) error {
	// 验证奖励类型
	amount, exists := rewardAmounts[rewardType]
	if !exists {
		return errors.New(errors.ErrCodeInvalidParams, "无效的奖励类型", nil)
	}

	// 开启事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return errors.New(errors.ErrCodeInternal, "开启事务失败", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查用户是否存在
	var user model.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeUserNotFound, "用户不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "获取用户信息失败", err)
	}

	// 检查用户状态
	if user.Status != 1 {
		tx.Rollback()
		return errors.New(errors.ErrCodeUserDisabled, "用户账号已被禁用", nil)
	}

	// 更新用户代币余额
	if err := tx.Model(&user).Update("token_balance", gorm.Expr("token_balance + ?", amount)).Error; err != nil {
		tx.Rollback()
		return errors.New(errors.ErrCodeInternal, "更新代币余额失败", err)
	}

	// 创建代币变动记录
	tokenRecord := &model.TokenRecord{
		UserID:       userID,
		ChangeAmount: amount,
		ChangeType:   rewardType,
		BalanceAfter: user.TokenBalance + amount,
		Remark:       model.StringPtr(getRewardRemark(rewardType)),
	}

	if err := tx.Create(tokenRecord).Error; err != nil {
		tx.Rollback()
		return errors.New(errors.ErrCodeInternal, "创建代币记录失败", err)
	}

	// 记录日志
	logs.Business().Info("代币奖励发放成功",
		zap.String("user_id", userID),
		zap.String("reward_type", rewardType),
		zap.Int("amount", amount),
	)

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "提交事务失败", err)
	}

	return nil
}

// getRewardRemark 获取奖励类型的备注说明
func getRewardRemark(rewardType string) string {
	switch rewardType {
	case RewardTypeDailyLogin:
		return "每日登录奖励"
	case RewardTypeProfile:
		return "完善资料奖励"
	case RewardTypeShare:
		return "分享奖励"
	case RewardTypeFeedback:
		return "反馈奖励"
	case RewardAd:
		return "观看广告奖励"
	case Recharge:
		return "充值"
	default:
		return "其他奖励"
	}
}
