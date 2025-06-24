package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态

// '订单状态：0=待支付，1=支付成功，2=支付失败，3=已退款',
const (
	OrderStatusPending   int8 = 0 // 待支付
	OrderStatusCompleted int8 = 1 // 支付成功
	OrderStatusCancelled int8 = 2 // 支付失败
	OrderStatusRefunded  int8 = 3 // 已退款
)

// CreateOrder 创建订单
func CreateOrder(db *gorm.DB, order *RechargeOrder) error {
	return db.Create(order).Error
}

// GetOrderByID 根据ID获取订单
func GetOrderByID(db *gorm.DB, orderID int64) (*RechargeOrder, error) {
	var order RechargeOrder
	err := db.Preload("User").First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func GetOrderByOrderNo(db *gorm.DB, orderNo string) (*RechargeOrder, error) {
	var order RechargeOrder
	err := db.Preload("User").Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder 更新订单
func UpdateOrder(db *gorm.DB, orderID int64, updates map[string]interface{}) error {
	return db.Model(&RechargeOrder{}).Where("order_id = ?", orderID).Updates(updates).Error
}

// GetUserOrders 获取用户订单列表
func GetUserOrders(db *gorm.DB, userID int64, page, pageSize int) ([]*RechargeOrder, int64, error) {
	var orders []*RechargeOrder
	var total int64

	offset := (page - 1) * pageSize

	err := db.Model(&RechargeOrder{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Where("user_id = ?", userID).
		Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GenerateOrderNo 生成订单号
func GenerateOrderNo() string {
	return fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

// CanUpdateOrderStatus 检查订单状态是否可以更新
func CanUpdateOrderStatus(oldStatus, newStatus int8) bool {
	switch oldStatus {
	case OrderStatusPending:
		return newStatus == OrderStatusCancelled
	case OrderStatusRefunded:
		return newStatus == OrderStatusCompleted
	default:
		return false
	}
}
