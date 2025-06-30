package service

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
)

func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	os.MkdirAll("tmp", os.ModePerm)
	filePath := "tmp/" + file.Filename
	if err = c.SaveUploadedFile(file, filePath); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "保存文件失败", err))
		return
	}
	defer func() {
		os.Remove(filePath) // 删除临时文件
	}()
	//上传
	uploadFile, err := model.UploadFile(filePath, config.GlobalConfig.DrawApi.UploadFileUrl)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "上传文件失败", err))
		return
	}
	ret := map[string]interface{}{
		"id":  strconv.Itoa(uploadFile.Data.Id),
		"url": uploadFile.Data.Url,
	}
	response.Success(c, ret)
}
