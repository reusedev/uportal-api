package types

import (
	"time"
)

// Refund 退款记录表结构体
type Refund struct {
	RefundID     int64         `gorm:"column:refund_id;primaryKey;autoIncrement" json:"refund_id"`                                                 // 退款ID，主键，自增
	OrderID      int64         `gorm:"column:order_id;index:idx_refunds_order;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order_id"`       // 原订单ID
	UserID       int64         `gorm:"column:user_id;index:idx_refunds_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"`          // 用户ID
	RefundAmount float64       `gorm:"column:refund_amount;type:decimal(10,2)" json:"refund_amount"`                                               // 退款金额(元)
	RefundTokens int           `gorm:"column:refund_tokens" json:"refund_tokens"`                                                                  // 收回代币数
	RefundMethod string        `gorm:"column:refund_method" json:"refund_method"`                                                                  // 退款方式
	Status       int8          `gorm:"column:status;default:0" json:"status"`                                                                      // 退款状态：0=处理中，1=成功，2=失败
	AdminID      *int          `gorm:"column:admin_id;index:idx_refunds_admin;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"admin_id"`      // 操作管理员ID
	Reason       string        `gorm:"column:reason" json:"reason"`                                                                                // 退款原因说明
	RefundTime   time.Time     `gorm:"column:refund_time;default:CURRENT_TIMESTAMP" json:"refund_time"`                                            // 退款完成时间
	User         User          `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`     // 关联用户信息
	Order        RechargeOrder `gorm:"foreignKey:OrderID;references:OrderID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order,omitempty"`  // 关联订单信息
	Admin        *AdminUser    `gorm:"foreignKey:AdminID;references:AdminID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"admin,omitempty"` // 关联管理员信息
}

// TableName 指定表名
func (Refund) TableName() string {
	return "refunds"
}
