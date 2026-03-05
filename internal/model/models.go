package model

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// User Telegram 用户表
type User struct {
	UserID                int64          `gorm:"column:id;primaryKey" json:"id"`
	FirstName             string         `gorm:"column:first_name;type:varchar(100)" json:"first_name"`
	LastName              string         `gorm:"column:last_name;type:varchar(100)" json:"last_name"`
	Username              string         `gorm:"column:username;type:varchar(100)" json:"username"`
	LanguageCode          string         `gorm:"column:language_code;type:varchar(10)" json:"language_code"`
	IsBot                 bool           `gorm:"column:is_bot" json:"is_bot"`
	IsPremium             bool           `gorm:"column:is_premium" json:"is_premium"`
	AllowsWriteToPm       bool           `gorm:"column:allows_write_to_pm" json:"allows_write_to_pm"`
	AddedToAttachmentMenu bool           `gorm:"column:added_to_attachment_menu" json:"added_to_attachment_menu"`
	PhotoURL              string         `gorm:"column:photo_url;type:varchar(500)" json:"photo_url"`
	TokenBalance          int            `gorm:"column:token_balance;not null;default:0" json:"token_balance"`
	Status                int8           `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt             time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

// RechargePlan 充值方案表（适配 Telegram Stars）
type RechargePlan struct {
	PlanID      int       `gorm:"column:plan_id;primaryKey;autoIncrement" json:"-"`
	Points      int       `gorm:"column:points;not null" json:"points"`
	Stars       int       `gorm:"column:stars;not null" json:"stars"`
	Title       string    `gorm:"column:title;type:varchar(100);not null" json:"title"`
	Description string    `gorm:"column:description;type:varchar(255)" json:"description"`
	Bonus       int       `gorm:"column:bonus;not null;default:0" json:"bonus"`
	Popular     bool      `gorm:"column:popular;not null;default:false" json:"popular"`
	Status      int8      `gorm:"column:status;not null;default:1" json:"-"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"-"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"-"`
}

// MarshalJSON 自定义序列化，将 PlanID 转为 string
func (p RechargePlan) MarshalJSON() ([]byte, error) {
	type Alias RechargePlan
	return json.Marshal(struct {
		ID string `json:"id"`
		Alias
	}{
		ID:    fmt.Sprintf("%d", p.PlanID),
		Alias: Alias(p),
	})
}

// RechargeOrder 充值订单表（适配 Telegram Stars 上报）
type RechargeOrder struct {
	OrderID         string    `gorm:"column:order_id;type:varchar(100);primaryKey" json:"order_id"`
	UserID          int64     `gorm:"column:user_id;not null;index" json:"user_id"`
	PlanID          string    `gorm:"column:plan_id;type:varchar(100)" json:"plan_id"`
	Points          int       `gorm:"column:points;not null" json:"points"`
	Stars           int       `gorm:"column:stars;not null" json:"stars"`
	PaymentChargeID string    `gorm:"column:payment_charge_id;type:varchar(100)" json:"payment_charge_id"`
	PaidAt          string    `gorm:"column:paid_at;type:varchar(50)" json:"paid_at"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TokenRecord 积分变动记录表
type TokenRecord struct {
	RecordID    int64     `gorm:"column:record_id;primaryKey;autoIncrement" json:"-"`
	UserID      int64     `gorm:"column:user_id;not null;index" json:"user_id"`
	Type        string    `gorm:"column:type;type:varchar(20);not null" json:"type"`
	Amount      int       `gorm:"column:amount;not null" json:"amount"`
	Balance     int       `gorm:"column:balance;not null" json:"balance"`
	Title       string    `gorm:"column:title;type:varchar(100)" json:"title"`
	Description string    `gorm:"column:description;type:varchar(255)" json:"description"`
	OrderID     string    `gorm:"column:order_id;type:varchar(100)" json:"order_id"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// MarshalJSON 自定义序列化，将 RecordID 转为 string
func (r TokenRecord) MarshalJSON() ([]byte, error) {
	type Alias TokenRecord
	return json.Marshal(struct {
		ID string `json:"id"`
		Alias
	}{
		ID:    fmt.Sprintf("%d", r.RecordID),
		Alias: Alias(r),
	})
}

func (User) TableName() string         { return "users" }
func (RechargePlan) TableName() string  { return "recharge_plans" }
func (RechargeOrder) TableName() string { return "recharge_orders" }
func (TokenRecord) TableName() string   { return "token_records" }
