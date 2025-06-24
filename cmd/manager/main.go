package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/reusedev/uportal-api/pkg/logs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/reusedev/uportal-api/internal/handler"
	"github.com/reusedev/uportal-api/internal/middleware"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/config"
)

var (
	configPath string
	doMigrate  bool
)

func init() {
	flag.StringVar(&configPath, "config", "config/config.yaml", "config file path")
	flag.BoolVar(&doMigrate, "migrate", false, "执行数据库迁移")
	flag.Parse()
}

func main() {
	// 1. 加载配置
	if err := config.LoadConfig(configPath); err != nil {
		panic(fmt.Sprintf("Load config error: %v", err))
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
		panic(fmt.Sprintf("Init logger error: %v", err))
	}
	defer logs.Sync()

	// 3. 初始化数据库
	if err := model.InitDB(); err != nil {
		logs.Business().Fatal("Init database error", zap.Error(err))
	}
	defer model.CloseDB()

	if doMigrate {
		logs.Business().Info("执行数据库迁移...")
		if err := model.Migrate(model.DB); err != nil {
			logs.Business().Error("数据库迁移失败 ", zap.Error(err))
		}
		log.Println("数据库迁移完成。")
	}

	// 4. 初始化Redis
	if err := model.InitRedis(); err != nil {
		logs.Business().Fatal("Init redis error", zap.Error(err))
	}
	defer model.CloseRedis()

	// 6. 创建Gin引擎
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()

	// 7. 注册中间件
	// 注意：中间件的注册顺序很重要
	engine.Use(middleware.Recovery(logs.Business())) // 恢复中间件应该最先注册
	engine.Use(middleware.Logger(logs.Business()))   // 日志中间件
	engine.Use(middleware.CORS())                    // CORS中间件

	// 8. 注册路由
	registerRoutes(engine, model.DB, cfg)

	// 9. 启动服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 10. 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Business().Fatal("Server error", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logs.Business().Info("Shutting down server...")
}

// registerRoutes 注册路由
func registerRoutes(engine *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// 初始化服务
	adminService := service.NewAdminService(db)
	taskService := service.NewTaskService(db, model.RedisClient, logs.Business(), cfg)
	configService := service.NewSystemConfigService(db)
	tokenRecordService := service.NewTokenRecordService(db)
	loginService := service.NewUserLoginLogService(db)
	tokenService := service.NewTokenService(db)

	// 初始化处理器
	adminHandler := handler.NewAdminHandler(adminService, loginService, tokenRecordService)
	taskHandler := handler.NewTaskHandler(taskService)
	configHandler := handler.NewSystemConfigHandler(configService)
	tokenHandler := handler.NewTokenHandler(tokenService)

	// 注册路由
	api := engine.Group("/admin")
	{
		// 管理员用户
		{
			// 认证相关
			handler.RegisterAdminRoutes(api, adminHandler)
			// 管理员相关
			admin := api.Group("/", middleware.AdminAuth())
			handler.RegisterAdminManagementRoutes(admin, adminHandler)
		}
		// 客户端用户
		{
			user := api.Group("/users", middleware.AdminAuth())
			handler.RegisterUserManagerRoutes(user, adminHandler)
		}

		// 系统配置
		{
			configs := api.Group("/configs", middleware.AdminAuth())
			handler.RegisterSystemConfigRoutes(configs, configHandler)
		}
		// 代币管理
		{
			reward := api.Group("/reward-tasks", middleware.AdminAuth())
			handler.RegisterRewardTaskRoutes(reward, taskHandler)
		}
		// 代币消耗规则
		{
			reward := api.Group("/token-consume-rules", middleware.AdminAuth())
			handler.RegisterTokenConsumeRulesRoutes(reward, taskHandler)
		}
		// 充值方案
		{
			recharge := api.Group("/recharge-plans", middleware.AdminAuth())
			handler.RegisterAdminTokenRoutes(recharge, tokenHandler)
		}
	}
}
