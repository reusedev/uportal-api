package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/logs"
	"github.com/robfig/cron"
	"go.uber.org/zap"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "config/config.yaml", "配置文件路径")
	flag.Parse()
}

func main() {
	// 1. 加载配置
	if err := config.LoadConfig(configPath); err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}
	cfg := config.Get()

	// 2. 初始化日志
	if err := logs.Init(&logs.Config{
		LogDir:          cfg.Logging.LogDir,
		BusinessLogFile: cfg.Logging.BusinessLogFile,
		DBLogFile:       cfg.Logging.DBLogFile,
		Level:           cfg.Logging.Level,
		Console:         cfg.Logging.Console,
		MaxSize:         cfg.Logging.MaxSize,
		MaxBackups:      cfg.Logging.MaxBackups,
		MaxAge:          cfg.Logging.MaxAge,
		Compress:        cfg.Logging.Compress,
	}); err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}
	defer logs.Sync()

	// 3. 初始化数据库
	if err := model.InitDB(); err != nil {
		logs.Business().Fatal("初始化数据库失败", zap.Error(err))
	}
	defer model.CloseDB()

	wechat_token.TokenJob()

	// 4. 创建消息发送处理器
	processor := service.NewNotifyProcessor(model.DB)
	if processor == nil {
		logs.Business().Fatal("创建通知处理器失败")
	}

	// 5. 创建定时任务
	c := cron.New()

	// 每5分钟执行一次，使用带超时的context
	c.AddFunc("@every 5m", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()
		if err := processor.ProcessNotifications(ctx); err != nil {
			logs.Business().Error("处理通知任务失败", zap.Error(err))
		}
	})

	// 启动定时任务
	c.Start()
	defer c.Stop()
	logs.Business().Info("后台通知服务已启动，每5分钟执行一次")

	// 6. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logs.Business().Info("正在关闭后台服务...")

	// 停止处理器并保存当前进度
	processor.Stop()

	logs.Business().Info("后台服务已关闭")
}
