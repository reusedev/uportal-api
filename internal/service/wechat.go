package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/logs"
	"go.uber.org/zap"
)

// WechatService 微信服务
type WechatService struct {
	config *config.Config
}

// NewWechatService 创建微信服务实例
func NewWechatService(config *config.Config) *WechatService {
	return &WechatService{
		config: config,
	}
}

// WxLoginResponse 微信登录响应
type WxLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WxLoginRequest 微信登录请求
type WxLoginRequest struct {
	Code          string  `json:"code" binding:"required"`
	Nickname      *string `json:"nickname" binding:"required,min=2,max=50"`
	AvatarURL     *string `json:"avatar_url" binding:"omitempty,url"`
	EncryptedData string  `json:"encrypted_data" binding:"omitempty"`
	IV            string  `json:"iv" binding:"omitempty"`
}

// WxLoginResult 微信登录结果
type WxLoginResult struct {
	OpenID     string
	SessionKey string
	UnionID    string
}

// Login 微信小程序登录
func (s *WechatService) Login(ctx context.Context, req *WxLoginRequest) (*WxLoginResult, error) {
	// 构建请求URL
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.config.Wechat.MiniProgram.AppID,
		s.config.Wechat.MiniProgram.AppSecret,
		req.Code,
	)

	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New(errors.ErrCodeWechatLoginFailed, "请求微信服务器失败", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var wxResp WxLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&wxResp); err != nil {
		return nil, errors.New(errors.ErrCodeWechatLoginFailed, "解析微信响应失败", err)
	}

	// 检查错误
	if wxResp.ErrCode != 0 {
		logs.Business().Error("微信登录失败",
			zap.Int("errcode", wxResp.ErrCode),
			zap.String("errmsg", wxResp.ErrMsg),
		)
		return nil, errors.New(errors.ErrCodeWechatLoginFailed, fmt.Sprintf("微信登录失败: %s", wxResp.ErrMsg), nil)
	}

	// 记录登录日志
	logs.Business().Info("微信登录成功",
		zap.String("openid", wxResp.OpenID),
		zap.String("unionid", wxResp.UnionID),
	)

	return &WxLoginResult{
		OpenID:     wxResp.OpenID,
		SessionKey: wxResp.SessionKey,
		UnionID:    wxResp.UnionID,
	}, nil
}

// DecryptUserInfo 解密用户信息
func (s *WechatService) DecryptUserInfo(sessionKey, encryptedData, iv string) (map[string]interface{}, error) {
	// TODO: 实现用户信息解密
	// 这里需要使用微信提供的解密算法
	// 可以参考：https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/signature.html
	return nil, nil
}

// GetAccessToken 获取小程序全局接口调用凭据
func (s *WechatService) GetAccessToken(ctx context.Context) (string, error) {
	// TODO: 实现获取 access_token 的逻辑
	// 需要实现缓存机制，因为 access_token 有效期较长
	return "", nil
}

// SendSubscribeMessage 发送订阅消息
func (s *WechatService) SendSubscribeMessage(ctx context.Context, openID, templateID string, data map[string]interface{}, page string) error {
	// TODO: 实现发送订阅消息的逻辑
	return nil
}
