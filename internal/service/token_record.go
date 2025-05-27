package service

import (
	"context"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"gorm.io/gorm"
)

type TokenRecordService struct {
	db *gorm.DB
}

func NewTokenRecordService(db *gorm.DB) *TokenRecordService {
	return &TokenRecordService{db: db}
}

// ListTokenRecordsRequest 获取代币记录请求
type ListTokenRecordsRequest struct {
	UserID     string     `json:"user_id"`
	ChangeType string     `json:"change_type"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	PageNum    int        `json:"page" binding:"required,min=1"`
	PageSize   int        `json:"limit" binding:"required,min=1,max=100"`
}

// ListTokenRecordsResponse 代币记录列表响应
type ListTokenRecordsResponse struct {
	List  []*model.TokenRecord `json:"list"`
	Total int64                `json:"total"`
}

// ListTokenRecords 获取用户代币记录列表
func (s *TokenRecordService) ListTokenRecords(ctx context.Context, req *ListTokenRecordsRequest) (*ListTokenRecordsResponse, error) {
	// 构建查询
	query := s.db.Model(&model.TokenRecord{})

	// 添加用户ID过滤
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 添加变动类型过滤
	if req.ChangeType != "" {
		query = query.Where("change_type = ?", req.ChangeType)
	}

	// 添加时间范围过滤
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.New(errors.ErrCodeDatabaseError, "获取代币记录总数失败", err)
	}

	// 获取分页数据
	var records []*model.TokenRecord
	offset := (req.PageNum - 1) * req.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&records).Error; err != nil {
		return nil, errors.New(errors.ErrCodeDatabaseError, "获取代币记录列表失败", err)
	}

	return &ListTokenRecordsResponse{
		List:  records,
		Total: total,
	}, nil
}
