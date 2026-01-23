// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

// IAuthMiddleware 认证中间件接口
type IAuthMiddleware interface {
	common.IBaseMiddleware
}

type authMiddleware struct {
	AuthService services.IAuthService `inject:""` // 认证服务
}

// NewAuthMiddleware 创建认证中间件实例
func NewAuthMiddleware() IAuthMiddleware {
	return &authMiddleware{}
}

// MiddlewareName 返回中间件名称
func (m *authMiddleware) MiddlewareName() string {
	return "AuthMiddleware"
}

// Order 返回中间件执行顺序
func (m *authMiddleware) Order() int {
	return 100
}

// Wrapper 返回中间件处理函数
// 验证 /api/admin 路径的请求（登录接口除外）
func (m *authMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 非 admin 路径直接放行
		if !strings.HasPrefix(c.Request.URL.Path, "/api/admin") {
			c.Next()
			return
		}

		// 登录接口无需认证
		if c.Request.URL.Path == "/api/admin/login" {
			c.Next()
			return
		}

		// 检查 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(common.HTTPStatusUnauthorized, gin.H{
				"code":    common.HTTPStatusUnauthorized,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(common.HTTPStatusUnauthorized, gin.H{
				"code":    common.HTTPStatusUnauthorized,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 token 有效性
		session, err := m.AuthService.ValidateToken(token)
		if err != nil {
			c.JSON(common.HTTPStatusUnauthorized, gin.H{
				"code":    common.HTTPStatusUnauthorized,
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 将会话信息存入上下文
		c.Set("admin_session", session)
		c.Next()
	}
}

// OnStart 启动时初始化
func (m *authMiddleware) OnStart() error {
	return nil
}

// OnStop 停止时清理
func (m *authMiddleware) OnStop() error {
	return nil
}

var _ IAuthMiddleware = (*authMiddleware)(nil)
