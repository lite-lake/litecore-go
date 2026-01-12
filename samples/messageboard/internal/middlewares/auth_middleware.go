// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"com.litelake.litecore/samples/messageboard/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 管理员认证中间件
// 基于 BaseMiddleware 接口实现
// 只对 /api/admin 路径进行认证检查
type AuthMiddleware struct {
	AuthService services.IAuthService `inject:""`
}

// NewAuthMiddleware 创建认证中间件实例
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// MiddlewareName 实现 BaseMiddleware 接口
func (m *AuthMiddleware) MiddlewareName() string {
	return "AuthMiddleware"
}

// Order 实现 BaseMiddleware 接口
// 数值越小越先执行，认证中间件应该在路由之后、控制器之前执行
func (m *AuthMiddleware) Order() int {
	return 100
}

// Wrapper 实现 BaseMiddleware 接口
// 返回 Gin 中间件函数
func (m *AuthMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对 /api/admin 路径进行认证检查
		if !strings.HasPrefix(c.Request.URL.Path, "/api/admin") {
			c.Next()
			return
		}

		// 登录接口不需要认证
		if c.Request.URL.Path == "/api/admin/login" {
			c.Next()
			return
		}

		// 获取 Authorization 请求头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 解析 Bearer Token
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

		// 使用 AuthService 验证会话
		session, err := m.AuthService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 将会话信息存入上下文，供后续处理器使用
		c.Set("admin_session", session)
		c.Next()
	}
}

// OnStart 实现 BaseMiddleware 接口
func (m *AuthMiddleware) OnStart() error {
	return nil
}

// OnStop 实现 BaseMiddleware 接口
func (m *AuthMiddleware) OnStop() error {
	return nil
}

var _ IAuthMiddleware = (*AuthMiddleware)(nil)
