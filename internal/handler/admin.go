package handler

import (
	"github.com/reusedev/uportal-api/pkg/consts"
	"strconv"
	"time"

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
	Page     int    `json:"page" binding:"required,min=1"`
	Limit    int    `json:"limit" binding:"required,min=1,max=100"`
	NickName string `json:"nickname"`
	Phone    string `json:"phone"`
	Status   *int   `json:"status"`
}

type OperateUsersRequest struct {
	UserId int  `json:"user_id" binding:"required"`
	Status *int `json:"status" binding:"required"`
}

type TokenUsersRequest struct {
	ChangeAmount *int   `json:"change_amount" binding:"required"`
	Remark       string `json:"remark"`
	UserId       string `json:"user_id" binding:"required"`
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
		Limit:    req.Limit,
		NickName: req.NickName,
		Phone:    req.Phone,
		Status:   req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, users, total, req.Page, req.Limit)
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

// TokenAdjustUser 调整用户代币
func (h *AdminHandler) TokenAdjustUser(c *gin.Context) {
	var req TokenUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid request parameters", err))
		return
	}

	updates := make(map[string]interface{})
	if req.ChangeAmount != nil {
		updates["token_balance"] = *req.ChangeAmount
	}
	updates["updated_at"] = time.Now()
	id, _ := strconv.Atoi(req.UserId)

	if err := h.adminService.UpdateUser(c.Request.Context(), int64(id), updates); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateUser 更新用户信息
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	var req OperateUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid request parameters", err))
		return
	}

	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	updates["updated_at"] = time.Now()
	if err := h.adminService.UpdateUser(c.Request.Context(), int64(req.UserId), updates); err != nil {
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
	Password    string `json:"password" binding:"required,min=6,max=32"`
	OldPassword string `json:"old_password" binding:"required,min=6,max=32"`
}

// ResetPassword 重置用户密码
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id := c.GetInt64(consts.UserId)
	if id == 0 {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid user ID", nil))
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid request parameters", err))
		return
	}

	if err := h.adminService.ResetPassword(c.Request.Context(), id, req.Password, req.OldPassword); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ListAdminUsersRequest 获取管理员列表请求
type ListAdminUsersRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	Limit    int    `form:"limit" binding:"required,min=1,max=100"`
	Username string `form:"username"`
	Role     string `form:"role"`
	Status   *int   `form:"status"`
}

// ListAdminUsers 获取管理员列表
func (h *AdminHandler) ListAdminUsers(c *gin.Context) {
	var req ListAdminUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "Invalid request parameters", err))
		return
	}

	admins, total, err := h.adminService.ListAdminUsers(c.Request.Context(), &service.ListAdminUsersParams{
		Page:     req.Page,
		PageSize: req.Limit,
		Username: req.Username,
		Role:     req.Role,
		Status:   req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, admins, total, req.Page, req.Limit)
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Role     string `json:"role" binding:"required,oneof=admin super_admin"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	Id       string `json:"id" binding:"required"`
	UserName string `json:"username" binding:"required,min=3,max=32"`
	Role     string `json:"role" binding:"omitempty,oneof=admin super_admin"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

type DeleteAdminRequest struct {
	Id string `json:"id" binding:"required"`
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	// 获取客户端信息
	loginReq := &service.AdminLoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		Platform:  c.GetHeader("X-Platform"),
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	admin, token, err := h.adminService.Login(c.Request.Context(), loginReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"admin": admin,
		"token": token,
	})
}

// CreateAdmin 创建管理员
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	admin, err := h.adminService.CreateAdmin(c.Request.Context(), &service.CreateAdminRequest{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
		Status:   req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, admin)
}

// UpdateAdmin 更新管理员信息
func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	var req UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err := h.adminService.UpdateAdmin(c.Request.Context(), req.Id, &service.UpdateAdminRequest{
		UserName: req.UserName,
		Role:     req.Role,
		Status:   req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteAdmin 删除管理员
func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	var req DeleteAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err := h.adminService.DeleteAdmin(c.Request.Context(), req.Id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterUserManagerRoutes 注册用户列表路由（普通用户可访问）
func RegisterUserManagerRoutes(r *gin.RouterGroup, h *AdminHandler) {
	users := r.Group("/users")
	{
		users.POST("/list", h.ListUsers)                // 获取用户列表
		users.GET("/:id", h.GetUser)                    // 获取用户详情
		users.POST("/operate", h.UpdateUser)            // 更新用户状态
		users.POST("/tokens/adjust", h.TokenAdjustUser) // 调整用户代币
	}
}

// RegisterAdminManagementRoutes 注册管理员管理路由
func RegisterAdminManagementRoutes(r *gin.RouterGroup, h *AdminHandler) {
	// 管理员管理路由
	r.POST("/auth/change-password", h.ResetPassword) // 获取管理员列表
	r.POST("/managers/list", h.ListAdminUsers)       // 获取管理员列表
	r.DELETE("/managers/edit", h.UpdateAdmin)        // 更新管理员信息
	r.DELETE("/managers/delete", h.DeleteAdmin)      // 删除管理员
}
