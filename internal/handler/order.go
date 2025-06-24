package handler

import (
	"github.com/reusedev/uportal-api/pkg/consts"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/model"
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

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	ProductID   string  `json:"product_id" binding:"required"`
	ProductName string  `json:"product_name" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	// 从上下文获取用户ID
	userID := c.GetString(consts.UserId)

	order, err := h.orderService.CreateOrder(c.Request.Context(), userID, req.Amount, req.ProductID, req.ProductName)
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

	// 验证订单所属用户
	userID := c.GetString("user_id")
	if order.UserID != userID {
		response.Error(c, errors.New(errors.ErrCodeForbidden, "无权查看此订单", nil))
		return
	}

	response.Success(c, order)
}

// GetAdminOrder 获取订单详情
func (h *OrderHandler) GetAdminOrder(c *gin.Context) {
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

// ListInfo 获取订单列表（管理员接口）
func (h *OrderHandler) ListInfo(c *gin.Context) {
	var req service.ListInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	orders, err := h.orderService.ListInfo(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, orders)
}

// ListOrders 获取订单列表（管理员接口）
func (h *OrderHandler) ListOrders(c *gin.Context) {
	var req service.ListOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	orders, total, err := h.orderService.ListOrders(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, orders, total)
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

	response.ListResponse(c, orders, total)
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

	userID := c.GetString("user_id")
	if order.UserID != userID {
		response.Error(c, errors.New(errors.ErrCodeForbidden, "无权操作此订单", nil))
		return
	}

	err = h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, model.OrderStatusCancelled, nil)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateOrderStatusRequest 更新订单状态请求
type UpdateOrderStatusRequest struct {
	Status      int8                   `json:"status" binding:"required"`
	PaymentInfo map[string]interface{} `json:"payment_info"`
}

// UpdateOrderStatus 更新订单状态（管理员接口）
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err = h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, req.Status, req.PaymentInfo)
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
	}
}

// RegisterAdminOrderRoutes 注册管理员订单相关路由
func RegisterAdminOrderRoutes(r *gin.RouterGroup, h *OrderHandler) {
	// 获取充值订单列表
	r.POST("/list", h.ListOrders)
	// 获取充值订单状态分类详情
	r.POST("/list_info", h.ListInfo)
	// 获取充值订单详情
	r.GET("/:id/info", h.GetAdminOrder)
}
