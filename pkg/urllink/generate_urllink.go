package urllink

import (
	"fmt"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"io"
	"net/http"
	"strings"
)

const (
	getUrlLinkUrl = "https://api.weixin.qq.com/wxa/generate_urllink?access_token="
)

func GetUrlLink() {
	token := wechat_token.GetToken()
	if token == "" {
		return
	}
	url := getUrlLinkUrl + token
	data := `{"query":"","expire_type":1,"expire_interval":1,"env_version":"release"}`
	resp, err := http.Post(url, "", strings.NewReader(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	fmt.Println(string(all))
}
