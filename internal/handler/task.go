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

	task, err := h.taskService.CreateTask(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, task)
}

// UpdateTask 更新任务
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	task, err := h.taskService.UpdateTask(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, task)
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

	tasks, total, err := h.taskService.ListTasks(c.Request.Context(), req.Page, req.Limit, req.Status, req.TaskName)
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

// RegisterRewardTaskRoutes 注册代币任务配置路由
func RegisterRewardTaskRoutes(r *gin.RouterGroup, h *TaskHandler) {
	r.POST("/list", h.ListTasks)
	r.POST("/create", h.CreateTask)
	r.PUT("/edit", h.UpdateTask)
}
