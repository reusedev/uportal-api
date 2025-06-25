package wechat_token

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/uportal-api/pkg/logs"
	"io"
	"net/http"
	"net/url"
)

const (
	getTokenUrl = "https://api.weixin.qq.com/cgi-bin/token"
)

// GetToken1 官方文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getAccessToken.html
func getToken(appId, secret string) string {
	u, _ := url.Parse(getTokenUrl)
	query := u.Query()
	query.Set("grant_type", "client_credential")
	query.Set("appid", appId)
	query.Set("secret", secret)
	u.RawQuery = query.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		logs.Business().Error(err.Error())
		return ""
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	return jsoniter.Get(all, "access_token").ToString()
}
