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
	order int
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
	return middlewarepkg.NewRateLimiterByIP(100, time.Minute).Wrapper()
}

func (m *rateLimiterMiddleware) OnStart() error {
	return nil
}

func (m *rateLimiterMiddleware) OnStop() error {
	return nil
}

var _ IRateLimiterMiddleware = (*rateLimiterMiddleware)(nil)
