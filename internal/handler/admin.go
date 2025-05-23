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
	Limit    int    `form:"limit" binding:"required,min=1,max=100"`
	NickName string `form:"nickname"`
	Email    string `form:"email"`
	Phone    string `form:"phone"`
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
		Limit:    req.Limit,
		NickName: req.NickName,
		Email:    req.Email,
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

// ListAdminUsersRequest 获取管理员列表请求
type ListAdminUsersRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
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
		PageSize: req.PageSize,
		Username: req.Username,
		Role:     req.Role,
		Status:   req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.ListResponse(c, admins, total, req.Page, req.PageSize)
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
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	Password string `json:"password" binding:"omitempty,min=6,max=32"`
	Role     string `json:"role" binding:"omitempty,oneof=admin super_admin"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1"`
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
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, admin)
}

// UpdateAdmin 更新管理员信息
func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的管理员ID", err))
		return
	}

	var req UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err = h.adminService.UpdateAdmin(c.Request.Context(), id, &service.UpdateAdminRequest{
		Password: req.Password,
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的管理员ID", err))
		return
	}

	err = h.adminService.DeleteAdmin(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterUserListRoutes 注册用户列表路由（普通用户可访问）
func RegisterUserListRoutes(r *gin.RouterGroup, h *AdminHandler) {
	users := r.Group("/users")
	{
		users.GET("/list", h.ListUsers)                    // 获取用户列表
		users.GET("/:id", h.GetUser)                       // 获取用户详情
		users.PUT("/:id", h.UpdateUser)                    // 更新用户信息
		users.DELETE("/:id", h.DeleteUser)                 // 删除用户
		users.POST("/:id/reset-password", h.ResetPassword) // 重置用户密码
	}
}

// RegisterAdminManagementRoutes 注册管理员管理路由
func RegisterAdminManagementRoutes(r *gin.RouterGroup, h *AdminHandler) {
	// 管理员管理路由
	admins := r.Group("/auth")
	{
		admins.POST("/login", h.Login)       // 管理员登录
		admins.GET("", h.ListAdminUsers)     // 获取管理员列表
		admins.POST("create", h.CreateAdmin) // 创建管理员
		admins.PUT("/:id", h.UpdateAdmin)    // 更新管理员信息
		admins.DELETE("/:id", h.DeleteAdmin) // 删除管理员
	}
}
