package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/consts"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
)

type NotifyHandler struct {
	notifyService *service.NotifyService
}

// NewNotifyHandler 创建消息通知处理器
func NewNotifyHandler(notifyService *service.NotifyService) *NotifyHandler {
	return &NotifyHandler{
		notifyService: notifyService,
	}
}

func (h *NotifyHandler) Subscribe(c *gin.Context) {
	var req service.SubscribeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	userId := c.GetString(consts.UserId)
	if err := h.notifyService.Notify(c.Request.Context(), &req, userId); err != nil {
		response.Error(c, errors.New(errors.ErrCodeServiceUnavailable, "内部异常", err))
		return
	}
	response.Success(c, nil)
}

func (h *NotifyHandler) Send(c *gin.Context) {
	var req service.SendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	if err := h.notifyService.Send(c.Request.Context(), &req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeServiceUnavailable, "内部异常", err))
		return
	}
	response.Success(c, nil)
}

// RegisterNotifyRoutes 注册消息通知相关路由
func RegisterNotifyRoutes(r *gin.RouterGroup, h *NotifyHandler) {
	// 用户订单接口
	r.POST("/subscribe", h.Subscribe)
}
