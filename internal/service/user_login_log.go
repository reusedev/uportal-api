package service

import (
	"context"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"gorm.io/gorm"
)

type UserLoginLogService struct {
	db *gorm.DB
}

func NewUserLoginLogService(db *gorm.DB) *UserLoginLogService {
	return &UserLoginLogService{db: db}
}

// ListUserLoginLogsRequest 获取登录日志请求
type ListUserLoginLogsRequest struct {
	UserID    string     `form:"user_id" binding:"required"`
	StartTime *time.Time `form:"start_time" binding:"required"`
	EndTime   *time.Time `form:"end_time" binding:"required"`
	PageNum   int        `form:"page_num" binding:"required,min=1"`
	PageSize  int        `form:"page_size" binding:"required,min=1,max=100"`
}

// ListUserLoginLogsResponse 登录日志列表响应
type ListUserLoginLogsResponse struct {
	List  []*model.UserLoginLog `json:"data"`
	Total int64                 `json:"count"`
}

// ListUserLoginLogs 获取用户登录日志列表
func (s *UserLoginLogService) ListUserLoginLogs(ctx context.Context, req *ListUserLoginLogsRequest) (*ListUserLoginLogsResponse, error) {
	// 构建查询
	query := s.db.Model(&model.UserLoginLog{})

	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.StartTime != nil {
		query = query.Where("start_time >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("end_time <= ?", req.EndTime)
	}
	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.New(errors.ErrCodeDatabaseError, "获取登录日志总数失败", err)
	}

	// 获取分页数据
	var logs []*model.UserLoginLog
	offset := (req.PageNum - 1) * req.PageSize
	if err := query.Order("login_time DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&logs).Error; err != nil {
		return nil, errors.New(errors.ErrCodeDatabaseError, "获取登录日志列表失败", err)
	}

	return &ListUserLoginLogsResponse{
		List:  logs,
		Total: total,
	}, nil
}
