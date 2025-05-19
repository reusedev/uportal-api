package types

import (
	"time"
)

// User 用户表结构体
type User struct {
	UserID       int64      `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`        // 用户ID，主键，自增
	Phone        string     `gorm:"column:phone;uniqueIndex:uk_users_phone" json:"phone"`          // 手机号
	Email        string     `gorm:"column:email;uniqueIndex:uk_users_email" json:"email"`          // 邮箱
	PasswordHash string     `gorm:"column:password_hash" json:"-"`                                 // 密码哈希
	Nickname     string     `gorm:"column:nickname" json:"nickname"`                               // 用户昵称
	AvatarURL    string     `gorm:"column:avatar_url" json:"avatar_url"`                           // 头像URL
	Language     string     `gorm:"column:language;default:zh-CN" json:"language"`                 // 界面语言偏好
	Status       int8       `gorm:"column:status;default:1;index:idx_users_status" json:"status"`  // 账号状态：1=正常，0=禁用
	TokenBalance int        `gorm:"column:token_balance;default:0" json:"token_balance"`           // 代币余额
	CreatedAt    time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"` // 注册时间
	UpdatedAt    time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"` // 记录更新时间
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at"`                     // 最后登录时间
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
