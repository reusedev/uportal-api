package errors

import (
	"fmt"
	"net/http"
)

// Error 业务错误类型
type Error struct {
	Code    int    `json:"code"`    // 错误码
	Message string `json:"message"` // 错误信息
	Err     error  `json:"-"`       // 原始错误
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code: %d, message: %s, error: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *Error) Unwrap() error {
	return e.Err
}

// HTTPStatus 获取错误对应的HTTP状态码
func (e *Error) HTTPStatus() int {
	switch e.Code {
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeInvalidParams:
		return http.StatusBadRequest
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// New 创建新的业务错误
func New(code int, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 错误代码定义
const (
	// 系统级错误码 (1000-1999)
	ErrCodeSuccess            = 0    // 成功
	ErrCodeInternal           = 1000 // 内部错误
	ErrCodeInvalidParams      = 1001 // 无效参数
	ErrCodeUnauthorized       = 1002 // 未授权
	ErrCodeForbidden          = 1003 // 禁止访问
	ErrCodeNotFound           = 1004 // 资源不存在
	ErrCodeServiceUnavailable = 1005 // 服务不可用

	// 用户相关错误码 (2000-2999)
	ErrCodeUserNotFound      = 2000 // 用户不存在
	ErrCodeUserDisabled      = 2001 // 用户已禁用
	ErrCodeInvalidPassword   = 2002 // 密码错误
	ErrCodePhoneExists       = 2003 // 手机号已存在
	ErrCodeEmailExists       = 2004 // 邮箱已存在
	ErrCodeInvalidPhone      = 2005 // 无效的手机号
	ErrCodeInvalidEmail      = 2006 // 无效的邮箱
	ErrCodeInvalidVerifyCode = 2007 // 无效的验证码

	// 微信相关错误码 (3000-3999)
	ErrCodeWechatLoginFailed = 3000 // 微信登录失败
	ErrCodeWechatPayFailed   = 3001 // 微信支付失败

	// 代币相关错误码 (4000-4999)
	ErrCodeInsufficientBalance = 4000 // 余额不足
	ErrCodeInvalidAmount       = 4001 // 无效的金额
	ErrCodeTaskNotAvailable    = 4002 // 任务不可用
	ErrCodeTaskLimitExceeded   = 4003 // 任务次数超限

	// 系统级错误码 (10000-10099)
	ErrCodeSystemError     = 10000
	ErrCodeDatabaseError   = 10005
	ErrCodeRedisError      = 10006
	ErrCodeThirdPartyError = 10007
)

// 预定义错误
var (
	// 系统级错误
	ErrInternal           = New(ErrCodeInternal, "服务器内部错误", nil)
	ErrInvalidParams      = New(ErrCodeInvalidParams, "无效的参数", nil)
	ErrUnauthorized       = New(ErrCodeUnauthorized, "未授权访问", nil)
	ErrForbidden          = New(ErrCodeForbidden, "禁止访问", nil)
	ErrNotFound           = New(ErrCodeNotFound, "资源不存在", nil)
	ErrServiceUnavailable = New(ErrCodeServiceUnavailable, "服务暂时不可用", nil)

	// 用户相关错误
	ErrUserNotFound      = New(ErrCodeUserNotFound, "用户不存在", nil)
	ErrUserDisabled      = New(ErrCodeUserDisabled, "用户已被禁用", nil)
	ErrInvalidPassword   = New(ErrCodeInvalidPassword, "密码错误", nil)
	ErrPhoneExists       = New(ErrCodePhoneExists, "手机号已被注册", nil)
	ErrEmailExists       = New(ErrCodeEmailExists, "邮箱已被注册", nil)
	ErrInvalidPhone      = New(ErrCodeInvalidPhone, "无效的手机号", nil)
	ErrInvalidEmail      = New(ErrCodeInvalidEmail, "无效的邮箱地址", nil)
	ErrInvalidVerifyCode = New(ErrCodeInvalidVerifyCode, "无效的验证码", nil)

	// 微信相关错误
	ErrWechatLoginFailed = New(ErrCodeWechatLoginFailed, "微信登录失败", nil)
	ErrWechatPayFailed   = New(ErrCodeWechatPayFailed, "微信支付失败", nil)

	// 代币相关错误
	ErrInsufficientBalance = New(ErrCodeInsufficientBalance, "余额不足", nil)
	ErrInvalidAmount       = New(ErrCodeInvalidAmount, "无效的金额", nil)
	ErrTaskNotAvailable    = New(ErrCodeTaskNotAvailable, "任务不可用", nil)
	ErrTaskLimitExceeded   = New(ErrCodeTaskLimitExceeded, "任务次数已达上限", nil)
)
