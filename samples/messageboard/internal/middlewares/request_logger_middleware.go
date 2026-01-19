// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentMiddleware "com.litelake.litecore/component/middleware"
)

// IRequestLoggerMiddleware 请求日志中间件接口
type IRequestLoggerMiddleware interface {
	common.IBaseMiddleware
}

type requestLoggerMiddleware struct {
	inner common.IBaseMiddleware
	order int
}

// NewRequestLoggerMiddleware 创建请求日志中间件
func NewRequestLoggerMiddleware() IRequestLoggerMiddleware {
	return &requestLoggerMiddleware{
		inner: componentMiddleware.NewRequestLoggerMiddleware(),
		order: 20,
	}
}

func (m *requestLoggerMiddleware) MiddlewareName() string {
	return m.inner.MiddlewareName()
}

func (m *requestLoggerMiddleware) Order() int {
	return m.order
}

func (m *requestLoggerMiddleware) Wrapper() gin.HandlerFunc {
	return m.inner.Wrapper()
}

func (m *requestLoggerMiddleware) OnStart() error {
	return m.inner.OnStart()
}

func (m *requestLoggerMiddleware) OnStop() error {
	return m.inner.OnStop()
}

var _ IRequestLoggerMiddleware = (*requestLoggerMiddleware)(nil)
