package handler

import (
	"github.com/reusedev/uportal-api/pkg/consts"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
)

// TokenHandler Token处理器
type TokenHandler struct {
	tokenService *service.TokenService
}

// NewTokenHandler 创建Token处理器实例
func NewTokenHandler(tokenService *service.TokenService) *TokenHandler {
	return &TokenHandler{tokenService: tokenService}
}

// DeleteConsumptionRule 删除Token消费规则
func (h *TokenHandler) DeleteConsumptionRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的规则ID", err))
		return
	}

	err = h.tokenService.DeleteConsumptionRule(c.Request.Context(), int(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// CreateRechargePlan 创建充值套餐
func (h *TokenHandler) CreateRechargePlan(c *gin.Context) {
	var req service.CreateRechargePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	plan, err := h.tokenService.CreateRechargePlan(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, plan)
}

// UpdateRechargePlan 更新充值套餐
func (h *TokenHandler) UpdateRechargePlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的套餐ID", err))
		return
	}

	var req service.UpdateRechargePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err = h.tokenService.UpdateRechargePlan(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteRechargePlan 删除充值套餐
func (h *TokenHandler) DeleteRechargePlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的套餐ID", err))
		return
	}

	err = h.tokenService.DeleteRechargePlan(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ListRechargePlans 获取充值套餐列表
func (h *TokenHandler) ListRechargePlans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	plans, total, err := h.tokenService.ListRechargePlans(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, plans, total)
}

// GetUserTokenBalance 获取用户Token余额
func (h *TokenHandler) GetUserTokenBalance(c *gin.Context) {
	userID := c.GetString(consts.UserId)
	balance, err := h.tokenService.GetUserTokenBalance(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, balance)
}

// TokenIsBuy 用户余额是否充足
func (h *TokenHandler) TokenIsBuy(c *gin.Context) {
	var req service.TokenIsBuyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	isBuy, err := h.tokenService.TokenIsBuy(c.Request.Context(), req.UserId, req.FeatureCode, req.Num)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, isBuy)
}

// TokenBuy 用户金币消耗
func (h *TokenHandler) TokenBuy(c *gin.Context) {
	var req service.TokenBuyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	num := req.Num
	if req.Type == consts.Return {
		num *= -1
	}
	cost, err := h.tokenService.ConsumeToken(c.Request.Context(), req.UserId, req.FeatureCode, num)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, cost)
}

// GetUserTokenRecords 获取用户Token记录
func (h *TokenHandler) GetUserTokenRecords(c *gin.Context) {
	var req service.ListUserTokenRecords
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	userID := c.GetString(consts.UserId)
	records, err := h.tokenService.GetUserTokenRecords(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, records)
}

// GetRechargeAmount 计算充值金额
func (h *TokenHandler) GetRechargeAmount(c *gin.Context) {
	planID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的套餐ID", err))
		return
	}

	amount, err := h.tokenService.GetRechargeAmount(c.Request.Context(), planID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"amount": amount})
}

// GetConsumptionAmount 获取服务消费Token数量
func (h *TokenHandler) GetConsumptionAmount(c *gin.Context) {
	serviceType := c.Param("service_type")
	if serviceType == "" {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "服务类型不能为空", nil))
		return
	}

	amount, err := h.tokenService.GetConsumptionAmount(c.Request.Context(), serviceType)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"amount": amount})
}

func (h *TokenHandler) ReportPointsReward(c *gin.Context) {
	// 从上下文获取当前用户ID
	userID, exists := c.Get(consts.UserId)
	if !exists {
		response.Error(c, errors.New(errors.ErrCodeUnauthorized, "未登录", nil))
		return
	}

	// 绑定请求参数
	var req ReportPointsRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	// 处理代币奖励
	if err := h.tokenService.ProcessPointsReward(c.Request.Context(), userID.(string), req.Type); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterTokenRoutes 注册 Token 相关路由
func RegisterTokenRoutes(r *gin.RouterGroup, h *TokenHandler) {
	r.GET("/balance", h.GetUserTokenBalance)  // 获取代币余额
	r.POST("/records", h.GetUserTokenRecords) // 获取用户代币明细记录
	r.POST("/reward", h.ReportPointsReward)   // 上报金币奖励
	r.GET("/plans", h.ListRechargePlans)
}

// RegisterCloudRoutes 注册 Cloud 相关路由
func RegisterCloudRoutes(r *gin.RouterGroup, h *TokenHandler) {
	r.POST("/is_buy", h.TokenIsBuy) // 用户余额是否充足
	r.POST("/buy", h.TokenBuy)      // 用户余额是否充足
}

// RegisterAdminTokenRoutes 注册管理员 Token 相关路由
func RegisterAdminTokenRoutes(r *gin.RouterGroup, h *TokenHandler) {
	tokens := r.Group("/tokens")
	{
		tokens.POST("/plans", h.CreateRechargePlan)
		tokens.PUT("/plans/:id", h.UpdateRechargePlan)
		tokens.DELETE("/plans/:id", h.DeleteRechargePlan)
	}
}
