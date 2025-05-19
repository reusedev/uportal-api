package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
	"go.uber.org/zap"
)

// Recovery 恢复中间件
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)

				// 返回500错误
				response.Error(c, errors.New(errors.ErrCodeInternal, "Internal Server Error", nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}
