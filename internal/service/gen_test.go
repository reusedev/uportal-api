package service

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestNewAdminService(t *testing.T) {
	password, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	t.Logf(string(password))
}

//func TestNewAlipayService(t *testing.T) {
//	svc := jsapi.JsapiApiService{Client: s.wxPayClient}
//	resp, _, err := svc.PrepayWithRequestPayment(ctx,
//		jsapi.PrepayRequest{
//			Appid:       core.String(s.config.Wechat.Pay.AppID),
//			Mchid:       core.String(s.config.Wechat.Pay.MchID),
//			Description: core.String(description),
//			OutTradeNo:  core.String(tradeNo),
//			NotifyUrl:   core.String(s.config.Wechat.Pay.NotifyUrl),
//			Amount: &jsapi.Amount{
//				Total:    core.Int64(int64(amount * 100)), // 转换为分
//				Currency: core.String("CNY"),
//			},
//			Payer: &jsapi.Payer{
//				Openid: core.String(userAuth.ProviderUserID),
//			},
//		},
//	)
//}
