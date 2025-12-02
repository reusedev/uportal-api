package urllink

import (
	"fmt"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"testing"
)

func TestGetToken(t *testing.T) {
	//GetToken()
	configPath := "/Users/love/GolandProjects/shuzilm/uportal-api/config/config_prod.yaml"
	if err := config.LoadConfig(configPath); err != nil {
		panic(fmt.Sprintf("Load config error: %v", err))
	}
	wechat_token.TokenJob()
	GetUrlLink("16")
	token := wechat_token.GetToken()
	t.Log(GetScheme(token))
}
