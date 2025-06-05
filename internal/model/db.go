package model

import (
	"fmt"
	"gorm.io/gorm/schema"
	"time"

	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/logs"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// DB 全局数据库连接
	DB *gorm.DB
)

// InitDB 初始化数据库连接
func InitDB() error {
	cfg := config.Get().Database

	// 构建DSN，添加外键约束检查参数
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&foreign_key_checks=1",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.New(
			&gormLogger{logger: logs.DB()},
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		DisableForeignKeyConstraintWhenMigrating: false, // 确保启用外键约束
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "",
		},
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("connect to database error: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database instance error: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping database error: %v", err)
	}

	// 设置全局数据库连接
	DB = db

	logs.Business().Info("Database connected successfully")
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("get database instance error: %v", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// gormLogger 实现 gorm.Logger 接口
type gormLogger struct {
	logger *zap.Logger
}

func (l *gormLogger) Printf(format string, args ...interface{}) {
	l.logger.Sugar().Infof(format, args...)
}
