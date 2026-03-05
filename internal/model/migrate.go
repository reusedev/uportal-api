package model

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
func Migrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	err := db.AutoMigrate(
		&User{},
		&RechargePlan{},
		&RechargeOrder{},
		&TokenRecord{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate tables: %v", err)
	}

	log.Println("Database migration completed successfully!")
	return nil
}
