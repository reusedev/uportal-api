package service

import (
	"context"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminService 管理员服务
type AdminService struct {
	db *gorm.DB
}

// NewAdminService 创建管理员服务
func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{
		db: db,
	}
}

// ListUsersParams 获取用户列表参数
type ListUsersParams struct {
	Page     int
	PageSize int
	Username string
	Email    string
	Type     string
	Status   *int
}

// ListUsers 获取用户列表
func (s *AdminService) ListUsers(ctx context.Context, params *ListUsersParams) ([]*model.User, int64, error) {
	query := s.db.Model(&model.User{})

	// 添加查询条件
	if params.Username != "" {
		query = query.Where("username LIKE ?", "%"+params.Username+"%")
	}
	if params.Email != "" {
		query = query.Where("email LIKE ?", "%"+params.Email+"%")
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "Failed to count users", err)
	}

	// 获取分页数据
	var users []*model.User
	if err := query.Offset((params.Page - 1) * params.PageSize).
		Limit(params.PageSize).
		Find(&users).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "Failed to get users", err)
	}

	return users, total, nil
}

// GetUser 获取用户详情
func (s *AdminService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	user, err := model.GetUserByID(s.db, id)
	if err != nil {
		return nil, errors.New(errors.ErrCodeNotFound, "User not found", err)
	}
	return user, nil
}

// UpdateUser 更新用户信息
func (s *AdminService) UpdateUser(ctx context.Context, id int64, updates map[string]interface{}) error {
	// 检查用户是否存在
	if _, err := model.GetUserByID(s.db, id); err != nil {
		return errors.New(errors.ErrCodeNotFound, "User not found", err)
	}

	// 如果更新邮箱，检查是否已存在
	if email, ok := updates["email"].(string); ok && email != "" {
		var count int64
		if err := s.db.Model(&model.User{}).
			Where("email = ? AND id != ?", email, id).
			Count(&count).Error; err != nil {
			return errors.New(errors.ErrCodeInternal, "Failed to check email", err)
		}
		if count > 0 {
			return errors.New(errors.ErrCodeInvalidParams, "Email already exists", nil)
		}
	}

	if err := model.UpdateUser(s.db, id, updates); err != nil {
		return errors.New(errors.ErrCodeInternal, "Failed to update user", err)
	}

	return nil
}

// DeleteUser 删除用户
func (s *AdminService) DeleteUser(ctx context.Context, id int64) error {
	// 检查用户是否存在
	if _, err := model.GetUserByID(s.db, id); err != nil {
		return errors.New(errors.ErrCodeNotFound, "User not found", err)
	}

	if err := model.DeleteUser(s.db, id); err != nil {
		return errors.New(errors.ErrCodeInternal, "Failed to delete user", err)
	}

	return nil
}

// ResetPassword 重置用户密码
func (s *AdminService) ResetPassword(ctx context.Context, id int64, password string) error {
	// 检查用户是否存在
	if _, err := model.GetUserByID(s.db, id); err != nil {
		return errors.New(errors.ErrCodeNotFound, "User not found", err)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "Failed to encrypt password", err)
	}

	// 更新密码
	if err := model.UpdateUser(s.db, id, map[string]interface{}{
		"password": string(hashedPassword),
	}); err != nil {
		return errors.New(errors.ErrCodeInternal, "Failed to reset password", err)
	}

	return nil
}
