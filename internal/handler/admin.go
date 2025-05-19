package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// ListUsersRequest 获取用户列表请求
type ListUsersRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
	Username string `form:"username"`
	Email    string `form:"email"`
	Type     string `form:"type"`
	Status   *int   `form:"status"`
}

// ListUsers 获取用户列表
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid request parameters", err))
		return
	}

	users, total, err := h.adminService.ListUsers(c.Request.Context(), &service.ListUsersParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Username: req.Username,
		Email:    req.Email,
		Type:     req.Type,
		Status:   req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, users, total, req.Page, req.PageSize)
}

// GetUser 获取用户详情
func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid user ID", err))
		return
	}

	user, err := h.adminService.GetUser(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Type   string `json:"type" binding:"omitempty,oneof=user admin"`
	Status *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateUser 更新用户信息
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid user ID", err))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid request parameters", err))
		return
	}

	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.adminService.UpdateUser(c.Request.Context(), id, updates); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid user ID", err))
		return
	}

	if err := h.adminService.DeleteUser(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// ResetPassword 重置用户密码
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid user ID", err))
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid request parameters", err))
		return
	}

	if err := h.adminService.ResetPassword(c.Request.Context(), id, req.Password); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterAdminUserRoutes 注册管理员用户管理路由
func RegisterAdminUserRoutes(r *gin.RouterGroup, h *AdminHandler) {
	users := r.Group("/users")
	{
		users.GET("", h.ListUsers)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
		users.POST("/:id/reset-password", h.ResetPassword)
	}
}
