package service

import (
	"context"
	"encoding/json"
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
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, amount float64, productID string, productName string) (*model.Order, error) {
	// 创建订单
	order := &model.Order{
		UserID:      userID,
		OrderNo:     model.GenerateOrderNo(),
		Amount:      amount,
		ProductID:   productID,
		ProductName: productName,
		Status:      model.OrderStatusPending,
	}

	err := model.CreateOrder(s.db, order)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "创建订单失败", err)
	}

	return order, nil
}

// GetOrder 获取订单信息
func (s *OrderService) GetOrder(ctx context.Context, orderID int64) (*model.Order, error) {
	order, err := model.GetOrderByID(s.db, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "订单不存在", nil)
		}
		return nil, errors.New(errors.ErrCodeInternal, "查询订单失败", err)
	}
	return order, nil
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, page, pageSize int, userID int64, status string) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64

	query := s.db.Model(&model.Order{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "查询订单总数失败", err)
	}

	offset := (page - 1) * pageSize
	err = query.Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&orders).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "查询订单列表失败", err)
	}

	return orders, total, nil
}

// GetUserOrders 获取用户订单列表
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*model.Order, int64, error) {
	orders, total, err := model.GetUserOrders(s.db, userID, page, pageSize)
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "查询用户订单失败", err)
	}
	return orders, total, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status model.OrderStatus, paymentInfo map[string]interface{}) error {
	// 获取订单信息
	order, err := model.GetOrderByID(s.db, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
	if status == model.OrderStatusPaid && paymentInfo != nil {
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
