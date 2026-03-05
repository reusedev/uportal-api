package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertUser 根据 Telegram user ID 插入或更新用户信息
func UpsertUser(db *gorm.DB, user *User) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"first_name", "last_name", "username", "language_code",
			"is_bot", "is_premium", "allows_write_to_pm",
			"added_to_attachment_menu", "photo_url",
		}),
	}).Create(user).Error
}
