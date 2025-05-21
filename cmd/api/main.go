package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"

	"github.com/reusedev/uportal-api/config"
	"github.com/reusedev/uportal-api/internal/handler"
	"github.com/reusedev/uportal-api/internal/middleware"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/internal/service"
	pkgconfig "github.com/reusedev/uportal-api/pkg/config"
)

var (
	cfg    string
	zapLog *zap.Logger
)

func init() {
	flag.StringVar(&cfg, "config", "config.yaml", "config file")
}

func main() {
	// 1. 加载配置
	if err := config.LoadConfig(cfg); err != nil {
		log.Fatalf("Load config error: %v", err)
	}

	// 2. 初始化日志
	zapLog = initLogger()
	defer zapLog.Sync()

	// 3. 初始化数据库
	if err := model.InitDB(); err != nil {
		zapLog.Fatal("Init database error", zap.Error(err))
	}
	defer model.CloseDB()

	// 4. 初始化Redis
	if err := model.InitRedis(); err != nil {
		zapLog.Fatal("Init redis error", zap.Error(err))
	}
	defer model.CloseRedis()

	// 5. 创建Gin引擎
	gin.SetMode(config.GlobalConfig.Server.Mode)
	engine := gin.New()

	// 6. 注册中间件
	// 注意：中间件的注册顺序很重要
	engine.Use(middleware.Recovery(zapLog)) // 恢复中间件应该最先注册
	engine.Use(middleware.Logger(zapLog))   // 日志中间件
	engine.Use(middleware.CORS())           // CORS中间件

	// 7. 注册路由
	registerRoutes(engine, model.DB)

	// 8. 启动服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler:      engine,
		ReadTimeout:  config.GlobalConfig.Server.ReadTimeout,
		WriteTimeout: config.GlobalConfig.Server.WriteTimeout,
	}

	// 9. 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLog.Fatal("Server error", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLog.Info("Shutting down server...")
}

// initLogger 初始化日志
func initLogger() *zap.Logger {
	// 配置日志编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 配置日志输出
	var core zapcore.Core
	if config.GlobalConfig.Server.Mode == gin.DebugMode {
		// 开发模式：同时输出到控制台和文件
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

		// 创建日志文件
		logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Open log file error: %v", err)
		}

		// 配置多输出
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zapcore.InfoLevel),
		)
	} else {
		// 生产模式：只输出到文件
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Open log file error: %v", err)
		}
		core = zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zapcore.InfoLevel)
	}

	// 创建Logger
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// registerRoutes 注册路由
func registerRoutes(engine *gin.Engine, db *gorm.DB) {
	// 初始化服务
	authService := service.NewAuthService(db)
	adminService := service.NewAdminService(db)
	tokenService := service.NewTokenService(db)
	orderService := service.NewOrderService(db)
	paymentService, err := service.NewPaymentService(db, model.RedisClient, orderService, pkgconfig.Get())
	if err != nil {
		zapLog.Fatal("Init payment service error", zap.Error(err))
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
