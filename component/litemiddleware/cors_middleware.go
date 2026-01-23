package litemiddleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// CorsConfig CORS 配置
type CorsConfig struct {
	Name             string        // 中间件名称
	Order            *int          // 执行顺序（指针类型用于判断是否设置）
	AllowOrigins     []string      // 允许的源
	AllowMethods     []string      // 允许的 HTTP 方法
	AllowHeaders     []string      // 允许的请求头
	ExposeHeaders    []string      // 暴露的响应头
	AllowCredentials bool          // 是否允许携带凭证
	MaxAge           time.Duration // 预检请求缓存时间
}

// DefaultCorsConfig 默认 CORS 配置
func DefaultCorsConfig() *CorsConfig {
	defaultOrder := OrderCORS
	return &CorsConfig{
		Name:             "CorsMiddleware",
		Order:            &defaultOrder,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "Accept", "Cache-Control"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// joinStrings 拼接字符串
func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// corsMiddleware CORS 跨域中间件
type corsMiddleware struct {
	cfg *CorsConfig
}

// NewCorsMiddleware 创建 CORS 中间件
func NewCorsMiddleware(config *CorsConfig) common.IBaseMiddleware {
	if config == nil {
		config = DefaultCorsConfig()
	}
	return &corsMiddleware{cfg: config}
}

// NewCorsMiddlewareWithDefaults 使用默认配置创建 CORS 中间件
func NewCorsMiddlewareWithDefaults() common.IBaseMiddleware {
	return NewCorsMiddleware(nil)
}

// MiddlewareName 返回中间件名称
func (m *corsMiddleware) MiddlewareName() string {
	if m.cfg.Name != "" {
		return m.cfg.Name
	}
	return "CorsMiddleware"
}

// Order 返回执行顺序
func (m *corsMiddleware) Order() int {
	if m.cfg.Order != nil {
		return *m.cfg.Order
	}
	return OrderCORS
}

// Wrapper 返回 Gin 中间件函数
func (m *corsMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if len(m.cfg.AllowOrigins) > 0 {
			for _, allowedOrigin := range m.cfg.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					break
				}
			}
		}

		if m.cfg.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(m.cfg.AllowHeaders) > 0 {
			headers := joinStrings(m.cfg.AllowHeaders, ", ")
			c.Writer.Header().Set("Access-Control-Allow-Headers", headers)
		}

		if len(m.cfg.AllowMethods) > 0 {
			methods := joinStrings(m.cfg.AllowMethods, ", ")
			c.Writer.Header().Set("Access-Control-Allow-Methods", methods)
		}

		if len(m.cfg.ExposeHeaders) > 0 {
			headers := joinStrings(m.cfg.ExposeHeaders, ", ")
			c.Writer.Header().Set("Access-Control-Expose-Headers", headers)
		}

		if m.cfg.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", int(m.cfg.MaxAge.Seconds())))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(common.HTTPStatusNoContent)
			return
		}

		c.Next()
	}
}

// OnStart 服务器启动时触发
func (m *corsMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *corsMiddleware) OnStop() error {
	return nil
}
