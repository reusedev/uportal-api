package service

import (
	"context"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestNewAdminService(t *testing.T) {
	password, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	t.Logf(string(password))
}

func TestNewAlipayService(t *testing.T) {
	// 加载商户证书
	mchPrivateKey, err := utils.LoadPrivateKey("cert/apiclient_key.pem")
	if err != nil {
		panic(err)
	}

	// 创建微信支付客户端
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher("1719916090",
			"9fb930a1bce42a9115c2d5c08df36d36", mchPrivateKey, "cert/apiclient_cert.pem"),
	}
	c, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		panic(err)
	}

	svc := jsapi.JsapiApiService{Client: c}
	resp, result, err := svc.PrepayWithRequestPayment(context.Background(),
		jsapi.PrepayRequest{
			Appid:       core.String("wx02cfc4189bbb897c"),
			Mchid:       core.String("1719916090"),
			Description: core.String("测试商品"),
			OutTradeNo:  core.String("11234434554"),
			NotifyUrl:   core.String("https://drawtest.shumengkj.cn/api/payments/wechat/notify"),
			Amount: &jsapi.Amount{
				Total:    core.Int64(1), // 转换为分
				Currency: core.String("CNY"),
			},
			Payer: &jsapi.Payer{
				Openid: core.String("oXIoC7tDQN8WJBQ-6ZepTGa2fx38"),
			},
		},
	)
	fmt.Println(resp, result)
}
