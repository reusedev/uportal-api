package service

import (
	"context"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"gorm.io/gorm"
)

type SystemConfigService struct {
	db *gorm.DB
}

func NewSystemConfigService(db *gorm.DB) *SystemConfigService {
	return &SystemConfigService{db: db}
}

// GetConfigs 获取所有系统配置
func (s *SystemConfigService) GetConfigs(ctx context.Context) ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	if err := s.db.Find(&configs).Error; err != nil {
		return nil, errors.New(errors.ErrCodeDatabaseError, "获取系统配置失败", err)
	}
	return configs, nil
}

// CreateConfigRequest 创建系统配置请求
type CreateConfigRequest struct {
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// CreateConfig 创建系统配置
func (s *SystemConfigService) CreateConfig(ctx context.Context, req *CreateConfigRequest) error {
	config := &model.SystemConfig{
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		Description: &req.Description,
	}

	if err := s.db.Create(config).Error; err != nil {
		return errors.New(errors.ErrCodeDatabaseError, "创建系统配置失败", err)
	}
	return nil
}

// UpdateConfigRequest 更新系统配置请求
type UpdateConfigRequest struct {
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// UpdateConfig 更新系统配置
func (s *SystemConfigService) UpdateConfig(ctx context.Context, req *UpdateConfigRequest) error {
	result := s.db.Model(&model.SystemConfig{}).
		Where("config_key = ?", req.ConfigKey).
		Updates(map[string]interface{}{
			"config_value": req.ConfigValue,
			"description":  req.Description,
		})

	if result.Error != nil {
		return errors.New(errors.ErrCodeDatabaseError, "更新系统配置失败", result.Error)
	}
	return nil
}

// DeleteConfigRequest 删除系统配置请求
type DeleteConfigRequest struct {
	ConfigKey string `json:"config_key" binding:"required"`
}

// DeleteConfig 删除系统配置
func (s *SystemConfigService) DeleteConfig(ctx context.Context, req *DeleteConfigRequest) error {
	result := s.db.Where("config_key = ?", req.ConfigKey).Delete(&model.SystemConfig{})
	if result.Error != nil {
		return errors.New(errors.ErrCodeDatabaseError, "删除系统配置失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.ErrCodeNotFound, "系统配置不存在", nil)
	}
	return nil
}
