package message

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/uportal-api/pkg/logs"
	"io"
	"net/http"
	"strings"
)

const (
	sendMessageUrl = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token="
)

type MessageData struct {
	Touser           string        `json:"touser"`            // 接收者（用户）的 openid， 必须
	TemplateId       string        `json:"template_id"`       // 所需下发的订阅模板id， 必须
	Page             string        `json:"page,omitempty"`    // 点击模板卡片后的跳转页面，仅限本小程序内的页面。支持带参数,（示例index?foo=bar）。该字段不填则模板无跳转
	MiniprogramState string        `json:"miniprogram_state"` // 跳转小程序类型：developer为开发版；trial为体验版；formal为正式版；默认为正式版
	Lang             string        `json:"lang"`              // 进入小程序查看”的语言类型，支持zh_CN(简体中文)、en_US(英文)、zh_HK(繁体中文)、zh_TW(繁体中文)，默认为zh_CN
	Data             map[string]Kv `json:"data"`              // 模板内容，格式形如 { "key1": { "value": any }, "key2": { "value": any } }的object, 必须
}

type Kv struct {
	Value string `json:"value"`
}

// SendMessage 发送订阅消息
func SendMessage(token, data string) error {
	if token == "" {
		return nil
	}
	url := sendMessageUrl + token
	resp, err := http.Post(url, "", strings.NewReader(data))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		logs.Business().Error(fmt.Sprintf("send active failed: %d, url: %s, data: %s", resp.StatusCode, url, data))
		return errors.New("send active failed")
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if jsoniter.Get(all, "errcode").ToInt() != 0 {
		logs.Business().Error(fmt.Sprintf("send message failed ,token:%s, ret: %s", token, string(all)))
		return errors.New("send active failed")
	}
	return nil
}
