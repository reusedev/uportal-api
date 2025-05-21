package service

import (
	"context"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"gorm.io/gorm"
)

// OrderService 订单服务
type OrderService struct {
	db *gorm.DB
}

// NewOrderService 创建订单服务
func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{
		db: db,
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID        int64   `json:"user_id" binding:"required"`
	PlanID        int     `json:"plan_id" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=wechat alipay"`
	Amount        float64 `json:"amount" binding:"required,min=0.01"`
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*model.RechargeOrder, error) {
	var order *model.RechargeOrder
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 获取充值套餐信息
		var plan model.RechargePlan
		err := tx.First(&plan, req.PlanID).Error
		if err != nil {
			return errors.New(errors.ErrCodeInvalidParams, "充值套餐不存在", err)
		}

		// 创建订单
		order = &model.RechargeOrder{
			UserID:        req.UserID,
			PlanID:        &req.PlanID,
			TokenAmount:   plan.TokenAmount,
			AmountPaid:    plan.Price,
			PaymentMethod: req.PaymentMethod,
			Status:        0, // 待支付
		}

		return tx.Create(order).Error
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, orderID int64) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	err := s.db.Preload("User").Preload("Plan").First(&order, orderID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "订单不存在", err)
		}
		return nil, errors.New(errors.ErrCodeInternal, "获取订单失败", err)
	}
	return &order, nil
}

// ListOrdersRequest 获取订单列表请求
type ListOrdersRequest struct {
	Page          int        `form:"page" binding:"required,min=1"`
	PageSize      int        `form:"page_size" binding:"required,min=1,max=100"`
	UserID        *int64     `form:"user_id"`
	Status        *int8      `form:"status"`
	PaymentMethod *string    `form:"payment_method"`
	StartTime     *time.Time `form:"start_time"`
	EndTime       *time.Time `form:"end_time"`
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, req *ListOrdersRequest) ([]*model.RechargeOrder, int64, error) {
	query := s.db.Model(&model.RechargeOrder{})

	// 添加查询条件
	if req.UserID != nil && *req.UserID > 0 {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Status != nil && *req.Status >= 0 {
		query = query.Where("status = ?", *req.Status)
	}
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// 获取总数
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取订单总数失败", err)
	}

	// 获取分页数据
	var orders []*model.RechargeOrder
	err = query.Preload("User").Preload("Plan").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Order("created_at DESC").
		Find(&orders).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取订单列表失败", err)
	}

	return orders, total, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status int8) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取订单信息
		var order model.RechargeOrder
		err := tx.First(&order, orderID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New(errors.ErrCodeNotFound, "订单不存在", err)
			}
			return errors.New(errors.ErrCodeInternal, "获取订单失败", err)
		}

		// 检查订单状态是否可以更新
		if order.Status != 0 && order.Status != 1 {
			return errors.New(errors.ErrCodeInvalidParams, "订单状态不允许更新", nil)
		}

		// 更新订单状态
		updates := map[string]interface{}{
			"status": status,
		}
		if status == 1 { // 支付成功
			updates["paid_at"] = gorm.Expr("CURRENT_TIMESTAMP")
		}

		err = tx.Model(&model.RechargeOrder{}).Where("order_id = ?", orderID).Updates(updates).Error
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "更新订单状态失败", err)
		}

		// 如果支付成功，更新用户代币余额
		if status == 1 {
			err = model.UpdateUserTokenBalance(tx, order.UserID, order.TokenAmount)
			if err != nil {
				return errors.New(errors.ErrCodeInternal, "更新用户代币余额失败", err)
			}
		}

		return nil
	})
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, orderID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取订单信息
		var order model.RechargeOrder
		err := tx.First(&order, orderID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New(errors.ErrCodeNotFound, "订单不存在", err)
			}
			return errors.New(errors.ErrCodeInternal, "获取订单失败", err)
		}

		// 检查订单状态是否可以取消
		if order.Status != 0 {
			return errors.New(errors.ErrCodeInvalidParams, "订单状态不允许取消", nil)
		}

		// 更新订单状态为已取消
		err = tx.Model(&model.RechargeOrder{}).Where("order_id = ?", orderID).
			Update("status", 4).Error // 4表示已取消
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "取消订单失败", err)
		}

		return nil
	})
}

// GetUserOrders 获取用户的订单列表
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*model.RechargeOrder, int64, error) {
	query := s.db.Model(&model.RechargeOrder{}).Where("user_id = ?", userID)

	// 获取总数
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取订单总数失败", err)
	}

	// 获取分页数据
	var orders []*model.RechargeOrder
	err = query.Preload("Plan").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取订单列表失败", err)
	}

	return orders, total, nil
}
