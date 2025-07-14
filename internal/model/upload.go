package model

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"net/http"
	"os"
)

type Resp struct {
	Code int `json:"code,omitempty"`
	Data struct {
		Id     int    `json:"id"`
		Url    string `json:"url,omitempty"`
		Status string `json:"status,omitempty"`
	}
}

// UploadFile 文件上传-透传
func UploadFile(filePath, url string) (*Resp, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, err
	}
	var result Resp
	resp, err := resty.New().R().SetFile("file", filePath).SetResult(&result).Post(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("is not 200")
	}
	return &result, nil
}

func GetUrlById(id string, url string, tp string) string {
	if id == "" {
		return ""
	}
	var result Resp
	_, err := resty.New().R().SetQueryParams(map[string]string{
		"id":   id,
		"type": tp,
	}).SetResult(&result).Get(url)
	if err != nil {
		return ""
	}
	return result.Data.Url
}
