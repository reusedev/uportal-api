package types

import (
	"time"
)

// UserAuth 用户第三方认证表结构体
type UserAuth struct {
	AuthID         int64     `gorm:"column:auth_id;primaryKey;autoIncrement" json:"auth_id"`                                                 // 认证记录ID，主键，自增
	UserID         int64     `gorm:"column:user_id;index:idx_user_auth_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"`    // 用户ID
	Provider       string    `gorm:"column:provider" json:"provider"`                                                                        // 登录平台类型
	ProviderUserID string    `gorm:"column:provider_user_id" json:"provider_user_id"`                                                        // 第三方平台内用户唯一ID
	CreatedAt      time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`                                          // 绑定时间
	User           User      `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"` // 关联用户信息
}

// TableName 指定表名
func (UserAuth) TableName() string {
	return "user_auth"
}
