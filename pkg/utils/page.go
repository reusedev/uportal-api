package utils

import (
	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// GetPage 从查询参数中获取页码
func GetPage(c *gin.Context) int {
	page, err := GetIntQuery(c, "page", DefaultPage)
	if err != nil || page < 1 {
		return DefaultPage
	}
	return page
}

// GetPageSize 从查询参数中获取每页数量
func GetPageSize(c *gin.Context) int {
	pageSize, err := GetIntQuery(c, "page_size", DefaultPageSize)
	if err != nil || pageSize < 1 {
		return DefaultPageSize
	}
	if pageSize > MaxPageSize {
		return MaxPageSize
	}
	return pageSize
}

// GetOffset 计算分页偏移量
func GetOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}
