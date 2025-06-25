package service

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
	"os"
	"strconv"
)

func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
	}
	os.MkdirAll("tmp", os.ModePerm)
	filePath := "tmp/" + file.Filename
	if err = c.SaveUploadedFile(file, filePath); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
	}
	defer func() {
		os.Remove(filePath) // 删除临时文件
	}()
	//上传
	uploadFile, err := model.UploadFile(filePath, config.GlobalConfig.DrawApi.UploadFileUrl)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
	}
	ret := map[string]interface{}{
		"id":  strconv.Itoa(uploadFile.Data.Id),
		"url": uploadFile.Data.Url,
	}
	response.Success(c, ret)
}
