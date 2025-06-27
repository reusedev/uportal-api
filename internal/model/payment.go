package model

import (
	"gorm.io/gorm"
)

// CreateNotifyRecord 创建通知记录
func CreateNotifyRecord(db *gorm.DB, record *PaymentNotifyRecord) error {
	return db.Create(record).Error
}

// GetNotifyRecord 获取通知记录
func GetNotifyRecord(db *gorm.DB, orderID string, transactionID string) (*PaymentNotifyRecord, error) {
	var record PaymentNotifyRecord
	err := db.Where("order_id = ? AND transaction_id = ?", orderID, transactionID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateNotifyRecord 更新通知记录
func UpdateNotifyRecord(db *gorm.DB, recordID int64, updates map[string]interface{}) error {
	return db.Model(&PaymentNotifyRecord{}).Where("record_id = ?", recordID).Updates(updates).Error
}

// ListPendingNotifyRecords 获取待处理的通知记录
func ListPendingNotifyRecords(db *gorm.DB, limit int) ([]*PaymentNotifyRecord, error) {
	var records []*PaymentNotifyRecord
	err := db.Where("process_status = 0 AND retry_count < 3").
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// 通知处理状态常量
const (
	NotifyStatusPending = 0 // 待处理
	NotifyStatusSuccess = 1 // 处理成功
	NotifyStatusFailed  = 2 // 处理失败
	MaxRetryCount       = 3 // 最大重试次数
)
