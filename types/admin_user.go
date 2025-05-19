package types

import (
	"time"
)

// AdminUser 管理员用户表结构体
type AdminUser struct {
	AdminID      int        `gorm:"column:admin_id;primaryKey;autoIncrement" json:"admin_id"`      // 管理员ID，主键，自增
	Username     string     `gorm:"column:username;uniqueIndex:uk_admin_username" json:"username"` // 登录用户名
	PasswordHash string     `gorm:"column:password_hash" json:"-"`                                 // 密码哈希
	Role         string     `gorm:"column:role;default:admin" json:"role"`                         // 角色
	Status       int8       `gorm:"column:status;default:1" json:"status"`                         // 账号状态：1=正常，0=停用
	CreatedAt    time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at"`                     // 最后登录时间
}

// TableName 指定表名
func (AdminUser) TableName() string {
	return "admin_users"
}
