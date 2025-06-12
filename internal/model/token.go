package model

import (
	"gorm.io/gorm/clause"
	"strconv"
	"time"

	"github.com/reusedev/uportal-api/pkg/errors"
	"gorm.io/gorm"
)

// CreateTokenConsumptionRule 创建Token消费规则
func CreateTokenConsumptionRule(db *gorm.DB, rule *TokenConsumeRule) error {
	return db.Create(rule).Error
}

// GetTokenConsumptionRule 获取Token消费规则
func GetTokenConsumptionRule(db *gorm.DB, id int) (*TokenConsumeRule, error) {
	var rule TokenConsumeRule
	err := db.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetTokenConsumptionRuleByService 根据服务类型获取Token消费规则
func GetTokenConsumptionRuleByService(db *gorm.DB, serviceType string) (*TokenConsumeRule, error) {
	var rule TokenConsumeRule
	err := db.Where("feature_code = ? AND status = 1", serviceType).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateTokenConsumptionRule 更新Token消费规则
func UpdateTokenConsumptionRule(db *gorm.DB, id int, updates map[string]interface{}) error {
	return db.Model(&TokenConsumeRule{}).Where("feature_id = ?", id).Updates(updates).Error
}

// DeleteTokenConsumptionRule 删除Token消费规则
func DeleteTokenConsumptionRule(db *gorm.DB, id int) error {
	return db.Delete(&TokenConsumeRule{}, id).Error
}

// ListTokenConsumptionRules 获取Token消费规则列表
func ListTokenConsumptionRules(db *gorm.DB) ([]*TokenConsumeRule, int64, error) {
	var rules []*TokenConsumeRule
	var total int64

	err := db.Model(&TokenConsumeRule{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Find(&rules).Error
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// GetTokenConsumptionRules 获取Token消费规则列表
func GetTokenConsumptionRules(db *gorm.DB, class string) ([]*TokenConsumeRule, error) {
	var rules []*TokenConsumeRule

	err := db.Model(&TokenConsumeRule{}).Where("class = ?", class).Find(&rules).Error
	if err != nil {
		return nil, err
	}

	return rules, nil
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

// CreateTokenRecord 创建代币记录
func CreateTokenRecord(db *gorm.DB, record *TokenRecord) error {
	return db.Create(record).Error
}

// GetTokenRecords 获取用户的代币记录列表
func GetTokenRecords(db *gorm.DB, userID string, start, limit int) ([]*TokenRecord, error) {
	var records []*TokenRecord
	err := db.Where("user_id = ? and record_id > ?", userID, start).
		Order("change_time DESC").Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetUserTokenBalance 获取用户Token余额
func GetUserTokenBalance(db *gorm.DB, userID string) (int64, error) {
	var user User
	err := db.Select("token_balance").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return 0, err
	}
	return int64(user.TokenBalance), nil
}

// UpdateUserTokenBalance 更新用户代币余额
func UpdateUserTokenBalance(db *gorm.DB, userID string, changeAmount int) error {
	var user User
	err := db.Transaction(func(tx *gorm.DB) error {
		// 获取用户当前余额
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error
		if err != nil {
			return err
		}

		// 更新用户余额
		newBalance := user.TokenBalance + changeAmount
		if newBalance < 0 {
			return errors.New(errors.ErrCodeInsufficientBalance, "代币余额不足", nil)
		}

		return tx.Model(&User{}).Where("id = ?", userID).
			Update("token_balance", newBalance).Error
	})

	return err
}

// GetUserTokenIsBuy 获取用户Token余额
func GetUserTokenIsBuy(db *gorm.DB, userID, featureCode string, num int) (int, error) {
	var isBuy int
	err := db.Transaction(func(tx *gorm.DB) error {
		var rewardTask TokenConsumeRule
		err := tx.Model(&TokenConsumeRule{}).Where("feature_code = ?", featureCode).First(&rewardTask).Error
		if err != nil {
			return err
		}
		var user User
		err = tx.Model(&User{}).Where("id = ?", userID).First(&user).Error
		if err != nil {
			return err
		}
		if user.TokenBalance >= rewardTask.TokenCost*num {
			isBuy = 1
		}
		return nil
	})
	return isBuy, err
}

// ConsumeToken 消费Token
func ConsumeToken(db *gorm.DB, userID string, amount int64, serviceType string, description string) error {
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
		err = UpdateUserTokenBalance(tx, userID, -int(amount))
		if err != nil {
			return err
		}

		// 创建Token记录
		record := &TokenRecord{
			UserID:       userID,
			ChangeAmount: -int(amount),
			BalanceAfter: int(balance - amount),
			ChangeType:   "CONSUME",
			Remark:       &description,
			ChangeTime:   time.Now(),
		}
		return CreateTokenRecord(tx, record)
	})
}

// AddToken 增加Token
func AddToken(db *gorm.DB, userID string, amount int64, recordType int, orderID string, description string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 获取用户当前余额
		balance, err := GetUserTokenBalance(tx, userID)
		if err != nil {
			return err
		}

		// 更新用户余额
		err = UpdateUserTokenBalance(tx, userID, int(amount))
		if err != nil {
			return err
		}

		// 创建Token记录
		record := &TokenRecord{
			UserID:       userID,
			ChangeAmount: int(amount),
			BalanceAfter: int(balance + amount),
			ChangeType:   getChangeType(recordType),
			Remark:       &description,
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

// CreateTokenConsumptionRecord 创建代币消费记录
func CreateTokenConsumptionRecord(db *gorm.DB, userID string, featureID int, amount int, remark string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 获取用户当前余额
		var user User
		err := tx.First(&user, userID).Error
		if err != nil {
			return err
		}

		// 检查余额是否足够
		if user.TokenBalance < amount {
			return errors.New(errors.ErrCodeInsufficientBalance, "代币余额不足", nil)
		}

		// 更新用户余额
		newBalance := user.TokenBalance - amount
		err = tx.Model(&User{}).Where("id = ?", userID).
			Update("token_balance", newBalance).Error
		if err != nil {
			return err
		}

		// 创建消费记录
		record := &TokenRecord{
			UserID:       userID,
			ChangeAmount: -amount,
			BalanceAfter: newBalance,
			ChangeType:   "consume",
			FeatureID:    &featureID,
			Remark:       &remark,
		}

		return tx.Create(record).Error
	})
}

// CreateTokenRewardRecord 创建代币奖励记录
func CreateTokenRewardRecord(db *gorm.DB, userID string, taskID int, amount int, remark string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 获取用户当前余额
		var user User
		err := tx.First(&user, userID).Error
		if err != nil {
			return err
		}

		// 更新用户余额
		newBalance := user.TokenBalance + amount
		err = tx.Model(&User{}).Where("id = ?", userID).
			Update("token_balance", newBalance).Error
		if err != nil {
			return err
		}

		// 创建奖励记录
		record := &TokenRecord{
			UserID:       userID,
			ChangeAmount: amount,
			BalanceAfter: newBalance,
			ChangeType:   "reward",
			TaskID:       &taskID,
			Remark:       &remark,
		}

		return tx.Create(record).Error
	})
}
