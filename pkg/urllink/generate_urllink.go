package urllink

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"io"
	"net/http"
	"strings"
)

const (
	getUrlLinkUrl = "https://api.weixin.qq.com/wxa/generate_urllink?access_token="
)

func GetUrlLink(workId string) string {
	token := wechat_token.GetToken()
	if token == "" {
		return ""
	}
	url := getUrlLinkUrl + token
	data := fmt.Sprintf(`{"path":"pages/creative/creative", "query":"id=%s","expire_type":1,"expire_interval":30,"env_version":"release"}`, workId)
	resp, err := http.Post(url, "", strings.NewReader(data))
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	fmt.Println(string(all))
	link := jsoniter.Get(all, "url_link").ToString()
	return link
}
