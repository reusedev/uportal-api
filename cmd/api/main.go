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

	"github.com/reusedev/uportal-api/internal/handler"
	"github.com/reusedev/uportal-api/internal/middleware"
	"github.com/reusedev/uportal-api/internal/model"
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

	// 5. 创建Gin引擎
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()

	// 6. 注册中间件
	engine.Use(middleware.Recovery(logs.Business()))
	engine.Use(middleware.Logger(logs.Business()))
	engine.Use(middleware.CORS())
	engine.Any("/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	})

	// 注册路由
	registerRoutes(engine)

	// 7. 启动服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 8. 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Business().Fatal("Server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logs.Business().Info("Shutting down server...")
}

// registerRoutes 注册所有路由
func registerRoutes(engine *gin.Engine) {
	db := model.DB

	pointsHandler := handler.NewPointsHandler(db)
	paymentHandler := handler.NewPaymentHandler(db)
	userHandler := handler.NewUserHandler(db)

	api := engine.Group("/api")
	handler.RegisterPointsRoutes(api.Group("/points"), pointsHandler)
	handler.RegisterPaymentRoutes(api.Group("/payment"), paymentHandler)
	handler.RegisterUserRoutes(api.Group("/user"), userHandler)
}
