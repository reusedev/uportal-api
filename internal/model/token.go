package model

import (
	"strconv"
	"time"

	"github.com/reusedev/uportal-api/types"
	"gorm.io/gorm"
)

// TokenConsumptionRule Token消费规则
type TokenConsumptionRule struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	ServiceType string    `gorm:"not null;uniqueIndex;comment:服务类型" json:"service_type"`
	TokenAmount int64     `gorm:"not null;comment:消耗Token数量" json:"token_amount"`
	Description string    `gorm:"size:200;comment:描述" json:"description"`
	Status      int       `gorm:"not null;default:1;comment:状态 1-启用 2-禁用" json:"status"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}

// TableName 指定表名
func (TokenConsumptionRule) TableName() string {
	return "token_consume_rules"
}

// RechargePlan 充值套餐
type RechargePlan struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;not null;comment:套餐名称" json:"name"`
	TokenAmount int64     `gorm:"not null;comment:Token数量" json:"token_amount"`
	Price       float64   `gorm:"not null;comment:价格" json:"price"`
	Discount    float64   `gorm:"not null;default:1.0;comment:折扣" json:"discount"`
	Description string    `gorm:"size:200;comment:套餐描述" json:"description"`
	Status      int       `gorm:"not null;default:1;comment:状态 1-启用 2-禁用" json:"status"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}

// CreateTokenConsumptionRule 创建Token消费规则
func CreateTokenConsumptionRule(db *gorm.DB, rule *TokenConsumptionRule) error {
	return db.Create(rule).Error
}

// GetTokenConsumptionRule 获取Token消费规则
func GetTokenConsumptionRule(db *gorm.DB, id int64) (*TokenConsumptionRule, error) {
	var rule TokenConsumptionRule
	err := db.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetTokenConsumptionRuleByService 根据服务类型获取Token消费规则
func GetTokenConsumptionRuleByService(db *gorm.DB, serviceType string) (*TokenConsumptionRule, error) {
	var rule TokenConsumptionRule
	err := db.Where("service_type = ? AND status = 1", serviceType).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateTokenConsumptionRule 更新Token消费规则
func UpdateTokenConsumptionRule(db *gorm.DB, id int64, updates map[string]interface{}) error {
	return db.Model(&TokenConsumptionRule{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteTokenConsumptionRule 删除Token消费规则
func DeleteTokenConsumptionRule(db *gorm.DB, id int64) error {
	return db.Delete(&TokenConsumptionRule{}, id).Error
}

// ListTokenConsumptionRules 获取Token消费规则列表
func ListTokenConsumptionRules(db *gorm.DB, offset, limit int) ([]*TokenConsumptionRule, int64, error) {
	var rules []*TokenConsumptionRule
	var total int64

	err := db.Model(&TokenConsumptionRule{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Offset(offset).Limit(limit).Find(&rules).Error
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// CreateRechargePlan 创建充值套餐
func CreateRechargePlan(db *gorm.DB, plan *RechargePlan) error {
	return db.Create(plan).Error
}

// GetRechargePlan 获取充值套餐
func GetRechargePlan(db *gorm.DB, id int64) (*RechargePlan, error) {
	var plan RechargePlan
	err := db.First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// UpdateRechargePlan 更新充值套餐
func UpdateRechargePlan(db *gorm.DB, id int64, updates map[string]interface{}) error {
	return db.Model(&RechargePlan{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteRechargePlan 删除充值套餐
func DeleteRechargePlan(db *gorm.DB, id int64) error {
	return db.Delete(&RechargePlan{}, id).Error
}

// ListRechargePlans 获取充值套餐列表
func ListRechargePlans(db *gorm.DB, offset, limit int) ([]*RechargePlan, int64, error) {
	var plans []*RechargePlan
	var total int64

	err := db.Model(&RechargePlan{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Where("status = 1").Offset(offset).Limit(limit).Find(&plans).Error
	if err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

// CreateTokenRecord 创建Token记录
func CreateTokenRecord(db *gorm.DB, record *types.TokenRecord) error {
	return db.Create(record).Error
}

// GetTokenRecords 获取用户的Token记录
func GetTokenRecords(db *gorm.DB, userID int64, offset, limit int) ([]*types.TokenRecord, int64, error) {
	var records []*types.TokenRecord
	var total int64

	err := db.Model(&types.TokenRecord{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetUserTokenBalance 获取用户Token余额
func GetUserTokenBalance(db *gorm.DB, userID int64) (int64, error) {
	var user types.User
	err := db.Select("token_balance").First(&user, userID).Error
	if err != nil {
		return 0, err
	}
	return int64(user.TokenBalance), nil
}

// UpdateUserTokenBalance 更新用户Token余额
func UpdateUserTokenBalance(db *gorm.DB, userID int64, amount int64) error {
	return db.Model(&types.User{}).Where("user_id = ?", userID).
		Update("token_balance", gorm.Expr("token_balance + ?", amount)).Error
}

// ConsumeToken 消费Token
func ConsumeToken(db *gorm.DB, userID int64, amount int64, serviceType string, description string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 获取用户当前余额
		balance, err := GetUserTokenBalance(tx, userID)
		if err != nil {
			return err
		}

		// 检查余额是否足够
		if balance < amount {
			return gorm.ErrRecordNotFound // TODO: 使用自定义错误
		}

		// 更新用户余额
		err = UpdateUserTokenBalance(tx, userID, -amount)
		if err != nil {
			return err
		}

		// 创建Token记录
		record := &types.TokenRecord{
			UserID:       userID,
			ChangeAmount: -int(amount),
			BalanceAfter: int(balance - amount),
			ChangeType:   "CONSUME",
			Remark:       description,
			ChangeTime:   time.Now(),
		}
		return CreateTokenRecord(tx, record)
	})
}

// AddToken 增加Token
func AddToken(db *gorm.DB, userID int64, amount int64, recordType int, orderID string, description string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 获取用户当前余额
		balance, err := GetUserTokenBalance(tx, userID)
		if err != nil {
			return err
		}

		// 更新用户余额
		err = UpdateUserTokenBalance(tx, userID, amount)
		if err != nil {
			return err
		}

		// 创建Token记录
		record := &types.TokenRecord{
			UserID:       userID,
			ChangeAmount: int(amount),
			BalanceAfter: int(balance + amount),
			ChangeType:   getChangeType(recordType),
			Remark:       description,
			ChangeTime:   time.Now(),
		}
		if orderID != "" {
			orderIDInt, _ := strconv.ParseInt(orderID, 10, 64)
			record.OrderID = &orderIDInt
		}
		return CreateTokenRecord(tx, record)
	})
}

// getChangeType 获取变动类型
func getChangeType(recordType int) string {
	switch recordType {
	case 1:
		return "RECHARGE"
	case 2:
		return "CONSUME"
	case 3:
		return "REWARD"
	case 4:
		return "REFUND"
	default:
		return "OTHER"
	}
}
