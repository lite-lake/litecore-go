// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	middlewarepkg "github.com/lite-lake/litecore-go/component/middleware"
)

type IRateLimiterMiddleware interface {
	common.IBaseMiddleware
}

type rateLimiterMiddleware struct {
	wrapperFunc func(c *gin.Context)
	order       int
}

func NewRateLimiterMiddleware() IRateLimiterMiddleware {
	return &rateLimiterMiddleware{order: 45}
}

func (m *rateLimiterMiddleware) MiddlewareName() string {
	return "RateLimiterMiddleware"
}

func (m *rateLimiterMiddleware) Order() int {
	return m.order
}

func (m *rateLimiterMiddleware) Wrapper() gin.HandlerFunc {
	if m.wrapperFunc == nil {
		// 创建限流中间件并缓存
		m.wrapperFunc = middlewarepkg.NewRateLimiterByIP(100, time.Minute).Wrapper()
	}
	return m.wrapperFunc
}

func (m *rateLimiterMiddleware) OnStart() error {
	return nil
}

func (m *rateLimiterMiddleware) OnStop() error {
	return nil
}

var _ IRateLimiterMiddleware = (*rateLimiterMiddleware)(nil)
