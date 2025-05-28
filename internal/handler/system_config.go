package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/response"
)

type SystemConfigHandler struct {
	configService *service.SystemConfigService
}

func NewSystemConfigHandler(configService *service.SystemConfigService) *SystemConfigHandler {
	return &SystemConfigHandler{
		configService: configService,
	}
}

// GetConfigs 获取系统配置列表
func (h *SystemConfigHandler) GetConfigs(c *gin.Context) {
	configs, err := h.configService.GetConfigs(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, configs)
}

// CreateConfig 创建系统配置
func (h *SystemConfigHandler) CreateConfig(c *gin.Context) {
	var req service.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.configService.CreateConfig(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateConfig 更新系统配置
func (h *SystemConfigHandler) UpdateConfig(c *gin.Context) {
	var req service.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.configService.UpdateConfig(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "系统配置更新成功"})
}

// DeleteConfig 删除系统配置
func (h *SystemConfigHandler) DeleteConfig(c *gin.Context) {
	var req service.DeleteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.configService.DeleteConfig(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RegisterSystemConfigRoutes 注册系统配置路由
func RegisterSystemConfigRoutes(r *gin.RouterGroup, h *SystemConfigHandler) {
	{
		r.GET("", h.GetConfigs)           // 获取系统配置列表
		r.POST("/create", h.CreateConfig) // 创建系统配置
		r.POST("/edit", h.UpdateConfig)   // 更新系统配置
		r.POST("/delete", h.DeleteConfig) // 删除系统配置
	}
}
