package model

import "gorm.io/gorm"

// CreateOrder 创建充值订单
func CreateOrder(db *gorm.DB, order *RechargeOrder) error {
	return db.Create(order).Error
}

// GetRechargePlans 获取所有启用的充值方案
func GetRechargePlans(db *gorm.DB) ([]*RechargePlan, error) {
	var plans []*RechargePlan
	err := db.Where("status = ?", 1).Find(&plans).Error
	if err != nil {
		return nil, err
	}
	return plans, nil
}
