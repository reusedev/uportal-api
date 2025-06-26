package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService *service.PaymentService
	alipayService  *service.AlipayService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService *service.PaymentService, alipayService *service.AlipayService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		alipayService:  alipayService,
	}
}

// CreateWxPayOrder 创建微信支付订单
func (h *PaymentHandler) CreateWxPayOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	// 获取订单信息
	order, err := h.paymentService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 创建支付订单
	resp, err := h.paymentService.CreateWxPayOrder(c.Request.Context(), orderID, "", order.AmountPaid)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// HandleWxPayNotify 处理微信支付回调
func (h *PaymentHandler) HandleWxPayNotify(c *gin.Context) {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "读取请求体失败", err))
		return
	}

	// 获取请求头
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 处理回调
	err = h.paymentService.HandleWxPayNotify(c.Request.Context(), body, headers)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 返回成功
	c.String(200, "success")
}

// QueryWxPayOrder 查询微信支付订单
func (h *PaymentHandler) QueryWxPayOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	// 查询支付订单
	resp, err := h.paymentService.QueryWxPayOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// CloseWxPayOrder 关闭微信支付订单
func (h *PaymentHandler) CloseWxPayOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	// 关闭支付订单
	err = h.paymentService.CloseWxPayOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// CreateAlipayOrder 创建支付宝支付订单
func (h *PaymentHandler) CreateAlipayOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	// 获取订单信息
	order, err := h.paymentService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 创建支付订单
	payUrl, err := h.alipayService.CreateAlipayOrder(c.Request.Context(), orderID, "", order.AmountPaid)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"pay_url": payUrl,
	})
}

// HandleAlipayNotify 处理支付宝支付回调
func (h *PaymentHandler) HandleAlipayNotify(c *gin.Context) {
	// 获取所有请求参数
	notifyData := make(map[string]string)
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			notifyData[k] = v[0]
		}
	}

	// 处理回调
	err := h.alipayService.HandleAlipayNotify(c.Request.Context(), notifyData)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 返回成功
	c.String(200, "success")
}

// QueryAlipayOrder 查询支付宝支付订单
func (h *PaymentHandler) QueryAlipayOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	// 查询支付订单
	resp, err := h.alipayService.QueryAlipayOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// CloseAlipayOrder 关闭支付宝支付订单
func (h *PaymentHandler) CloseAlipayOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", err))
		return
	}

	// 关闭支付订单
	err = h.alipayService.CloseAlipayOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterPaymentRoutes 注册支付相关路由
func RegisterPaymentRoutes(r *gin.RouterGroup, h *PaymentHandler, authMiddleware gin.HandlerFunc) {
	payments := r.Group("/payments")
	{
		{
			// 微信支付
			wx := payments.Group("/wechat", authMiddleware)
			{
				wx.POST("/orders/:id", h.CreateWxPayOrder)
				wx.GET("/orders/:id", h.QueryWxPayOrder)
				wx.POST("/orders/:id/close", h.CloseWxPayOrder)
			}

			// 支付宝支付
			alipay := payments.Group("/alipay", authMiddleware)
			{
				alipay.POST("/orders/:id", h.CreateAlipayOrder)
				alipay.GET("/orders/:id", h.QueryAlipayOrder)
				alipay.POST("/orders/:id/close", h.CloseAlipayOrder)
			}
		}

		// 支付回调（不需要认证）
		payments.POST("/wechat/notify", h.HandleWxPayNotify)
		payments.POST("/alipay/notify", h.HandleAlipayNotify)
	}
}
