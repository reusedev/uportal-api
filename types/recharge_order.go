package types

import (
	"time"
)

// RechargeOrder 充值订单表结构体
type RechargeOrder struct {
	OrderID       int64         `gorm:"column:order_id;primaryKey;autoIncrement" json:"order_id"`                                                   // 订单ID，主键，自增
	UserID        int64         `gorm:"column:user_id;index:idx_recharge_orders_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"`  // 用户ID
	PlanID        *int          `gorm:"column:plan_id;index:idx_recharge_orders_plan;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"plan_id"` // 方案ID
	TokenAmount   int           `gorm:"column:token_amount" json:"token_amount"`                                                                    // 本次订单获得的代币数量
	AmountPaid    float64       `gorm:"column:amount_paid;type:decimal(10,2)" json:"amount_paid"`                                                   // 支付金额(元)
	PaymentMethod string        `gorm:"column:payment_method" json:"payment_method"`                                                                // 支付方式
	Status        int8          `gorm:"column:status;default:0" json:"status"`                                                                      // 订单状态：0=待支付，1=支付成功，2=支付失败，3=已退款
	TransactionID string        `gorm:"column:transaction_id" json:"transaction_id"`                                                                // 第三方交易号
	CreatedAt     time.Time     `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`                                              // 订单创建时间
	PaidAt        *time.Time    `gorm:"column:paid_at" json:"paid_at"`                                                                              // 支付完成时间
	UpdatedAt     time.Time     `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`                                              // 更新时间
	User          User          `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`     // 关联用户信息
	Plan          *RechargePlan `gorm:"foreignKey:PlanID;references:PlanID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"plan,omitempty"`    // 关联充值方案信息
}

// TableName 指定表名
func (RechargeOrder) TableName() string {
	return "recharge_orders"
}
