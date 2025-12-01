package qrcode

import (
	"fmt"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"testing"
)

func TestGetWorksQrcode(t *testing.T) {
	configPath := "/Users/love/GolandProjects/shuzilm/uportal-api/config/config_prod.yaml"
	if err := config.LoadConfig(configPath); err != nil {
		panic(fmt.Sprintf("Load config error: %v", err))
	}
	wechat_token.TokenJob()
	token := wechat_token.GetToken()
	scene := "utm_source=customer_service"
	page := "pages/recharge/recharge"
	GetQrcode(token, scene, "./tmp.png", page, true)
}
