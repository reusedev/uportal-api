package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetIntParam 从 URL 参数中获取整数参数
func GetIntParam(c *gin.Context, key string) (int, error) {
	val := c.Param(key)
	return strconv.Atoi(val)
}

// GetIntQuery 从查询参数中获取整数参数
func GetIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	val := c.Query(key)
	if val == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(val)
}

// GetInt64Param 从 URL 参数中获取 int64 参数
func GetInt64Param(c *gin.Context, key string) (int64, error) {
	val := c.Param(key)
	return strconv.ParseInt(val, 10, 64)
}

// GetInt64Query 从查询参数中获取 int64 参数
func GetInt64Query(c *gin.Context, key string, defaultValue int64) (int64, error) {
	val := c.Query(key)
	if val == "" {
		return defaultValue, nil
	}
	return strconv.ParseInt(val, 10, 64)
}

// GetStringParam 从 URL 参数中获取字符串参数
func GetStringParam(c *gin.Context, key string) string {
	return c.Param(key)
}

// GetStringQuery 从查询参数中获取字符串参数
func GetStringQuery(c *gin.Context, key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}
