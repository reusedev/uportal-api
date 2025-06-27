package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 日志中间件
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()
		if path == "/" {
			return
		}
		// 记录请求信息
		cost := time.Since(start)
		logger.Info("request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)

		// 如果响应时间超过阈值，记录警告日志
		if cost > time.Second {
			logger.Warn("slow request",
				zap.String("path", path),
				zap.String("query", query),
				zap.Duration("cost", cost),
			)
		}

		// 如果发生错误，记录错误日志
		if len(c.Errors) > 0 {
			logger.Error("request error",
				zap.String("path", path),
				zap.String("query", query),
				zap.String("errors", fmt.Sprintf("%v", c.Errors)),
			)
		}
	}
}
