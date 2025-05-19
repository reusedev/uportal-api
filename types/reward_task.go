package types

import (
	"time"
)

// RewardTask 代币任务配置表结构体
type RewardTask struct {
	TaskID          int        `gorm:"column:task_id;primaryKey;autoIncrement" json:"task_id"`    // 任务ID，主键，自增
	TaskName        string     `gorm:"column:task_name" json:"task_name"`                         // 任务名称
	TaskDesc        string     `gorm:"column:task_desc" json:"task_desc"`                         // 任务描述
	TokenReward     int        `gorm:"column:token_reward" json:"token_reward"`                   // 完成一次任务获得的代币数
	DailyLimit      int        `gorm:"column:daily_limit;default:0" json:"daily_limit"`           // 每日奖励上限
	IntervalSeconds int        `gorm:"column:interval_seconds;default:0" json:"interval_seconds"` // 两次完成任务的最小间隔秒数
	ValidFrom       *time.Time `gorm:"column:valid_from" json:"valid_from"`                       // 任务生效时间
	ValidTo         *time.Time `gorm:"column:valid_to" json:"valid_to"`                           // 任务截止时间
	Repeatable      int8       `gorm:"column:repeatable;default:1" json:"repeatable"`             // 是否可重复完成：1=是，0=否
	Status          int8       `gorm:"column:status;default:1" json:"status"`                     // 任务状态：1=启用，0=停用
}

// TableName 指定表名
func (RewardTask) TableName() string {
	return "reward_tasks"
}
