package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
)

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	// 从上下文获取用户ID
	req.UserID = c.GetInt64("user_id")

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, order)
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, order)
}

// ListOrders 获取订单列表（管理员接口）
func (h *OrderHandler) ListOrders(c *gin.Context) {
	var req service.ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	orders, total, err := h.orderService.ListOrders(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, orders, total, req.Page, req.PageSize)
}

// GetUserOrders 获取用户订单列表
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID := c.GetInt64("user_id")

	orders, total, err := h.orderService.GetUserOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, orders, total, page, pageSize)
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	// 验证订单所属用户
	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	userID := c.GetInt64("user_id")
	if order.UserID != userID {
		response.Error(c, errors.New(errors.ErrCodeForbidden, "无权操作此订单", nil))
		return
	}

	err = h.orderService.CancelOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateOrderStatus 更新订单状态（管理员接口）
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	var req struct {
		Status        int8   `json:"status" binding:"required,oneof=0 1 2 3 4"`
		TransactionID string `json:"transaction_id" binding:"required_if=status 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err = h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, req.Status)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterOrderRoutes 注册订单相关路由
func RegisterOrderRoutes(r *gin.RouterGroup, h *OrderHandler, authMiddleware gin.HandlerFunc) {
	orders := r.Group("/orders")
	orders.Use(authMiddleware)
	{
		// 用户订单接口
		orders.POST("", h.CreateOrder)
		orders.GET("/:id", h.GetOrder)
		orders.GET("", h.GetUserOrders)
		orders.POST("/:id/cancel", h.CancelOrder)

		// 管理员订单接口
		admin := orders.Group("/admin")
		{
			admin.GET("", h.ListOrders)
			admin.PUT("/:id/status", h.UpdateOrderStatus)
		}
	}
}
