package model

import "gorm.io/gorm"

// GetAdinUserByID 根据ID获取管理员用户
func GetAdinUserByID(db *gorm.DB, id int64) (*AdminUser, error) {
	var user AdminUser
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateAdminUser 更新管理员用户信息
func UpdateAdminUser(db *gorm.DB, id int64, updates map[string]interface{}) error {
	return db.Model(&AdminUser{}).Where("admin_id = ?", id).Updates(updates).Error
}
