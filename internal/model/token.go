package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetUserBalance 获取用户积分余额
func GetUserBalance(db *gorm.DB, userID int64) (int, error) {
	var user User
	err := db.Select("token_balance").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return 0, err
	}
	return user.TokenBalance, nil
}

// GetTokenRecords 获取用户积分变动记录（cursor 分页）
func GetTokenRecords(db *gorm.DB, userID int64, prev int64, limit int) ([]*TokenRecord, error) {
	var records []*TokenRecord
	tx := db.Where("user_id = ?", userID)
	if prev > 0 {
		tx = tx.Where("record_id < ?", prev)
	}
	err := tx.Order("record_id DESC").Limit(limit).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// AddPoints 在事务中给用户增加积分并写入变动记录
func AddPoints(db *gorm.DB, userID int64, amount int, recordType, title, desc, orderID string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 锁定用户行，获取当前余额
		var user User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}

		newBalance := user.TokenBalance + amount

		// 更新余额
		if err := tx.Model(&User{}).Where("id = ?", userID).
			Update("token_balance", newBalance).Error; err != nil {
			return err
		}

		// 写入积分变动记录
		record := &TokenRecord{
			UserID:      userID,
			Type:        recordType,
			Amount:      amount,
			Balance:     newBalance,
			Title:       title,
			Description: desc,
			OrderID:     orderID,
		}
		return tx.Create(record).Error
	})
}
