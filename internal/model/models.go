package model

import (
	"encoding/json"
	"time"

	"github.com/reusedev/uportal-api/pkg/constants"
	"gorm.io/gorm"
)

// User 用户表结构体
type User struct {
	UserID       string         `gorm:"column:user_id;type:varchar(13);primaryKey" json:"id"`                    // 用户ID，主键，自增
	Phone        *string        `gorm:"column:phone;type:varchar(20);uniqueIndex:uk_users_phone" json:"phone"`   // 手机号
	Email        *string        `gorm:"column:email;type:varchar(100);uniqueIndex:uk_users_email" json:"email"`  // 邮箱
	PasswordHash *string        `gorm:"column:password_hash;type:varchar(255)" json:"-"`                         // 密码哈希
	Nickname     *string        `gorm:"column:nickname;type:varchar(50)" json:"nickname"`                        // 用户昵称
	AvatarURL    *string        `gorm:"column:avatar_url;type:varchar(255)" json:"avatar"`                       // 头像URL
	Language     string         `gorm:"column:language;type:varchar(10);not null;default:zh-CN" json:"language"` // 界面语言偏好
	Status       int8           `gorm:"column:status;not null;default:1;index:idx_users_status" json:"status"`   // 账号状态：1=正常，0=禁用
	TokenBalance int            `gorm:"column:token_balance;not null;default:0" json:"token_balance"`            // 代币余额
	InviterID    *int64         `gorm:"column:inviter_id;index:idx_users_inviter" json:"inviter_id"`             // 邀请人ID
	CreatedAt    time.Time      `gorm:"column:created_at;not null;autoCreateTime" json:"-"`                      // 注册时间
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`             // 记录更新时间
	LastLoginAt  *time.Time     `gorm:"column:last_login_at" json:"last_login_at"`                               // 最后登录时间
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	UserAuths    []UserAuth     `gorm:"foreignKey:UserID" json:"-"`                                      // 第三方认证信息（不直接序列化）
	Inviter      *User          `gorm:"foreignKey:InviterID;references:UserID" json:"inviter,omitempty"` // 邀请人信息
}

// MarshalJSON 自定义 JSON 序列化方法
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User // 创建别名以避免递归调用
	providers := make([]string, 0, len(u.UserAuths))
	for _, auth := range u.UserAuths {
		providers = append(providers, auth.Provider)
	}

	return json.Marshal(struct {
		Alias
		Auths     []string `json:"auth_providers"`
		CreatedAt string   `json:"created_at"`
	}{
		Alias:     Alias(u),
		Auths:     providers,
		CreatedAt: u.CreatedAt.Format(constants.TimeFormatDateTime),
	})
}

// AdminUser 管理员用户表结构体
type AdminUser struct {
	AdminID      int        `gorm:"column:admin_id;primaryKey;autoIncrement" json:"admin_id"`                                // 管理员ID，主键，自增
	Username     string     `gorm:"column:username;type:varchar(50);not null;uniqueIndex:uk_admin_username" json:"username"` // 登录用户名
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`                                // 密码哈希
	Role         string     `gorm:"column:role;type:varchar(20);not null;default:admin" json:"role"`                         // 角色
	Status       int8       `gorm:"column:status;not null;default:1" json:"status"`                                          // 账号状态：1=正常，0=停用
	CreatedAt    time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"-"`                                      // 创建时间
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"-"`                                                           // 最后登录时间
}

// MarshalJSON 自定义 JSON 序列化方法
func (u AdminUser) MarshalJSON() ([]byte, error) {
	type Alias AdminUser
	var lastLoginAt string
	if u.LastLoginAt != nil {
		lastLoginAt = u.LastLoginAt.Format(constants.TimeFormatDateTime)
	}
	// 创建别名以避免递归调用
	return json.Marshal(struct {
		Alias
		CreatedAt   string `json:"created_at"`
		LastLoginAt string `json:"last_login_at"`
	}{
		Alias:       Alias(u),
		CreatedAt:   u.CreatedAt.Format(constants.TimeFormatDateTime),
		LastLoginAt: lastLoginAt,
	})
}

// UserAuth 用户第三方认证表结构体
type UserAuth struct {
	AuthID         int64     `gorm:"column:auth_id;primaryKey;autoIncrement" json:"auth_id"`                                                                        // 认证记录ID，主键，自增
	UserID         string    `gorm:"column:user_id;type:varchar(13);not null;index:idx_user_auth_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"` // 用户ID
	Provider       string    `gorm:"column:provider;type:varchar(20);not null" json:"provider"`                                                                     // 登录平台类型
	ProviderUserID string    `gorm:"column:provider_user_id;type:varchar(100);not null" json:"provider_user_id"`                                                    // 第三方平台内用户唯一ID
	CreatedAt      time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`                                                                   // 绑定时间
	User           User      `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`                        // 关联用户信息
}

// UserLoginLog 用户登录日志表结构体
type UserLoginLog struct {
	LogID         int64     `gorm:"column:log_id;primaryKey;autoIncrement" json:"log_id"`                                                                          // 日志ID，主键，自增
	UserID        string    `gorm:"column:user_id;type:varchar(13);not null;index:idx_login_log_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"` // 用户ID
	LoginTime     time.Time `gorm:"column:login_time;not null;autoCreateTime" json:"login_time"`                                                                   // 登录时间
	LoginMethod   string    `gorm:"column:login_method;type:varchar(20);not null" json:"login_method"`                                                             // 登录方式
	LoginPlatform *string   `gorm:"column:login_platform;type:varchar(20)" json:"login_platform"`                                                                  // 登录平台
	IPAddress     *string   `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`                                                                          // 登录IP地址
	DeviceInfo    *string   `gorm:"column:device_info;type:varchar(100)" json:"device_info"`                                                                       // 设备信息
	User          User      `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`                        // 关联用户信息
}

// RechargePlan 充值方案表结构体
type RechargePlan struct {
	PlanID      int       `gorm:"column:plan_id;primaryKey;autoIncrement" json:"plan_id"`            // 方案ID，主键，自增
	TokenAmount int       `gorm:"column:token_amount;not null" json:"token_amount"`                  // 方案提供的代币数量
	Price       float64   `gorm:"column:price;type:decimal(10,2);not null" json:"price"`             // 售价(元)
	Currency    string    `gorm:"column:currency;type:char(3);not null;default:CNY" json:"currency"` // 货币类型代码
	Description *string   `gorm:"column:description;type:varchar(100)" json:"description"`           // 方案描述
	Status      int8      `gorm:"column:status;not null;default:1" json:"status"`                    // 方案状态：1=可用，0=下架
	CreatedAt   time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`       // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`       // 更新时间
}

// RechargeOrder 充值订单表结构体
type RechargeOrder struct {
	OrderID       int64         `gorm:"column:order_id;primaryKey;autoIncrement" json:"order_id"`                                                                            // 订单ID，主键，自增
	UserID        string        `gorm:"column:user_id;type:varchar(13);not null;index:idx_recharge_orders_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"` // 用户ID
	PlanID        *int          `gorm:"column:plan_id;index:idx_recharge_orders_plan;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"plan_id"`                          // 方案ID
	TokenAmount   int           `gorm:"column:token_amount;not null" json:"token_amount"`                                                                                    // 本次订单获得的代币数量
	AmountPaid    float64       `gorm:"column:amount_paid;type:decimal(10,2);not null" json:"amount_paid"`                                                                   // 支付金额(元)
	PaymentMethod string        `gorm:"column:payment_method;type:varchar(20);not null" json:"payment_method"`                                                               // 支付方式
	Status        int8          `gorm:"column:status;not null;default:0" json:"status"`                                                                                      // 订单状态：0=待支付，1=支付成功，2=支付失败，3=已退款
	TransactionID *string       `gorm:"column:transaction_id;type:varchar(100)" json:"transaction_id"`                                                                       // 第三方交易号
	CreatedAt     time.Time     `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`                                                                         // 订单创建时间
	PaidAt        *time.Time    `gorm:"column:paid_at" json:"paid_at"`                                                                                                       // 支付完成时间
	UpdatedAt     time.Time     `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`                                                                         // 更新时间
	User          User          `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`                              // 关联用户信息
	Plan          *RechargePlan `gorm:"foreignKey:PlanID;references:PlanID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"plan,omitempty"`                             // 关联充值方案信息
}

// Refund 退款记录表结构体
type Refund struct {
	RefundID     int64         `gorm:"column:refund_id;primaryKey;autoIncrement" json:"refund_id"`                                                                  // 退款ID，主键，自增
	OrderID      int64         `gorm:"column:order_id;not null;index:idx_refunds_order;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order_id"`               // 原订单ID
	UserID       string        `gorm:"column:user_id;type:varchar(13);not null;index:idx_refunds_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"` // 用户ID
	RefundAmount float64       `gorm:"column:refund_amount;type:decimal(10,2);not null" json:"refund_amount"`                                                       // 退款金额(元)
	RefundTokens int           `gorm:"column:refund_tokens;not null" json:"refund_tokens"`                                                                          // 收回代币数
	RefundMethod string        `gorm:"column:refund_method;type:varchar(20);not null" json:"refund_method"`                                                         // 退款方式
	Status       int8          `gorm:"column:status;not null;default:0" json:"status"`                                                                              // 退款状态：0=处理中，1=成功，2=失败
	AdminID      *int          `gorm:"column:admin_id;index:idx_refunds_admin;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"admin_id"`                       // 操作管理员ID
	Reason       *string       `gorm:"column:reason;type:varchar(255)" json:"reason"`                                                                               // 退款原因说明
	RefundTime   time.Time     `gorm:"column:refund_time;not null;autoCreateTime" json:"refund_time"`                                                               // 退款完成时间
	User         User          `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`                      // 关联用户信息
	Order        RechargeOrder `gorm:"foreignKey:OrderID;references:OrderID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order,omitempty"`                   // 关联订单信息
	Admin        *AdminUser    `gorm:"foreignKey:AdminID;references:AdminID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"admin,omitempty"`                  // 关联管理员信息
}

// TokenConsumeRule 代币消耗功能表结构体
type TokenConsumeRule struct {
	FeatureID   int     `gorm:"column:feature_id;primaryKey;autoIncrement" json:"feature_id"`       // 功能ID，主键，自增
	FeatureName string  `gorm:"column:feature_name;type:varchar(100);not null" json:"feature_name"` // 功能名称
	FeatureDesc *string `gorm:"column:feature_desc;type:varchar(255)" json:"feature_desc"`          // 功能描述
	TokenCost   int     `gorm:"column:token_cost;not null" json:"token_cost"`                       // 使用一次该功能消耗的代币数
	FeatureCode *string `gorm:"column:feature_code;type:varchar(50)" json:"feature_code"`           // 功能代码
	Status      int8    `gorm:"column:status;not null;default:1" json:"status"`                     // 功能状态：1=启用，0=停用
}

// TokenRecord 用户代币记录表结构体
type TokenRecord struct {
	RecordID     int64             `gorm:"column:record_id;primaryKey;autoIncrement" json:"id"`                                                                               // 记录ID，主键，自增
	UserID       string            `gorm:"column:user_id;type:varchar(13);not null;index:idx_token_records_user;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"` // 用户ID
	ChangeAmount int               `gorm:"column:change_amount;not null" json:"points"`                                                                                       // 代币变动数
	BalanceAfter int               `gorm:"column:balance_after;not null" json:"balance_after"`                                                                                // 变动后余额
	ChangeType   string            `gorm:"column:change_type;type:varchar(20);not null" json:"source"`                                                                        // 变动类型
	TaskID       *int              `gorm:"column:task_id;index:idx_token_records_task;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"task_id"`                          // 任务ID来源
	FeatureID    *int              `gorm:"column:feature_id;index:idx_token_records_feature;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"feature_id"`                 // 功能ID来源
	OrderID      *int64            `gorm:"column:order_id;index:idx_token_records_order;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"order_id"`                       // 订单ID来源
	AdminID      *int64            `gorm:"column:admin_id;index:idx_token_records_admin;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"admin_id"`                       // 管理员ID来源
	Remark       *string           `gorm:"column:remark;type:varchar(255)" json:"remark"`                                                                                     // 备注说明
	ChangeTime   time.Time         `gorm:"column:change_time;not null;autoCreateTime" json:"created_at"`                                                                      // 变动时间
	User         User              `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user,omitempty"`                            // 关联用户信息
	Task         *RewardTask       `gorm:"foreignKey:TaskID;references:TaskID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"task,omitempty"`                           // 关联任务信息
	Feature      *TokenConsumeRule `gorm:"foreignKey:FeatureID;references:FeatureID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"feature,omitempty"`                  // 关联功能信息
	Order        *RechargeOrder    `gorm:"foreignKey:OrderID;references:OrderID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"order,omitempty"`                        // 关联订单信息
	Admin        *AdminUser        `gorm:"foreignKey:AdminID;references:AdminID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE" json:"admin,omitempty"`                        // 关联管理员信息
}

// RewardTask 代币任务配置表结构体
type RewardTask struct {
	TaskID          int        `gorm:"column:task_id;primaryKey;autoIncrement" json:"task_id"`             // 任务ID，主键，自增
	TaskName        string     `gorm:"column:task_name;type:varchar(100);not null" json:"task_name"`       // 任务名称
	TaskDesc        *string    `gorm:"column:task_desc;type:varchar(255)" json:"task_desc"`                // 任务描述
	TokenReward     int        `gorm:"column:token_reward;not null" json:"token_reward"`                   // 完成一次任务获得的代币数
	DailyLimit      int        `gorm:"column:daily_limit;not null;default:0" json:"daily_limit"`           // 每日奖励上限
	IntervalSeconds int        `gorm:"column:interval_seconds;not null;default:0" json:"interval_seconds"` // 两次完成任务的最小间隔秒数
	ValidFrom       *time.Time `gorm:"column:valid_from;type:date" json:"-"`                               // 任务生效时间
	ValidTo         *time.Time `gorm:"column:valid_to;type:date" json:"-"`                                 // 任务截止时间
	Repeatable      int8       `gorm:"column:repeatable;not null;default:1" json:"repeatable"`             // 是否可重复完成：1=是，0=否
	Status          int8       `gorm:"column:status;not null;default:1" json:"status"`                     // 任务状态：1=启用，0=停用
}

func (t RewardTask) MarshalJSON() ([]byte, error) {
	type Alias RewardTask // 创建别名以避免递归调用
	var validFrom, validTo string
	if t.ValidFrom != nil {
		validFrom = t.ValidFrom.Format(time.DateOnly)
	}
	if t.ValidTo != nil {
		validTo = t.ValidTo.Format(time.DateOnly)
	}

	return json.Marshal(struct {
		Alias
		ValidFrom string `json:"valid_from"`
		ValidTo   string `json:"valid_to"`
	}{
		Alias:     Alias(t),
		ValidFrom: validFrom,
		ValidTo:   validTo,
	})
}

// SystemConfig 系统配置表结构体
type SystemConfig struct {
	ConfigKey   string  `gorm:"column:config_key;type:varchar(50);primaryKey;not null" json:"config_key"` // 配置键，主键
	ConfigValue string  `gorm:"column:config_value;type:varchar(100);not null" json:"config_value"`       // 配置值
	Description *string `gorm:"column:description;type:varchar(100)" json:"description"`                  // 配置描述
}

// PaymentNotifyRecord 支付回调通知记录
type PaymentNotifyRecord struct {
	RecordID      int64          `gorm:"column:record_id;primaryKey;autoIncrement" json:"record_id"`
	OrderID       int64          `gorm:"column:order_id;not null;uniqueIndex:uk_order_transaction" json:"order_id"`
	TransactionID string         `gorm:"column:transaction_id;type:varchar(64);not null;uniqueIndex:uk_order_transaction" json:"transaction_id"`
	NotifyType    string         `gorm:"column:notify_type;type:varchar(32);not null" json:"notify_type"`
	NotifyTime    time.Time      `gorm:"column:notify_time;not null;autoCreateTime" json:"notify_time"`
	ProcessStatus int8           `gorm:"column:process_status;not null;default:0" json:"process_status"`
	RetryCount    int            `gorm:"column:retry_count;not null;default:0" json:"retry_count"`
	ErrorMessage  *string        `gorm:"column:error_message;type:varchar(255)" json:"error_message"`
	ProcessTime   *time.Time     `gorm:"column:process_time" json:"process_time"`
	CreatedAt     time.Time      `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	Order         *RechargeOrder `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"order,omitempty"`
}

// TaskCompletionRecord 任务完成记录
type TaskCompletionRecord struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      string    `gorm:"column:user_id;type:varchar(13);not null" json:"user_id"`
	TaskID      int       `gorm:"column:task_id;not null" json:"task_id"`
	TokenReward int       `gorm:"column:token_reward;not null" json:"token_reward"`
	CompletedAt time.Time `gorm:"column:completed_at;not null;default:CURRENT_TIMESTAMP" json:"completed_at"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

// Notification 通知
type Notification struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"column:user_id;type:varchar(13);not null" json:"user_id"`
	Type      string    `gorm:"column:type;not null;size:32" json:"type"`
	Title     string    `gorm:"column:title;not null;size:128" json:"title"`
	Content   string    `gorm:"column:content;not null;type:text" json:"content"`
	Status    int8      `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

// InviteRecord 邀请记录表结构体
type InviteRecord struct {
	RecordID    int64     `gorm:"column:record_id;primaryKey;autoIncrement" json:"record_id"`                             // 记录ID，主键，自增
	InviterID   string    `gorm:"column:inviter_id;size:13;not null;index:idx_invite_inviter(13)" json:"inviter_id"`      // 邀请人ID
	InviteeID   string    `gorm:"column:invitee_id;size:13;not null;uniqueIndex:uk_invite_invitee(13)" json:"invitee_id"` // 被邀请人ID
	TokenReward int       `gorm:"column:token_reward;not null" json:"token_reward"`                                       // 邀请奖励代币数
	Status      int8      `gorm:"column:status;not null;default:0" json:"status"`                                         // 状态：0=待发放，1=已发放，2=发放失败
	CreatedAt   time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`                            // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`                            // 更新时间
	Inviter     User      `gorm:"foreignKey:InviterID;references:UserID" json:"inviter,omitempty"`                        // 邀请人信息
	Invitee     User      `gorm:"foreignKey:InviteeID;references:UserID" json:"invitee,omitempty"`                        // 被邀请人信息
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (AdminUser) TableName() string {
	return "admin_users"
}

func (UserAuth) TableName() string {
	return "user_auth"
}

func (UserLoginLog) TableName() string {
	return "user_login_log"
}

func (RechargePlan) TableName() string {
	return "recharge_plans"
}

func (RechargeOrder) TableName() string {
	return "recharge_orders"
}

func (Refund) TableName() string {
	return "refunds"
}

func (TokenConsumeRule) TableName() string {
	return "token_consume_rules"
}

func (TokenRecord) TableName() string {
	return "token_records"
}

func (RewardTask) TableName() string {
	return "reward_tasks"
}

func (SystemConfig) TableName() string {
	return "system_config"
}

func (PaymentNotifyRecord) TableName() string {
	return "payment_notify_records"
}

func (TaskCompletionRecord) TableName() string {
	return "task_completion_records"
}

func (Notification) TableName() string {
	return "notifications"
}

func (InviteRecord) TableName() string {
	return "invite_records"
}

// StringPtr 创建字符串指针
func StringPtr(s string) *string {
	return &s
}
