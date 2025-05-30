package service

import (
	"context"
	stderrors "errors"
	"strings"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/jwt"
	"github.com/reusedev/uportal-api/pkg/logs"
	"go.uber.org/zap"
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

// SortParam 排序参数
type SortParam struct {
	Field string // 排序字段
	Order string // 排序方向：asc 或 desc
}

// ListUsersParams 获取用户列表参数
type ListUsersParams struct {
	Page     int
	Limit    int
	NickName string
	Phone    string
	Status   *int
	Sort     []SortParam // 排序参数
}

// ListUsers 获取用户列表
func (s *AdminService) ListUsers(ctx context.Context, params *ListUsersParams) ([]*model.User, int64, error) {
	query := s.db.Model(&model.User{})

	// 添加查询条件
	if params.NickName != "" {
		query = query.Where("nickname LIKE ?", "%"+params.NickName+"%")
	}
	if params.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+params.Phone+"%")
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	// 添加排序
	for _, sort := range params.Sort {
		if sort.Order == "desc" {
			query = query.Order(sort.Field + " DESC")
		} else {
			query = query.Order(sort.Field + " ASC")
		}
	}

	// 如果没有排序参数，默认按创建时间倒序
	if len(params.Sort) == 0 {
		query = query.Order("created_at DESC")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取用户总数失败", err)
	}

	// 获取分页数据
	var users []*model.User
	if err := query.Offset((params.Page - 1) * params.Limit).
		Limit(params.Limit).
		Find(&users).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取用户列表失败", err)
	}

	return users, total, nil
}

// GetUser 获取用户详情
func (s *AdminService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := s.db.Preload("UserAuths").First(&user, id).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(errors.ErrCodeNotFound, "User not found", err)
		}
		return nil, errors.New(errors.ErrCodeInternal, "Failed to get user", err)
	}
	return &user, nil
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
func (s *AdminService) ResetPassword(ctx context.Context, id int64, password, oldPassword string) error {
	// 检查用户是否存在
	adminUser, err := model.GetAdinUserByID(s.db, id)
	if err != nil {
		return errors.New(errors.ErrCodeNotFound, "User not found", err)
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(adminUser.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New(errors.ErrCodeUnauthorized, "用户名或密码错误", err)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "Failed to encrypt password", err)
	}

	// 更新密码
	if err = model.UpdateAdminUser(s.db, id, map[string]interface{}{
		"password_hash": string(hashedPassword),
	}); err != nil {
		return errors.New(errors.ErrCodeInternal, "Failed to reset password", err)
	}

	return nil
}

// ListAdminUsersParams 获取管理员列表参数
type ListAdminUsersParams struct {
	Page     int
	PageSize int
	Username string
	Role     string
	Status   *int
}

// ListAdminUsers 获取管理员列表
func (s *AdminService) ListAdminUsers(ctx context.Context, params *ListAdminUsersParams) ([]*model.AdminUser, int64, error) {
	query := s.db.Model(&model.AdminUser{})

	// 添加查询条件
	if params.Username != "" {
		query = query.Where("username LIKE ?", "%"+params.Username+"%")
	}
	if params.Role != "" {
		query = query.Where("role = ?", params.Role)
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "Failed to count admin users", err)
	}

	// 获取分页数据
	var admins []*model.AdminUser
	if err := query.Offset((params.Page - 1) * params.PageSize).
		Limit(params.PageSize).
		Order("created_at DESC").
		Find(&admins).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "Failed to get admin users", err)
	}

	return admins, total, nil
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Platform  string `json:"-"` // 登录平台，从请求头获取
	IP        string `json:"-"` // 登录IP，从请求头获取
	UserAgent string `json:"-"` // 设备信息，从请求头获取
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
	UserName string `json:"username" binding:"required,min=3,max=32"`
	Role     string `json:"role" binding:"omitempty,oneof=admin super_admin"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// Login 管理员登录
func (s *AdminService) Login(ctx context.Context, req *AdminLoginRequest) (*model.AdminUser, string, error) {
	var admin model.AdminUser
	err := s.db.Where("username = ?", req.Username).First(&admin).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New(errors.ErrCodeUnauthorized, "用户名或密码错误", nil)
		}
		return nil, "", errors.New(errors.ErrCodeInternal, "查询管理员失败", err)
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, "", errors.New(errors.ErrCodeUnauthorized, "用户名或密码错误", nil)
	}

	// 检查状态
	if admin.Status != 1 {
		return nil, "", errors.New(errors.ErrCodeForbidden, "账号已被禁用", nil)
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(int64(admin.AdminID), strings.Contains(admin.Role, "super"),
		admin.Username, req.Password, admin.Role)
	if err != nil {
		return nil, "", errors.New(errors.ErrCodeInternal, "生成token失败", err)
	}

	// 更新最后登录时间
	now := time.Now()
	if err := s.db.Model(&admin).Update("last_login_at", now).Error; err != nil {
		// 仅记录错误，不影响登录流程
		logs.Business().Warn("更新管理员最后登录时间失败",
			zap.Int("admin_id", admin.AdminID),
			zap.Error(err),
		)
	}

	return &admin, token, nil
}

// CreateAdmin 创建管理员
func (s *AdminService) CreateAdmin(ctx context.Context, req *CreateAdminRequest) (*model.AdminUser, error) {
	// 检查用户名是否已存在
	var count int64
	if err := s.db.Model(&model.AdminUser{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "检查用户名失败", err)
	}
	if count > 0 {
		return nil, errors.New(errors.ErrCodeInvalidParams, "用户名已存在", nil)
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "生成密码哈希失败", err)
	}

	// 创建管理员
	admin := &model.AdminUser{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Role:         req.Role,
		Status:       *req.Status,
	}

	if err := s.db.Create(admin).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "创建管理员失败", err)
	}

	return admin, nil
}

// UpdateAdmin 更新管理员信息
func (s *AdminService) UpdateAdmin(ctx context.Context, id string, req *UpdateAdminRequest) error {
	// 检查管理员是否存在
	var admin model.AdminUser
	if err := s.db.First(&admin, id).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "管理员不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询管理员失败", err)
	}

	// 不允许修改超级管理员的角色
	//if admin.Role == "super_admin" && req.Role != "" && req.Role != "super_admin" {
	//	return errors.New(errors.ErrCodeForbidden, "不能修改超级管理员的角色", nil)
	//}

	updates := make(map[string]interface{})
	if req.UserName != "" {
		updates["username"] = req.UserName
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := s.db.Model(&admin).Updates(updates).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "更新管理员失败", err)
	}

	return nil
}

// DeleteAdmin 删除管理员
func (s *AdminService) DeleteAdmin(ctx context.Context, id string) error {
	// 检查管理员是否存在
	var admin model.AdminUser
	if err := s.db.First(&admin, id).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "管理员不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询管理员失败", err)
	}

	// 不允许删除超级管理员
	if admin.Role == "super_admin" {
		return errors.New(errors.ErrCodeForbidden, "不能删除超级管理员", nil)
	}

	if err := s.db.Delete(&admin).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "删除管理员失败", err)
	}

	return nil
}
