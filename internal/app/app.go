package app

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/logs"
	"gorm.io/gorm"
)

func InitServices(cfg *config.Config, db *gorm.DB, redis *redis.Client) (*service.AuthService, *service.AdminService, *service.TokenService, *service.TaskService, *service.PaymentService, error) {
	// 初始化微信服务
	wechatSvc := service.NewWechatService(cfg)

	// 初始化认证服务
	authSvc := service.NewAuthService(db, wechatSvc)

	// 初始化其他服务
	adminSvc := service.NewAdminService(db)
	tokenSvc := service.NewTokenService(db)
	taskSvc := service.NewTaskService(db, redis, logs.Business(), cfg)
	paymentSvc, err := service.NewPaymentService(db, redis, nil, cfg)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("init payment service error: %v", err)
	}

	return authSvc, adminSvc, tokenSvc, taskSvc, paymentSvc, nil
}
