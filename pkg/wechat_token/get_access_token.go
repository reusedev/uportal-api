package wechat_token

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

// GetStableAccessToken 官方文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getStableAccessToken.html
func getStableAccessToken(appId, secret string, forceRefresh bool) (string, error) {
	type reqBody struct {
		GrantType    string `json:"grant_type"`
		AppID        string `json:"appid"`
		Secret       string `json:"secret"`
		ForceRefresh bool   `json:"force_refresh,omitempty"`
	}
	type respBody struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	body := reqBody{
		GrantType:    "client_credential",
		AppID:        appId,
		Secret:       secret,
		ForceRefresh: forceRefresh,
	}

	data, err := jsoniter.Marshal(body)
	if err != nil {
		return "", err
	}
	resp, err := http.Post("https://api.weixin.qq.com/cgi-bin/stable_token", "application/json", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result respBody
	if err := jsoniter.Unmarshal(all, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return result.AccessToken, nil
}
