package urllink

import (
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
	"strings"
)

const (
	getSchemeUrl = "https://api.weixin.qq.com/wxa/generatescheme?access_token="
)

func GetScheme(token string) string {
	if token == "" {
		return ""
	}
	url := getSchemeUrl + token
	data := `{"is_expire":true,"expire_type":1,"expire_interval":30}`
	resp, err := http.Post(url, "", strings.NewReader(data))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	dp := jsoniter.Get(all, "openlink").ToString()
	return dp
}
