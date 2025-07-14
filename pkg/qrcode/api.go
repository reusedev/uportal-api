package qrcode

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/uportal-api/pkg/logs"
	"io"
	"net/http"
	"os"
)

const (
	getUnlimitedQRCodeUrl = "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token="
)

type QrcodeDataReq struct {
	Scene string `json:"scene"`
}

type QrcodeDataResp struct {
}

// GetQrcode 获取二维码
func GetQrcode(token, scene, savePath string) error {
	if token == "" {
		return nil
	}
	url := getUnlimitedQRCodeUrl + token
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(QrcodeDataReq{Scene: scene}).
		SetDoNotParseResponse(true). // 保留原始 []byte
		Post(url)

	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		logs.Business().Error(fmt.Sprintf("send active failed: %d, url: %s, scene: %s", resp.StatusCode, url, scene))
		return errors.New("send active failed")
	}
	defer resp.RawBody().Close()
	respData, err := io.ReadAll(resp.RawBody())
	if err != nil {
		return err
	}

	if jsoniter.Get(respData, "errcode").ToInt() != 0 {
		logs.Business().Error(fmt.Sprintf("微信二维码接口错误: ,token:%s, ret: %s", token, string(respData)))
		return errors.New("send active failed")
	}

	return os.WriteFile(savePath, respData, 0644)
}
