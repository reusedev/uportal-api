package service

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/types"
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
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*types.RechargeOrder, error) {
	var order *types.RechargeOrder
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 获取充值套餐信息
		plan, err := model.GetRechargePlan(tx, int64(req.PlanID))
		if err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(errors.ErrCodeNotFound, "充值套餐不存在", nil)
			}
			return errors.New(errors.ErrCodeInternal, "获取充值套餐失败", err)
		}

		// 检查套餐状态
		if plan.Status != 1 {
			return errors.New(errors.ErrCodeInvalidParams, "充值套餐已下架", nil)
		}

		// 创建订单
		order = &types.RechargeOrder{
			UserID:        req.UserID,
			PlanID:        &req.PlanID,
			TokenAmount:   int(plan.TokenAmount),
			AmountPaid:    req.Amount,
			PaymentMethod: req.PaymentMethod,
			Status:        0, // 待支付
		}

		err = tx.Create(order).Error
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "创建订单失败", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, orderID int64) (*types.RechargeOrder, error) {
	var order types.RechargeOrder
	err := s.db.Preload("User").Preload("Plan").First(&order, orderID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(errors.ErrCodeNotFound, "订单不存在", nil)
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
func (s *OrderService) ListOrders(ctx context.Context, req *ListOrdersRequest) ([]*types.RechargeOrder, int64, error) {
	query := s.db.Model(&types.RechargeOrder{})

	// 添加查询条件
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.PaymentMethod != nil {
		query = query.Where("payment_method = ?", *req.PaymentMethod)
	}
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取订单总数失败", err)
	}

	// 获取分页数据
	var orders []*types.RechargeOrder
	err := query.Preload("User").Preload("Plan").
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
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status int8, transactionID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status": status,
		}

		// 如果订单支付成功，更新支付时间和交易号
		if status == 1 {
			now := time.Now()
			updates["paid_at"] = now
			updates["transaction_id"] = transactionID

			// 获取订单信息
			order, err := s.GetOrder(ctx, orderID)
			if err != nil {
				return err
			}

			// 增加用户Token余额
			err = model.AddToken(tx, order.UserID, int64(order.TokenAmount), 1, // 1表示充值
				strconv.FormatInt(orderID, 10), "充值获得Token")
			if err != nil {
				return errors.New(errors.ErrCodeInternal, "增加Token余额失败", err)
			}
		}

		err := tx.Model(&types.RechargeOrder{}).Where("order_id = ?", orderID).Updates(updates).Error
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "更新订单状态失败", err)
		}

		return nil
	})
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, orderID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取订单信息
		var order types.RechargeOrder
		err := tx.First(&order, orderID).Error
		if err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(errors.ErrCodeNotFound, "订单不存在", nil)
			}
			return errors.New(errors.ErrCodeInternal, "获取订单失败", err)
		}

		// 检查订单状态
		if order.Status != 0 {
			return errors.New(errors.ErrCodeInvalidParams, "只能取消待支付的订单", nil)
		}

		// 更新订单状态为已取消
		err = tx.Model(&types.RechargeOrder{}).Where("order_id = ?", orderID).
			Update("status", 4).Error // 4表示已取消
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "取消订单失败", err)
		}

		return nil
	})
}

// GetUserOrders 获取用户的订单列表
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*types.RechargeOrder, int64, error) {
	query := s.db.Model(&types.RechargeOrder{}).Where("user_id = ?", userID)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取订单总数失败", err)
	}

	// 获取分页数据
	var orders []*types.RechargeOrder
	err := query.Preload("Plan").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取订单列表失败", err)
	}

	return orders, total, nil
}
