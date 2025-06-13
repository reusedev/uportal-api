package service

import (
	"context"
	stderrors "errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/logs"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

// AuthService 认证服务
type AuthService struct {
	db        *gorm.DB
	wechatSvc *WechatService
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, wechatSvc *WechatService) *AuthService {
	return &AuthService{
		db:        db,
		wechatSvc: wechatSvc,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"omitempty,len=11"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone     string `json:"phone" binding:"omitempty,len=11"`
	Email     string `json:"email" binding:"omitempty,email"`
	Password  string `json:"password" binding:"required,min=6,max=32"`
	Platform  string `json:"-"` // 登录平台，从请求头获取
	IP        string `json:"-"` // 登录IP，从请求头获取
	UserAgent string `json:"-"` // 设备信息，从请求头获取
}

// ThirdPartyLoginRequest 第三方登录请求
type ThirdPartyLoginRequest struct {
	Provider       string  `json:"provider" binding:"required,oneof=wechat apple google twitter"`
	ProviderUserID string  `json:"provider_user_id" binding:"required"`
	Nickname       *string `json:"nickname" binding:"required,min=2,max=50"`
	AvatarURL      *string `json:"avatar_url" binding:"omitempty,url"`
	Platform       string  `json:"-"` // 登录平台，从请求头获取
	IP             string  `json:"-"` // 登录IP，从请求头获取
	UserAgent      string  `json:"-"` // 设备信息，从请求头获取
}

type UpdateProfileReq struct {
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar"`
}

// WxMiniProgramLoginRequest 微信小程序登录请求
type WxMiniProgramLoginRequest struct {
	Code          string  `json:"code" binding:"required"`
	Nickname      *string `json:"nickname"`
	AvatarURL     *string `json:"avatar_url"`
	EncryptedData string  `json:"encrypted_data"`
	IV            string  `json:"iv"`
	Platform      string  `json:"-"`
	IP            string  `json:"-"`
	UserAgent     string  `json:"-"`
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// 检查手机号或邮箱是否已存在
	if req.Phone != "" {
		exists, err := s.checkPhoneExists(ctx, req.Phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.ErrPhoneExists
		}
	}

	if req.Email != "" {
		exists, err := s.checkEmailExists(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.ErrEmailExists
		}
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "生成密码哈希失败", err)
	}
	passwordHashStr := string(passwordHash)

	// 创建用户
	user := &model.User{
		Phone:        &req.Phone,
		Email:        &req.Email,
		PasswordHash: &passwordHashStr,
		Nickname:     &req.Nickname,
		Status:       1,
	}

	err = model.CreateUser(s.db, user)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "创建用户失败", err)
	}

	return user, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*model.User, string, error) {
	var user *model.User
	var err error

	// 根据手机号或邮箱查找用户
	if req.Phone != "" {
		user, err = model.GetUserByPhone(s.db, req.Phone)
	} else if req.Email != "" {
		user, err = model.GetUserByEmail(s.db, req.Email)
	} else {
		return nil, "", errors.New(errors.ErrCodeInvalidParams, "手机号或邮箱至少提供一个", nil)
	}

	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New(errors.ErrCodeNotFound, "用户不存在", nil)
		}
		return nil, "", errors.New(errors.ErrCodeInternal, "查询用户失败", err)
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, "", errors.New(errors.ErrCodeUnauthorized, "密码错误", nil)
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, "", errors.New(errors.ErrCodeForbidden, "账号已被禁用", nil)
	}

	// 生成JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", errors.New(errors.ErrCodeInternal, "生成token失败", err)
	}

	// 更新最后登录时间
	err = model.UpdateLastLoginTime(s.db, user.UserID)
	if err != nil {
		// 仅记录错误，不影响登录流程
		logs.Business().Warn("更新最后登录时间失败",
			zap.String("user_id", user.UserID),
			zap.Error(err),
		)
	}

	// 记录登录日志
	logEntry := &model.UserLoginLog{
		UserID:        user.UserID,
		LoginMethod:   "password",
		LoginPlatform: &req.Platform,
		IPAddress:     &req.IP,
		DeviceInfo:    &req.UserAgent,
	}
	if err := model.CreateLoginLog(s.db, logEntry); err != nil {
		// 仅记录错误，不影响登录流程
		logs.Business().Warn("创建登录日志失败",
			zap.String("user_id", user.UserID),
			zap.Error(err),
		)
	}

	return user, token, nil
}

// ThirdPartyLogin 第三方登录
func (s *AuthService) ThirdPartyLogin(ctx context.Context, req *ThirdPartyLoginRequest) (*model.User, string, error) {
	var user *model.User
	var token string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 查找是否已存在该第三方账号关联
		existingUser, err := model.GetUserByProvider(tx, req.Provider, req.ProviderUserID)
		if err == nil {
			// 已存在关联，直接登录
			if existingUser.Status != 1 {
				return errors.New(errors.ErrCodeForbidden, "账号已被禁用，有问题请联系客服！", nil)
			}

			// 生成token
			token, err = s.generateToken(existingUser)
			if err != nil {
				return errors.New(errors.ErrCodeInternal, "生成token失败", err)
			}

			// 更新用户信息
			updates := map[string]interface{}{
				"last_login_at": time.Now(),
				"updated_at":    time.Now(),
			}
			if req.Nickname != nil {
				updates["nickname"] = *req.Nickname
			}
			if req.AvatarURL != nil {
				updates["avatar_url"] = *req.AvatarURL
			}

			if err := model.UpdateUser(tx, existingUser.UserID, updates); err != nil {
				return errors.New(errors.ErrCodeInternal, "更新用户信息失败", err)
			}

			user = existingUser
			return nil
		}

		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			logs.DB().Error("查询用户失败",
				zap.Error(err))
			return errors.New(errors.ErrCodeInternal, "查询用户失败", err)
		}
		now := time.Now()
		// 不存在关联，创建新用户
		user = &model.User{
			TokenBalance: 1000,
			Status:       1,
			UserID:       model.GenerateUserID(),
			LastLoginAt:  &now,
		}
		logs.Business().Warn("创建登录日志失败",
			zap.String("user_id", user.UserID),
			zap.Error(err),
		)
		logs.Business().Warn("生成用户ID", zap.String("user_id", user.UserID))
		if req.Nickname != nil {
			user.Nickname = req.Nickname
		}
		if req.AvatarURL != nil {
			user.AvatarURL = req.AvatarURL
		}

		err = tx.Create(user).Error
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "创建用户失败", err)
		}

		// 创建第三方认证关联
		auth := &model.UserAuth{
			UserID:         user.UserID,
			Provider:       req.Provider,
			ProviderUserID: req.ProviderUserID,
		}
		if err := model.CreateUserAuth(tx, auth); err != nil {
			return errors.New(errors.ErrCodeInternal, "创建第三方认证失败", err)
		}
		remark := "注册赠送"
		record := &model.TokenRecord{
			UserID:       user.UserID,
			ChangeAmount: 1000,
			BalanceAfter: 1000,
			ChangeType:   "CONSUME",
			Remark:       &remark,
			ChangeTime:   time.Now(),
		}

		model.CreateTokenRecord(tx, record)

		// 生成token
		token, err = s.generateToken(user)
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "生成token失败", err)
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	// 记录登录日志
	logEntry := &model.UserLoginLog{
		UserID:        user.UserID,
		LoginMethod:   req.Provider,
		LoginPlatform: &req.Platform,
		IPAddress:     &req.IP,
		DeviceInfo:    &req.UserAgent,
	}
	if err := model.CreateLoginLog(s.db, logEntry); err != nil {
		// 仅记录错误，不影响登录流程
		logs.Business().Warn("创建登录日志失败",
			zap.String("user_id", user.UserID),
			zap.Error(err),
		)
	}

	return user, token, nil
}

// 检查手机号是否存在
func (s *AuthService) checkPhoneExists(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := s.db.Model(&model.User{}).Where("phone = ?", phone).Count(&count).Error
	if err != nil {
		return false, errors.New(errors.ErrCodeInternal, "查询手机号失败", err)
	}
	return count > 0, nil
}

// 检查邮箱是否存在
func (s *AuthService) checkEmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := s.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, errors.New(errors.ErrCodeInternal, "查询邮箱失败", err)
	}
	return count > 0, nil
}

// 生成JWT token
func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.UserID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GlobalConfig.JWT.Secret)) // TODO: 从配置中读取密钥
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := model.GetUserByID(s.db, id)
	if err != nil {
		return nil, errors.New(errors.ErrCodeNotFound, "User not found", err)
	}
	return user, nil
}

// UpdateUser 更新用户信息
func (s *AuthService) UpdateUser(ctx context.Context, id string, updates map[string]interface{}) error {
	err := model.UpdateUser(s.db, id, updates)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "更新用户信息失败", err)
	}
	return nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	// 获取用户信息
	user, err := model.GetUserByID(s.db, userID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "用户不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询用户失败", err)
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New(errors.ErrCodeUnauthorized, "旧密码错误", nil)
	}

	// 生成新密码哈希
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "生成密码哈希失败", err)
	}
	passwordHashStr := string(newPasswordHash)

	// 更新密码
	err = model.UpdateUser(s.db, userID, map[string]interface{}{
		"password_hash": passwordHashStr,
	})
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "更新密码失败", err)
	}

	return nil
}

// WxMiniProgramLogin 微信小程序登录
func (s *AuthService) WxMiniProgramLogin(ctx context.Context, req *WxMiniProgramLoginRequest) (*model.User, string, error) {
	// 调用微信服务获取 openid 和 session_key
	wxLoginReq := &WxLoginRequest{
		Code:          req.Code,
		Nickname:      req.Nickname,
		AvatarURL:     req.AvatarURL,
		EncryptedData: req.EncryptedData,
		IV:            req.IV,
	}

	wxResult, err := s.wechatSvc.Login(ctx, wxLoginReq)
	if err != nil {
		return nil, "", err
	}
	// 使用 openid 作为 provider_user_id 进行第三方登录
	thirdPartyReq := &ThirdPartyLoginRequest{
		Provider:       "wechat",
		ProviderUserID: wxResult.OpenID,
		Nickname:       req.Nickname,
		AvatarURL:      req.AvatarURL,
		Platform:       req.Platform,
		IP:             req.IP,
		UserAgent:      req.UserAgent,
	}

	// 调用第三方登录方法
	user, token, err := s.ThirdPartyLogin(ctx, thirdPartyReq)
	if err != nil {
		return nil, "", err
	}

	// 如果有加密数据，解密并更新用户信息
	if req.EncryptedData != "" && req.IV != "" {
		userInfo, err := s.wechatSvc.DecryptUserInfo(wxResult.SessionKey, req.EncryptedData, req.IV)
		if err != nil {
			logs.Business().Warn("解密用户信息失败",
				zap.String("user_id", user.UserID),
				zap.Error(err),
			)
		} else {
			// 更新用户信息
			updates := make(map[string]interface{})
			if nickname, ok := userInfo["nickName"].(string); ok && nickname != "" {
				updates["nickname"] = nickname
			}
			if avatarURL, ok := userInfo["avatarUrl"].(string); ok && avatarURL != "" {
				updates["avatar_url"] = avatarURL
			}
			if len(updates) > 0 {
				if err := model.UpdateUser(s.db, user.UserID, updates); err != nil {
					logs.Business().Warn("更新用户信息失败",
						zap.String("user_id", user.UserID),
						zap.Error(err),
					)
				}
			}
		}
	}

	return user, token, nil
}
