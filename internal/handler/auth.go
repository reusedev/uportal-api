package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/consts"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Email    string `json:"email" binding:"required,email"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	// 获取客户端信息
	req.Platform = c.GetHeader("X-Platform")
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	user, token, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// ThirdPartyLogin 第三方登录
func (h *AuthHandler) ThirdPartyLogin(c *gin.Context) {
	var req service.ThirdPartyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	// 获取客户端信息
	req.Platform = c.GetHeader("X-Platform")
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	user, token, err := h.authService.ThirdPartyLogin(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateProfile 更新用户信息
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req service.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	userID := c.GetInt64(consts.UserId)

	updates := make(map[string]interface{})
	if req.Nickname != nil {
		updates["nickname"] = req.Nickname
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = req.AvatarURL
	}

	err := h.authService.UpdateUser(c.Request.Context(), userID, updates)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		OldPassword string `json:"old_password" binding:"required,min=6,max=32"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// WxMiniProgramLogin 微信小程序登录
func (h *AuthHandler) WxMiniProgramLogin(c *gin.Context) {
	var req service.WxMiniProgramLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	// 获取客户端信息
	req.Platform = c.GetHeader("X-Platform")
	req.IP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	user, token, err := h.authService.WxMiniProgramLogin(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// RegisterUser 注册普通用户
func RegisterUser(r *gin.RouterGroup, h *AuthHandler) {
	// 公开路由
	//r.POST("/login", h.Register)
	r.POST("/login", h.WxMiniProgramLogin)          // 微信登陆
	r.POST("/third-party-login", h.ThirdPartyLogin) // 第三方登陆
}

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(r *gin.RouterGroup, h *AuthHandler) {
	//r.GET("/profile", h.GetProfile)
	r.PUT("/update", h.UpdateProfile)
	//r.PUT("/password", h.ChangePassword)
}
