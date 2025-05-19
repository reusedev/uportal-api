package model

import (
	"github.com/reusedev/uportal-api/types"
	"gorm.io/gorm"
)

// CreateOrder 创建订单
func CreateOrder(db *gorm.DB, order *types.RechargeOrder) error {
	return db.Create(order).Error
}

// GetOrder 获取订单
func GetOrder(db *gorm.DB, orderID int64) (*types.RechargeOrder, error) {
	var order types.RechargeOrder
	err := db.Preload("User").Preload("Plan").First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder 更新订单
func UpdateOrder(db *gorm.DB, orderID int64, updates map[string]interface{}) error {
	return db.Model(&types.RechargeOrder{}).Where("order_id = ?", orderID).Updates(updates).Error
}

// ListOrders 获取订单列表
func ListOrders(db *gorm.DB, userID int64, offset, limit int) ([]*types.RechargeOrder, int64, error) {
	var orders []*types.RechargeOrder
	var total int64

	query := db.Model(&types.RechargeOrder{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("User").Preload("Plan").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
