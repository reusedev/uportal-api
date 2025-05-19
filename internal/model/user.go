package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	UserID       int64          `gorm:"column:user_id;primarykey" json:"user_id"`
	Phone        string         `gorm:"column:phone;size:20;uniqueIndex:uk_users_phone" json:"phone"`
	Email        string         `gorm:"column:email;size:100;uniqueIndex:uk_users_email" json:"email"`
	PasswordHash string         `gorm:"column:password_hash;size:255" json:"-"`
	Nickname     string         `gorm:"column:nickname;size:50" json:"nickname"`
	AvatarURL    string         `gorm:"column:avatar_url;size:255" json:"avatar_url"`
	Language     string         `gorm:"column:language;size:10;not null;default:'zh-CN'" json:"language"`
	Status       int8           `gorm:"column:status;not null;default:1;index:idx_users_status" json:"status"`
	TokenBalance int64          `gorm:"column:token_balance;not null;default:0" json:"token_balance"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	LastLoginAt  *time.Time     `gorm:"column:last_login_at" json:"last_login_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserAuth 用户第三方认证模型
type UserAuth struct {
	AuthID         int64     `gorm:"column:auth_id;primarykey" json:"auth_id"`
	UserID         int64     `gorm:"column:user_id;not null;index:idx_user_auth_user" json:"user_id"`
	Provider       string    `gorm:"column:provider;size:20;not null" json:"provider"`
	ProviderUserID string    `gorm:"column:provider_user_id;size:100;not null" json:"provider_user_id"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定表名
func (UserAuth) TableName() string {
	return "user_auth"
}

// UserLoginLog 用户登录日志模型
type UserLoginLog struct {
	LogID         int64     `gorm:"column:log_id;primarykey" json:"log_id"`
	UserID        int64     `gorm:"column:user_id;not null;index:idx_login_log_user" json:"user_id"`
	LoginTime     time.Time `gorm:"column:login_time;not null;default:CURRENT_TIMESTAMP" json:"login_time"`
	LoginMethod   string    `gorm:"column:login_method;size:20;not null" json:"login_method"`
	LoginPlatform string    `gorm:"column:login_platform;size:20" json:"login_platform"`
	IPAddress     string    `gorm:"column:ip_address;size:45" json:"ip_address"`
	DeviceInfo    string    `gorm:"column:device_info;size:100" json:"device_info"`
}

// TableName 指定表名
func (UserLoginLog) TableName() string {
	return "user_login_log"
}

// CreateUser 创建用户
func CreateUser(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}

// GetUserByID 根据ID获取用户
func GetUserByID(db *gorm.DB, id int64) (*User, error) {
	var user User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone 根据手机号获取用户
func GetUserByPhone(db *gorm.DB, phone string) (*User, error) {
	var user User
	err := db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByProvider 根据第三方平台用户ID获取用户
func GetUserByProvider(db *gorm.DB, provider, providerUserID string) (*User, error) {
	var auth UserAuth
	err := db.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&auth).Error
	if err != nil {
		return nil, err
	}
	return GetUserByID(db, auth.UserID)
}

// UpdateUser 更新用户信息
func UpdateUser(db *gorm.DB, id int64, updates map[string]interface{}) error {
	return db.Model(&User{}).Where("user_id = ?", id).Updates(updates).Error
}

// DeleteUser 删除用户
func DeleteUser(db *gorm.DB, id int64) error {
	return db.Delete(&User{}, id).Error
}

// CreateUserAuth 创建第三方认证记录
func CreateUserAuth(db *gorm.DB, auth *UserAuth) error {
	return db.Create(auth).Error
}

// CreateLoginLog 创建登录日志
func CreateLoginLog(db *gorm.DB, log *UserLoginLog) error {
	return db.Create(log).Error
}

// UpdateLastLoginTime 更新最后登录时间
func UpdateLastLoginTime(db *gorm.DB, id int64) error {
	now := time.Now()
	return db.Model(&User{}).Where("user_id = ?", id).
		Update("last_login_at", now).Error
}

// UpdateTokenBalance 更新用户代币余额
func UpdateTokenBalance(db *gorm.DB, id int64, amount int) error {
	return db.Model(&User{}).Where("user_id = ?", id).
		Update("token_balance", gorm.Expr("token_balance + ?", amount)).Error
}
