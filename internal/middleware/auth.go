package middleware

import (
	"github.com/reusedev/uportal-api/pkg/consts"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/jwt"
	"github.com/reusedev/uportal-api/pkg/response"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errors.New(errors.ErrCodeUnauthorized, "未提供认证信息", nil))
			c.Abort()
			return
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, errors.New(errors.ErrCodeUnauthorized, "无效的认证格式", nil))
			c.Abort()
			return
		}

		// 解析token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.Error(c, errors.New(errors.ErrCodeUnauthorized, "无效的token", err))
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set("user_id", claims.UserID)

		// 检查是否是管理员路由
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/admin") {
			if !claims.IsAdmin {
				response.Error(c, errors.New(errors.ErrCodeForbidden, "需要管理员权限", nil))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// Auth 用户认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := jwt.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set(consts.UserId, claims.UserIdStr)
		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := jwt.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 验证是否为管理员
		if !claims.IsAdmin {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "需要管理员权限"})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set(consts.UserId, claims.UserID)
		c.Next()
	}
}
