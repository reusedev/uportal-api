package handler

import (
	"fmt"
	"net/http"

	"github.com/reusedev/uportal-api/pkg/consts"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/logs"
	"github.com/reusedev/uportal-api/pkg/response"
	"go.uber.org/zap"
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
	var req service.CreateWxPayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logs.Business().Error("[CreateWxPayOrder] 参数绑定失败", zap.Error(err))
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	// 获取支付方案信息
	plan, err := h.paymentService.GetPlan(c.Request.Context(), req.PlanId)
	if err != nil {
		logs.Business().Error("[CreateWxPayOrder] 获取支付方案失败", zap.Error(err), zap.Any("planId", req.PlanId))
		response.Error(c, err)
		return
	}
	userId := c.GetString(consts.UserId)

	// 创建支付订单
	resp, err := h.paymentService.CreateWxPayOrder(c.Request.Context(), userId, plan)
	if err != nil {
		logs.Business().Error("[CreateWxPayOrder] 创建微信支付订单失败", zap.Error(err), zap.String("userId", userId), zap.Any("planId", req.PlanId))
		response.Error(c, err)
		return
	}
	fmt.Printf("%+v", resp)

	response.Success(c, resp)
}

// HandleWxPayNotify 处理微信支付回调
func (h *PaymentHandler) HandleWxPayNotify(c *gin.Context) {
	// 处理回调
	err := h.paymentService.HandleWxPayNotify(c.Request.Context(), c.Request)
	if err != nil {
		logs.Business().Error("[HandleWxPayNotify] 处理微信支付回调失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "处理回调失败"})
		return
	}

	// 返回成功
	c.JSON(200, gin.H{"code": "SUCCESS"})
}

// QueryWxPayOrder 查询微信支付订单
func (h *PaymentHandler) QueryWxPayOrder(c *gin.Context) {
	var req service.QueryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logs.Business().Error("[QueryWxPayOrder] 参数绑定失败", zap.Error(err))
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	// 查询支付订单
	resp, err := h.paymentService.QueryWxPayOrder(c.Request.Context(), req.OrderId)
	if err != nil {
		logs.Business().Error("[QueryWxPayOrder] 查询微信支付订单失败", zap.Error(err), zap.Any("orderId", req.OrderId))
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// CloseWxPayOrder 关闭微信支付订单
func (h *PaymentHandler) CloseWxPayOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		logs.Business().Error("[CloseWxPayOrder] 订单ID为空")
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", nil))
		return
	}

	// 关闭支付订单
	err := h.paymentService.CloseWxPayOrder(c.Request.Context(), orderID)
	if err != nil {
		logs.Business().Error("[CloseWxPayOrder] 关闭微信支付订单失败", zap.Error(err), zap.String("orderId", orderID))
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// CreateAlipayOrder 创建支付宝支付订单
func (h *PaymentHandler) CreateAlipayOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		logs.Business().Error("[CreateAlipayOrder] 订单ID为空")
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", nil))
		return
	}

	// 获取订单信息
	order, err := h.paymentService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		logs.Business().Error("[CreateAlipayOrder] 获取订单信息失败", zap.Error(err), zap.String("orderId", orderID))
		response.Error(c, err)
		return
	}

	// 创建支付订单
	payUrl, err := h.alipayService.CreateAlipayOrder(c.Request.Context(), orderID, "", order.AmountPaid)
	if err != nil {
		logs.Business().Error("[CreateAlipayOrder] 创建支付宝支付订单失败", zap.Error(err), zap.String("orderId", orderID))
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
		logs.Business().Error("[HandleAlipayNotify] 处理支付宝支付回调失败", zap.Error(err), zap.Any("notifyData", notifyData))
		response.Error(c, err)
		return
	}

	// 返回成功
	c.String(200, "success")
}

// QueryAlipayOrder 查询支付宝支付订单
func (h *PaymentHandler) QueryAlipayOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		logs.Business().Error("[QueryAlipayOrder] 订单ID为空")
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", nil))
		return
	}

	// 查询支付订单
	resp, err := h.alipayService.QueryAlipayOrder(c.Request.Context(), orderID)
	if err != nil {
		logs.Business().Error("[QueryAlipayOrder] 查询支付宝支付订单失败", zap.Error(err), zap.String("orderId", orderID))
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// CloseAlipayOrder 关闭支付宝支付订单
func (h *PaymentHandler) CloseAlipayOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		logs.Business().Error("[CloseAlipayOrder] 订单ID为空")
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的订单ID", nil))
		return
	}

	// 关闭支付订单
	err := h.alipayService.CloseAlipayOrder(c.Request.Context(), orderID)
	if err != nil {
		logs.Business().Error("[CloseAlipayOrder] 关闭支付宝支付订单失败", zap.Error(err), zap.String("orderId", orderID))
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterPaymentRoutes 注册支付相关路由
func RegisterPaymentRoutes(r *gin.RouterGroup, h *PaymentHandler, t *TokenHandler) {

	r.POST("create_order", h.CreateWxPayOrder)
	r.POST("/plan_list", t.ListApiRechargePlans)
	r.POST("/result", h.QueryWxPayOrder)

	r.POST("/orders/:id/close", h.CloseWxPayOrder)
	// 支付宝支付
	//alipay := payments.Group("/alipay", authMiddleware)
	//{
	//	alipay.POST("/orders/:id", h.CreateAlipayOrder)
	//	alipay.GET("/orders/:id", h.QueryAlipayOrder)
	//	alipay.POST("/orders/:id/close", h.CloseAlipayOrder)

}

// RegisterPaymentNotifyRoutes 注册支付回调相关路由
func RegisterPaymentNotifyRoutes(r *gin.RouterGroup, h *PaymentHandler) {
	// 支付回调（不需要认证）
	r.POST("/wechat/notify", h.HandleWxPayNotify)
	r.POST("/alipay/notify", h.HandleAlipayNotify)
}
