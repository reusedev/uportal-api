package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // 待支付
	OrderStatusPaid      OrderStatus = "paid"      // 已支付
	OrderStatusCompleted OrderStatus = "completed" // 已完成
	OrderStatusCancelled OrderStatus = "cancelled" // 已取消
	OrderStatusRefunded  OrderStatus = "refunded"  // 已退款
)

// Order 订单表结构体
type Order struct {
	OrderID     int64          `gorm:"column:order_id;primaryKey;autoIncrement" json:"order_id"`                                               // 订单ID，主键，自增
	UserID      int64          `gorm:"column:user_id;index:idx_orders_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"`       // 用户ID
	OrderNo     string         `gorm:"column:order_no;type:varchar(64);uniqueIndex:uk_orders_no" json:"order_no"`                              // 订单号
	Amount      float64        `gorm:"column:amount;type:decimal(10,2)" json:"amount"`                                                         // 订单金额
	ProductID   string         `gorm:"column:product_id;type:varchar(64)" json:"product_id"`                                                   // 商品ID
	ProductName string         `gorm:"column:product_name;type:varchar(64)" json:"product_name"`                                               // 商品名称
	Status      OrderStatus    `gorm:"column:status;type:varchar(20);default:pending" json:"status"`                                           // 订单状态
	PaymentInfo string         `gorm:"column:payment_info;type:json" json:"payment_info"`                                                      // 支付信息
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`                                                     // 创建时间
	PaidAt      *time.Time     `gorm:"column:paid_at" json:"paid_at"`                                                                          // 支付时间
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`                                                     // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                                                                                         // 删除时间
	User        User           `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"` // 关联用户信息
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// CreateOrder 创建订单
func CreateOrder(db *gorm.DB, order *Order) error {
	return db.Create(order).Error
}

// GetOrderByID 根据ID获取订单
func GetOrderByID(db *gorm.DB, orderID int64) (*Order, error) {
	var order Order
	err := db.Preload("User").First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func GetOrderByOrderNo(db *gorm.DB, orderNo string) (*Order, error) {
	var order Order
	err := db.Preload("User").Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder 更新订单
func UpdateOrder(db *gorm.DB, orderID int64, updates map[string]interface{}) error {
	return db.Model(&Order{}).Where("order_id = ?", orderID).Updates(updates).Error
}

// GetUserOrders 获取用户订单列表
func GetUserOrders(db *gorm.DB, userID int64, page, pageSize int) ([]*Order, int64, error) {
	var orders []*Order
	var total int64

	offset := (page - 1) * pageSize

	err := db.Model(&Order{}).Where("user_id = ?", userID).Count(&total).Error
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
func CanUpdateOrderStatus(oldStatus, newStatus OrderStatus) bool {
	switch oldStatus {
	case OrderStatusPending:
		return newStatus == OrderStatusPaid || newStatus == OrderStatusCancelled
	case OrderStatusPaid:
		return newStatus == OrderStatusCompleted || newStatus == OrderStatusRefunded
	case OrderStatusRefunded:
		return newStatus == OrderStatusCompleted
	default:
		return false
	}
}
