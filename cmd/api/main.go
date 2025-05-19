package main

import (
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
)

func main() {
	// 1. 加载配置
	if err := config.LoadConfig(""); err != nil {
		log.Fatalf("Load config error: %v", err)
	}

	// 2. 初始化日志
	logger := initLogger()
	defer logger.Sync()

	// 3. 初始化数据库
	if err := model.InitDB(); err != nil {
		logger.Fatal("Init database error", zap.Error(err))
	}
	defer model.CloseDB()

	// 4. 初始化Redis
	if err := model.InitRedis(); err != nil {
		logger.Fatal("Init redis error", zap.Error(err))
	}
	defer model.CloseRedis()

	// 5. 创建Gin引擎
	gin.SetMode(config.GlobalConfig.Server.Mode)
	engine := gin.New()

	// 6. 注册中间件
	// 注意：中间件的注册顺序很重要
	engine.Use(middleware.Recovery(logger)) // 恢复中间件应该最先注册
	engine.Use(middleware.Logger(logger))   // 日志中间件
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
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
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
func registerRoutes(r *gin.Engine, db *gorm.DB) {
	// 初始化服务
	authService := service.NewAuthService(db)
	adminService := service.NewAdminService(db)
	tokenService := service.NewTokenService(db)
	orderService := service.NewOrderService(db)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	adminHandler := handler.NewAdminHandler(adminService)
	tokenHandler := handler.NewTokenHandler(tokenService)
	orderHandler := handler.NewOrderHandler(orderService)

	// 注册路由
	api := r.Group("/api/v1")
	{
		// 认证相关路由
		handler.RegisterAuthRoutes(api, authHandler)

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(middleware.Auth())
		{
			// 用户相关路由
			handler.RegisterUserRoutes(auth, authHandler)

			// Token相关路由
			handler.RegisterTokenRoutes(auth, tokenHandler)

			// 订单相关路由
			handler.RegisterOrderRoutes(auth, orderHandler, middleware.Auth())

			// 管理员路由
			admin := auth.Group("/admin")
			admin.Use(middleware.AdminAuth())
			{
				// 用户管理
				handler.RegisterAdminUserRoutes(admin, adminHandler)
				// Token管理
				handler.RegisterAdminTokenRoutes(admin, tokenHandler)
			}
		}
	}
}
