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
	flag.BoolVar(&doMigrate, "migrate", true, "执行数据库迁移")
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
	engine.Use(middleware.CORS())
	engine.Any("/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	}) // CORS中间件

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
	wechatSvc := service.NewWechatService(cfg)
	authService := service.NewAuthService(db, wechatSvc)
	tokenService := service.NewTokenService(db)
	orderService := service.NewOrderService(db)
	inviteService := service.NewInviteService(db)
	taskService := service.NewTaskService(db, model.RedisClient, logs.Business(), cfg)
	paymentService, err := service.NewPaymentService(db, model.RedisClient, orderService, cfg)
	if err != nil {
		logs.Business().Error("Init payment service error", zap.Error(err))
	}
	alipayService, err := service.NewAlipayService(db, orderService, cfg)
	if err != nil {
		logs.Business().Error("Init alipay service error", zap.Error(err))
	}

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	inviteHandler := handler.NewInviteHandler(inviteService)
	tokenHandler := handler.NewTokenHandler(tokenService)
	orderHandler := handler.NewOrderHandler(orderService)
	paymentHandler := handler.NewPaymentHandler(paymentService, alipayService)
	taskHandler := handler.NewTaskHandler(taskService)

	// 注册路由
	api := engine.Group("/api")
	{
		// 登陆
		handler.RegisterUser(api, authHandler)

		// 更新身份认证信息
		user := api.Group("profile", middleware.Auth())
		handler.RegisterUserRoutes(user, authHandler)
		// 邀请
		invite := api.Group("invite", middleware.Auth())
		handler.RegisterInviteRoutes(invite, inviteHandler)

		// 代币相关路由
		token := api.Group("/points", middleware.Auth())
		handler.RegisterTokenRoutes(token, tokenHandler)

		// 订单相关路由
		order := api.Group("/orders", middleware.Auth())
		handler.RegisterOrderRoutes(order, orderHandler, middleware.Auth())

		// 支付相关路由
		handler.RegisterPaymentRoutes(api, paymentHandler, middleware.Auth())

		// 任务相关路由
		tasks := api.Group("/tasks")
		{

			// 用户接口
			userTasks := tasks.Group("", middleware.Auth())
			{
				userTasks.GET("/available", taskHandler.GetAvailableTasks)
				userTasks.POST("/complete", taskHandler.CompleteTask)
				userTasks.GET("/records", taskHandler.GetUserTaskRecords)
				userTasks.GET("/statistics", taskHandler.GetUserTaskStatistics)
			}
		}
	}
}
