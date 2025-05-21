package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/reusedev/uportal-api/internal/handler"
	"github.com/reusedev/uportal-api/internal/middleware"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/logging"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "config/config.yaml", "config file path")
}

func main() {
	// 1. 解析命令行参数
	flag.Parse()

	// 2. 加载配置
	if err := config.LoadConfig(configPath); err != nil {
		panic(fmt.Sprintf("Load config error: %v", err))
	}
	cfg := config.Get()

	// 3. 初始化日志
	if err := logging.Init(&logging.Config{
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
	defer logging.Sync()

	// 4. 初始化数据库
	if err := model.InitDB(); err != nil {
		logging.Business().Fatal("Init database error", zap.Error(err))
	}
	defer model.CloseDB()

	// 4.1 执行数据库迁移
	if err := model.Migrate(model.DB); err != nil {
		logging.Business().Fatal("Database migration error", zap.Error(err))
	}

	// 5. 初始化Redis
	if err := model.InitRedis(); err != nil {
		logging.Business().Fatal("Init redis error", zap.Error(err))
	}
	defer model.CloseRedis()

	// 6. 创建Gin引擎
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()

	// 7. 注册中间件
	// 注意：中间件的注册顺序很重要
	engine.Use(middleware.Recovery(logging.Business())) // 恢复中间件应该最先注册
	engine.Use(middleware.Logger(logging.Business()))   // 日志中间件
	engine.Use(middleware.CORS())                       // CORS中间件

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
			logging.Business().Fatal("Server error", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Business().Info("Shutting down server...")
}

// registerRoutes 注册路由
func registerRoutes(engine *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// 初始化服务
	authService := service.NewAuthService(db)
	adminService := service.NewAdminService(db)
	tokenService := service.NewTokenService(db)
	orderService := service.NewOrderService(db)
	paymentService, err := service.NewPaymentService(db, model.RedisClient, orderService, cfg)
	if err != nil {
		logging.Business().Fatal("Init payment service error", zap.Error(err))
	}

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	adminHandler := handler.NewAdminHandler(adminService)
	tokenHandler := handler.NewTokenHandler(tokenService)
	orderHandler := handler.NewOrderHandler(orderService)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// 注册路由
	api := engine.Group("/api/v1")
	{
		// 认证相关路由
		handler.RegisterAuthRoutes(api, authHandler)

		// 管理员相关路由
		admin := api.Group("/admin", middleware.AuthMiddleware())
		handler.RegisterAdminUserRoutes(admin, adminHandler)

		// Token相关路由
		token := api.Group("/token", middleware.Auth())
		handler.RegisterTokenRoutes(token, tokenHandler)

		// 订单相关路由
		order := api.Group("/orders", middleware.Auth())
		handler.RegisterOrderRoutes(order, orderHandler, middleware.Auth())

		// 支付相关路由
		handler.RegisterPaymentRoutes(api, paymentHandler, middleware.Auth())
	}
}
