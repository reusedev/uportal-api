package consts

// 用户相关常量
const (
	// 用户类型
	UserTypeAdmin = "admin"
	UserTypeUser  = "user"

	// 用户状态
	UserStatusDisabled = 0
	UserStatusNormal   = 1

	// 用户角色
	UserRoleAdmin      = "admin"
	UserRoleSuperAdmin = "super_admin"

	// 认证类型
	AuthTypePassword = "password"
	AuthTypePhone    = "phone"
	AuthTypeEmail    = "email"
	AuthTypeWechat   = "wechat"
	AuthTypeAlipay   = "alipay"

	// 登录状态
	LoginStatusFailed  = 0
	LoginStatusSuccess = 1

	UserId = "user_id"
)
