package types

import (
	"time"
)

// RechargePlan 充值方案表结构体
type RechargePlan struct {
	PlanID      int       `gorm:"column:plan_id;primaryKey;autoIncrement" json:"plan_id"`        // 方案ID，主键，自增
	TokenAmount int       `gorm:"column:token_amount" json:"token_amount"`                       // 方案提供的代币数量
	Price       float64   `gorm:"column:price;type:decimal(10,2)" json:"price"`                  // 售价(元)
	Currency    string    `gorm:"column:currency;default:CNY" json:"currency"`                   // 货币类型代码
	Description string    `gorm:"column:description" json:"description"`                         // 方案描述
	Status      int8      `gorm:"column:status;default:1" json:"status"`                         // 方案状态：1=可用，0=下架
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"` // 更新时间
}

// TableName 指定表名
func (RechargePlan) TableName() string {
	return "recharge_plans"
}
