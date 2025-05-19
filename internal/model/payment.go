package model

import (
	"time"

	"github.com/reusedev/uportal-api/types"
	"gorm.io/gorm"
)

// PaymentNotifyRecord 支付回调通知记录
type PaymentNotifyRecord struct {
	RecordID      int64                `gorm:"column:record_id;primaryKey;autoIncrement" json:"record_id"`
	OrderID       int64                `gorm:"column:order_id;uniqueIndex:uk_order_transaction" json:"order_id"`
	TransactionID string               `gorm:"column:transaction_id;uniqueIndex:uk_order_transaction" json:"transaction_id"`
	NotifyType    string               `gorm:"column:notify_type" json:"notify_type"`
	NotifyTime    time.Time            `gorm:"column:notify_time" json:"notify_time"`
	ProcessStatus int8                 `gorm:"column:process_status;default:0" json:"process_status"`
	RetryCount    int                  `gorm:"column:retry_count;default:0" json:"retry_count"`
	ErrorMessage  string               `gorm:"column:error_message" json:"error_message"`
	ProcessTime   *time.Time           `gorm:"column:process_time" json:"process_time"`
	CreatedAt     time.Time            `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time            `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Order         *types.RechargeOrder `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// TableName 指定表名
func (PaymentNotifyRecord) TableName() string {
	return "payment_notify_records"
}

// CreateNotifyRecord 创建通知记录
func CreateNotifyRecord(db *gorm.DB, record *PaymentNotifyRecord) error {
	return db.Create(record).Error
}

// GetNotifyRecord 获取通知记录
func GetNotifyRecord(db *gorm.DB, orderID int64, transactionID string) (*PaymentNotifyRecord, error) {
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
