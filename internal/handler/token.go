package handler

import (
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

// CreateConsumptionRule 创建Token消费规则
func (h *TokenHandler) CreateConsumptionRule(c *gin.Context) {
	var req service.CreateConsumptionRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	rule, err := h.tokenService.CreateConsumptionRule(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, rule)
}

// UpdateConsumptionRule 更新Token消费规则
func (h *TokenHandler) UpdateConsumptionRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的规则ID", err))
		return
	}

	var req service.UpdateConsumptionRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err = h.tokenService.UpdateConsumptionRule(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteConsumptionRule 删除Token消费规则
func (h *TokenHandler) DeleteConsumptionRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的规则ID", err))
		return
	}

	err = h.tokenService.DeleteConsumptionRule(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ListConsumptionRules 获取Token消费规则列表
func (h *TokenHandler) ListConsumptionRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	rules, total, err := h.tokenService.ListConsumptionRules(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, rules, total, page, pageSize)
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

	response.ListResponse(c, plans, total, page, pageSize)
}

// GetUserTokenBalance 获取用户Token余额
func (h *TokenHandler) GetUserTokenBalance(c *gin.Context) {
	userID := c.GetInt64("user_id")
	balance, err := h.tokenService.GetUserTokenBalance(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"balance": balance})
}

// GetUserTokenRecords 获取用户Token记录
func (h *TokenHandler) GetUserTokenRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	records, total, err := h.tokenService.GetUserTokenRecords(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, records, total, page, pageSize)
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

// RegisterTokenRoutes 注册 Token 相关路由
func RegisterTokenRoutes(r *gin.RouterGroup, h *TokenHandler) {
	tokens := r.Group("/tokens")
	{
		tokens.GET("/balance", h.GetUserTokenBalance)
		tokens.GET("/history", h.GetUserTokenRecords)
		tokens.GET("/plans", h.ListRechargePlans)
	}
}

// RegisterAdminTokenRoutes 注册管理员 Token 相关路由
func RegisterAdminTokenRoutes(r *gin.RouterGroup, h *TokenHandler) {
	tokens := r.Group("/tokens")
	{
		tokens.POST("/plans", h.CreateRechargePlan)
		tokens.PUT("/plans/:id", h.UpdateRechargePlan)
		tokens.DELETE("/plans/:id", h.DeleteRechargePlan)
		tokens.POST("/consumption-rules", h.CreateConsumptionRule)
		tokens.PUT("/consumption-rules/:id", h.UpdateConsumptionRule)
		tokens.DELETE("/consumption-rules/:id", h.DeleteConsumptionRule)
		tokens.GET("/consumption-rules", h.ListConsumptionRules)
	}
}
