package model

import (
	"time"

	"gorm.io/gorm"
)

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
