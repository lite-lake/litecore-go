// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

// IAuthMiddleware 认证中间件接口
type IAuthMiddleware interface {
	common.BaseMiddleware
}

type authMiddleware struct {
	AuthService services.IAuthService `inject:""`
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware() IAuthMiddleware {
	return &authMiddleware{}
}

func (m *authMiddleware) MiddlewareName() string {
	return "AuthMiddleware"
}

func (m *authMiddleware) Order() int {
	return 100
}

func (m *authMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/admin") {
			c.Next()
			return
		}

		if c.Request.URL.Path == "/api/admin/login" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		token := parts[1]

		session, err := m.AuthService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		c.Set("admin_session", session)
		c.Next()
	}
}

func (m *authMiddleware) OnStart() error {
	return nil
}

func (m *authMiddleware) OnStop() error {
	return nil
}

var _ IAuthMiddleware = (*authMiddleware)(nil)
