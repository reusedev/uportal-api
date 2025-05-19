package types

import (
	"time"
)

// UserLoginLog 用户登录日志表结构体
type UserLoginLog struct {
	LogID         int64     `gorm:"column:log_id;primaryKey;autoIncrement" json:"log_id"`                                                   // 日志ID，主键，自增
	UserID        int64     `gorm:"column:user_id;index:idx_login_log_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"`    // 用户ID
	LoginTime     time.Time `gorm:"column:login_time;default:CURRENT_TIMESTAMP" json:"login_time"`                                          // 登录时间
	LoginMethod   string    `gorm:"column:login_method" json:"login_method"`                                                                // 登录方式
	LoginPlatform string    `gorm:"column:login_platform" json:"login_platform"`                                                            // 登录平台
	IPAddress     string    `gorm:"column:ip_address" json:"ip_address"`                                                                    // 登录IP地址
	DeviceInfo    string    `gorm:"column:device_info" json:"device_info"`                                                                  // 设备信息
	User          User      `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"` // 关联用户信息
}

// TableName 指定表名
func (UserLoginLog) TableName() string {
	return "user_login_log"
}
