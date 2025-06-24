package service

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"gorm.io/gorm"
)

// OrderService 订单服务
type OrderService struct {
	db *gorm.DB
}

// NewOrderService 创建订单服务实例
func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, userID string, amount float64, productID string, productName string) (*model.RechargeOrder, error) {
	// 创建订单
	order := &model.RechargeOrder{
		UserID:     userID,
		AmountPaid: amount,
		Status:     int8(model.OrderStatusPending),
	}

	err := model.CreateOrder(s.db, order)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "创建订单失败", err)
	}

	return order, nil
}

// GetOrder 获取订单信息
func (s *OrderService) GetOrder(ctx context.Context, orderID int64) (*model.RechargeOrder, error) {
	order, err := model.GetOrderByID(s.db, orderID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(errors.ErrCodeNotFound, "订单不存在", nil)
		}
		return nil, errors.New(errors.ErrCodeInternal, "查询订单失败", err)
	}
	return order, nil
}

type ListInfoRequest struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type ListInfoResp struct {
	TotalCount int `json:"total_count"`
	Status     int `json:"status"`
}

func (s *OrderService) ListInfo(ctx context.Context, req *ListInfoRequest) (interface{}, error) {
	var resp []ListInfoResp

	query := s.db.Model(&model.RechargeOrder{})
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	err := query.Select("status, count(*) as total_count").Group("status").Find(&resp).Error
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "查询订单总金额失败", err)
	}

	ret := map[string]int{
		"success_num": 0,
		"failed_num":  0,
		"pending_num": 0,
	}
	for _, i := range resp {
		switch i.Status {
		case 0:
			ret["pending_num"] = i.TotalCount
		case 1:
			ret["success_num"] = i.TotalCount
		case 2:
			ret["failed_num"] = i.TotalCount
		}
	}
	return ret, nil
}

type ListOrdersRequest struct {
	Page      int    `json:"page" binding:"required,min=1"`
	Limit     int    `json:"limit" binding:"required,min=1,max=100"`
	UserID    string `json:"user_id"`
	Status    *int   `json:"status"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, req *ListOrdersRequest) ([]*model.RechargeOrder, int64, error) {
	var orders []*model.RechargeOrder
	var total int64

	query := s.db
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "查询订单总数失败", err)
	}

	offset := (req.Page - 1) * req.Limit
	err = query.Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(req.Limit).
		Find(&orders).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "查询订单列表失败", err)
	}

	return orders, total, nil
}

// GetUserOrders 获取用户订单列表
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*model.RechargeOrder, int64, error) {
	orders, total, err := model.GetUserOrders(s.db, userID, page, pageSize)
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "查询用户订单失败", err)
	}
	return orders, total, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status int8, paymentInfo map[string]interface{}) error {
	// 获取订单信息
	order, err := model.GetOrderByID(s.db, orderID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "订单不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询订单失败", err)
	}

	// 检查订单状态是否可以更新
	if !model.CanUpdateOrderStatus(order.Status, status) {
		return errors.New(errors.ErrCodeInvalidParams, "订单状态不允许更新", nil)
	}

	// 更新订单状态
	updates := map[string]interface{}{
		"status": status,
	}

	// 如果支付成功，记录支付信息
	if status == model.OrderStatusCompleted && paymentInfo != nil {
		paymentInfoJSON, err := json.Marshal(paymentInfo)
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "序列化支付信息失败", err)
		}
		updates["payment_info"] = string(paymentInfoJSON)
		updates["paid_at"] = time.Now()
	}

	err = model.UpdateOrder(s.db, orderID, updates)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "更新订单状态失败", err)
	}

	return nil
}
