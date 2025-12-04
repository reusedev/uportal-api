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

	UserId   = "user_id"
	SetToken = "Set-Token"
)

const (
	PriceEnable  = 1
	PriceDisable = 0
)

const (
	Accept = "accept"
)

const (
	SendReady  = 0 // 待发送
	Sending    = 1 // 发送中
	SendFinish = 2 // 已发送
)

const (
	//AI绘画完成通知
	CompleteNotificationTmpId = "RqmsNBiXK9bClJ83Z0PkCguj1-wmetG6uGRz1qipf5w"
)

const (
	Wechat = "wechat"
)

const (
	OrderSeq = "order_"
)

const (
	Advertiser = "5CufYHrGRhd"
)
