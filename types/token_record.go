package types

import (
	"time"
)

// TokenRecord 用户代币记录表结构体
type TokenRecord struct {
	RecordID     int64             `gorm:"column:record_id;primaryKey;autoIncrement" json:"record_id"`                                                        // 记录ID，主键，自增
	UserID       int64             `gorm:"column:user_id;index:idx_token_records_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"`           // 用户ID
	ChangeAmount int               `gorm:"column:change_amount" json:"change_amount"`                                                                         // 代币变动数
	BalanceAfter int               `gorm:"column:balance_after" json:"balance_after"`                                                                         // 变动后余额
	ChangeType   string            `gorm:"column:change_type" json:"change_type"`                                                                             // 变动类型
	TaskID       *int              `gorm:"column:task_id;index:idx_token_records_task;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"task_id"`          // 任务ID来源
	FeatureID    *int              `gorm:"column:feature_id;index:idx_token_records_feature;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"feature_id"` // 功能ID来源
	OrderID      *int64            `gorm:"column:order_id;index:idx_token_records_order;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"order_id"`       // 订单ID来源
	AdminID      *int64            `gorm:"column:admin_id;index:idx_token_records_admin;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"admin_id"`       // 管理员ID来源
	Remark       string            `gorm:"column:remark" json:"remark"`                                                                                       // 备注说明
	ChangeTime   time.Time         `gorm:"column:change_time;default:CURRENT_TIMESTAMP" json:"change_time"`                                                   // 变动时间
	User         User              `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`            // 关联用户信息
	Task         *RewardTask       `gorm:"foreignKey:TaskID;references:TaskID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"task,omitempty"`           // 关联任务信息
	Feature      *TokenConsumeRule `gorm:"foreignKey:FeatureID;references:FeatureID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"feature,omitempty"`  // 关联功能信息
	Order        *RechargeOrder    `gorm:"foreignKey:OrderID;references:OrderID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"order,omitempty"`        // 关联订单信息
	Admin        *AdminUser        `gorm:"foreignKey:AdminID;references:AdminID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"admin,omitempty"`        // 关联管理员信息
}

// TableName 指定表名
func (TokenRecord) TableName() string {
	return "token_records"
}
