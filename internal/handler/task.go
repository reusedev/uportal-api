package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
	"github.com/reusedev/uportal-api/pkg/utils"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	_, err := h.taskService.CreateTask(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateTask 更新任务
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	_, err := h.taskService.UpdateTask(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID, err := utils.GetIntParam(c, "task_id")
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的任务ID", err))
		return
	}

	if err := h.taskService.DeleteTask(c.Request.Context(), taskID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID, err := utils.GetIntParam(c, "task_id")
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的任务ID", err))
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, task)
}

// ListTasks 获取任务列表
func (h *TaskHandler) ListTasks(c *gin.Context) {
	var req service.ListTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	tasks, total, err := h.taskService.ListTasks(c.Request.Context(), req.Status, req.TaskName)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, tasks, total)
}

// GetAvailableTasks 获取用户可用的任务列表
func (h *TaskHandler) GetAvailableTasks(c *gin.Context) {
	userID := c.GetInt64("user_id")
	tasks, err := h.taskService.GetAvailableTasks(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, tasks)
}

// CompleteTask 完成任务
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req service.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	result, err := h.taskService.CompleteTask(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetUserTaskRecords 获取用户任务完成记录
func (h *TaskHandler) GetUserTaskRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)

	records, total, err := h.taskService.GetUserTaskRecords(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"list":  records,
		"total": total,
	})
}

// GetTaskStatistics 获取任务统计信息
func (h *TaskHandler) GetTaskStatistics(c *gin.Context) {
	taskID, err := utils.GetIntParam(c, "task_id")
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的任务ID", err))
		return
	}

	stats, err := h.taskService.GetTaskStatistics(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, stats)
}

// GetUserTaskStatistics 获取用户任务统计信息
func (h *TaskHandler) GetUserTaskStatistics(c *gin.Context) {
	userID := c.GetInt64("user_id")
	stats, err := h.taskService.GetUserTaskStatistics(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, stats)
}

// ListConsumptionRulesRequest 获取代币消耗规则列表请求
type ListConsumptionRulesRequest struct {
	//Page     int  `json:"page" binding:"required,min=1"`
	//PageSize int  `json:"limit" binding:"required,min=1,max=100"`
	Status *int `json:"status,omitempty"` // 可选的状态过滤
}

// ListConsumptionRules 获取代币消耗规则列表
func (h *TaskHandler) ListConsumptionRules(c *gin.Context) {
	var req ListConsumptionRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	rules, total, err := h.taskService.ListConsumptionRules(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, rules, total)
}

// CreateConsumptionRuleRequest 创建代币消耗规则请求
type CreateConsumptionRuleRequest struct {
	FeatureName string `json:"feature_name"`
	FeatureDesc string `json:"feature_desc"`
	TokenCost   *int   `json:"token_cost,omitempty"`
	FeatureCode string `json:"feature_code"`
	Status      *int8  `json:"status"`
}

// CreateConsumptionRule 创建代币消耗规则
func (h *TaskHandler) CreateConsumptionRule(c *gin.Context) {
	var req CreateConsumptionRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	_, err := h.taskService.CreateConsumptionRule(c.Request.Context(), &service.CreateConsumptionRuleRequest{
		FeatureName: req.FeatureName,
		FeatureDesc: req.FeatureDesc,
		TokenCost:   req.TokenCost,
		FeatureCode: req.FeatureCode,
		Status:      req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateConsumptionRuleRequest 更新代币消耗规则请求
type UpdateConsumptionRuleRequest struct {
	FeatureId   int    `json:"feature_id" binding:"required,min=1"`
	FeatureName string `json:"feature_name"`
	FeatureDesc string `json:"feature_desc"`
	TokenCost   *int64 `json:"token_cost,omitempty"`
	FeatureCode string `json:"feature_code"`
	Status      *int8  `json:"status"`
}

// UpdateConsumptionRule 更新代币消耗规则
func (h *TaskHandler) UpdateConsumptionRule(c *gin.Context) {
	var req UpdateConsumptionRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err := h.taskService.UpdateConsumptionRule(c.Request.Context(), req.FeatureId, &service.UpdateConsumptionRuleRequest{
		FeatureName: req.FeatureName,
		FeatureDesc: req.FeatureDesc,
		TokenCost:   req.TokenCost,
		FeatureCode: req.FeatureCode,
		Status:      req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterRewardTaskRoutes 注册代币任务配置路由
func RegisterRewardTaskRoutes(r *gin.RouterGroup, h *TaskHandler) {
	r.POST("/list", h.ListTasks)
	r.POST("/create", h.CreateTask)
	r.POST("/edit", h.UpdateTask)

	r.POST("/consumption-rules/list", h.ListConsumptionRules)    // 获取代币消耗规则列表
	r.POST("/consumption-rules/create", h.CreateConsumptionRule) // 创建代币消耗规则
	r.POST("/consumption-rules/update", h.UpdateConsumptionRule) // 更新代币消耗规则
}

// RegisterTokenConsumeRulesRoutes 代币消耗规则
func RegisterTokenConsumeRulesRoutes(r *gin.RouterGroup, h *TaskHandler) {
	r.POST("/list", h.ListConsumptionRules)    // 获取代币消耗规则列表
	r.POST("/create", h.CreateConsumptionRule) // 创建代币消耗规则
	r.POST("/update", h.UpdateConsumptionRule) // 更新代币消耗规则
}
