package model

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
func Migrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	// 第一步：创建所有表，但不添加外键约束
	// 创建一个新的数据库连接，禁用外键约束
	newDB, err := gorm.Open(db.Dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("create new database connection error: %v", err)
	}

	// 创建所有表
	err = newDB.AutoMigrate(
		&User{},                // 基础用户表
		&AdminUser{},           // 管理员表
		&SystemConfig{},        // 系统配置表
		&RechargePlan{},        // 充值方案表
		&TokenConsumeRule{},    // 代币消耗规则表
		&RewardTask{},          // 奖励任务表
		&Order{},               // 订单表
		&UserAuth{},            // 用户认证表
		&UserLoginLog{},        // 用户登录日志表
		&RechargeOrder{},       // 充值订单表
		&Refund{},              // 退款记录表
		&TokenRecord{},         // 代币记录表
		&PaymentNotifyRecord{}, // 支付通知记录表
		&InviteRecord{},        // 邀请记录表
	)
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	// 第二步：添加外键约束
	// 注意：这里需要按照依赖关系的顺序添加约束
	constraints := []struct {
		table    string
		foreign  string
		refTable string
		refKey   string
	}{
		{"orders", "user_id", "users", "user_id"},
		{"user_auth", "user_id", "users", "user_id"},
		{"user_login_log", "user_id", "users", "user_id"},
		{"recharge_orders", "user_id", "users", "user_id"},
		{"recharge_orders", "plan_id", "recharge_plans", "plan_id"},
		{"refunds", "user_id", "users", "user_id"},
		{"refunds", "order_id", "recharge_orders", "order_id"},
		{"refunds", "admin_id", "admin_users", "admin_id"},
		{"token_records", "user_id", "users", "user_id"},
		{"token_records", "task_id", "reward_tasks", "task_id"},
		{"token_records", "feature_id", "token_consume_rules", "feature_id"},
		{"token_records", "order_id", "recharge_orders", "order_id"},
		{"token_records", "admin_id", "admin_users", "admin_id"},
		{"payment_notify_records", "order_id", "recharge_orders", "order_id"},
		{"invite_records", "inviter_id", "users", "user_id"},
		{"invite_records", "invitee_id", "users", "user_id"},
	}

	for _, c := range constraints {
		sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT fk_%s_%s FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE CASCADE ON UPDATE CASCADE",
			c.table, c.table, c.foreign, c.foreign, c.refTable, c.refKey)
		if err := db.Exec(sql).Error; err != nil {
			// 如果约束已存在，忽略错误
			if !strings.Contains(err.Error(), "Duplicate key name") {
				return fmt.Errorf("failed to add foreign key constraint for %s.%s: %v", c.table, c.foreign, err)
			}
		}
	}

	// 初始化基础数据
	if err := initBaseData(db); err != nil {
		return fmt.Errorf("failed to initialize base data: %v", err)
	}

	log.Println("Database migration completed successfully!")
	return nil
}

// initBaseData 初始化基础数据
func initBaseData(db *gorm.DB) error {
	// 检查是否已经存在管理员账号
	var count int64
	if err := db.Model(&AdminUser{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果不存在管理员账号，创建默认管理员
	if count == 0 {
		admin := &AdminUser{
			Username:     "admin",
			PasswordHash: "$2a$10$uXh03r/wuLRWVvTQNzUthugLWcVDE5mburOLWvql0FK5lUKK2owqa", // admin123
			Role:         "super_admin",
			Status:       1,
		}
		if err := db.Create(admin).Error; err != nil {
			return err
		}
		log.Println("Created default admin user")
	}

	// 检查并初始化系统配置
	var configCount int64
	if err := db.Model(&SystemConfig{}).Count(&configCount).Error; err != nil {
		return err
	}

	if configCount == 0 {
		siteName := "站点名称"
		siteDesc := "站点描述"
		tokenPrice := "Token单价（元）"
		minRecharge := "最小充值金额（元）"
		maxRecharge := "最大充值金额（元）"

		configs := []SystemConfig{
			{
				ConfigKey:   "site_name",
				ConfigValue: "U Portal",
				Description: &siteName,
			},
			{
				ConfigKey:   "site_description",
				ConfigValue: "AI 对话平台",
				Description: &siteDesc,
			},
			{
				ConfigKey:   "token_price",
				ConfigValue: "0.1",
				Description: &tokenPrice,
			},
			{
				ConfigKey:   "min_recharge_amount",
				ConfigValue: "10",
				Description: &minRecharge,
			},
			{
				ConfigKey:   "max_recharge_amount",
				ConfigValue: "10000",
				Description: &maxRecharge,
			},
		}
		if err := db.Create(&configs).Error; err != nil {
			return err
		}
		log.Println("Initialized system configurations")
	}

	// 检查并初始化Token消耗规则
	var ruleCount int64
	if err := db.Model(&TokenConsumeRule{}).Count(&ruleCount).Error; err != nil {
		return err
	}

	if ruleCount == 0 {
		chatDesc := "AI对话功能"
		chatCode := "chat"
		imageDesc := "AI绘图功能"
		imageCode := "image"

		rules := []TokenConsumeRule{
			{
				FeatureName: "AI对话",
				FeatureDesc: &chatDesc,
				TokenCost:   1,
				FeatureCode: &chatCode,
				Status:      1,
			},
			{
				FeatureName: "AI绘图",
				FeatureDesc: &imageDesc,
				TokenCost:   10,
				FeatureCode: &imageCode,
				Status:      1,
			},
		}
		if err := db.Create(&rules).Error; err != nil {
			return err
		}
		log.Println("Initialized token consume rules")
	}

	// 检查并初始化奖励任务
	var taskCount int64
	if err := db.Model(&RewardTask{}).Count(&taskCount).Error; err != nil {
		return err
	}

	if taskCount == 0 {
		dailyLoginDesc := "每日登录奖励"
		shareDesc := "分享奖励"
		inviteDesc := "邀请新用户奖励"

		tasks := []RewardTask{
			{
				TaskName:        "每日登录",
				TaskKey:         "daily_login",
				TaskDesc:        &dailyLoginDesc,
				TokenReward:     10,
				DailyLimit:      1,
				IntervalSeconds: 86400, // 24小时
				Repeatable:      1,
				Status:          1,
			},
			{
				TaskName:        "分享",
				TaskKey:         "share",
				TaskDesc:        &shareDesc,
				TokenReward:     5,
				DailyLimit:      3,
				IntervalSeconds: 3600, // 1小时
				Repeatable:      1,
				Status:          1,
			},
			{
				TaskName:        "邀请新用户",
				TaskKey:         "invite",
				TaskDesc:        &inviteDesc,
				TokenReward:     50,
				DailyLimit:      0, // 不限制每日次数
				IntervalSeconds: 0, // 不限制间隔
				Repeatable:      1,
				Status:          1,
			},
		}
		if err := db.Create(&tasks).Error; err != nil {
			return err
		}
		log.Println("Initialized reward tasks")
	}

	return nil
}
