package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`           // 错误码
	Message string      `json:"message"`        // 错误信息
	Data    interface{} `json:"data,omitempty"` // 响应数据
}

type ResponseList struct {
	Code    int         `json:"code"`           // 错误码
	Message string      `json:"message"`        // 错误信息
	Data    interface{} `json:"data,omitempty"` // 响应数据
	Count   int64       `json:"count"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.ErrCodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, err error) {
	if e, ok := err.(*errors.Error); ok {
		c.JSON(http.StatusOK, Response{
			Code:    e.Code,
			Message: e.Message,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code:    errors.ErrCodeInternal,
		Message: err.Error(),
	})
}

// List 分页列表
type List struct {
	Total    int64       `json:"total"`     // 总数
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页数量
	Data     interface{} `json:"data"`      // 数据列表
}

// ListResponse 返回分页列表
func ListResponse(c *gin.Context, data interface{}, total int64) {
	c.JSON(http.StatusOK, ResponseList{
		Code:    errors.ErrCodeSuccess,
		Message: "success",
		Data:    data,
		Count:   total,
	})
}

// Page 分页参数
type Page struct {
	Page     int `form:"page" binding:"required,min=1"`             // 页码
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"` // 每页数量
}

// GetOffset 获取偏移量
func (p *Page) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *Page) GetLimit() int {
	return p.PageSize
}
