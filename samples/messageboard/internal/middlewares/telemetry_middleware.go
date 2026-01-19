// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentMiddleware "com.litelake.litecore/component/middleware"
	"com.litelake.litecore/samples/messageboard/internal/infras"
)

// ITelemetryMiddleware 遥测中间件接口
type ITelemetryMiddleware interface {
	common.BaseMiddleware
}

type telemetryMiddleware struct {
	inner            common.BaseMiddleware
	order            int
	TelemetryManager infras.TelemetryManager `inject:""`
}

// NewTelemetryMiddleware 创建遥测中间件
func NewTelemetryMiddleware() ITelemetryMiddleware {
	return &telemetryMiddleware{
		inner: componentMiddleware.NewTelemetryMiddleware(nil),
		order: 50,
	}
}

func (m *telemetryMiddleware) MiddlewareName() string {
	return m.inner.MiddlewareName()
}

func (m *telemetryMiddleware) Order() int {
	return m.order
}

func (m *telemetryMiddleware) Wrapper() gin.HandlerFunc {
	if m.TelemetryManager == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	m.inner = componentMiddleware.NewTelemetryMiddleware(m.TelemetryManager)
	return m.inner.Wrapper()
}

func (m *telemetryMiddleware) OnStart() error {
	if m.inner == nil {
		return nil
	}
	return m.inner.OnStart()
}

func (m *telemetryMiddleware) OnStop() error {
	if m.inner == nil {
		return nil
	}
	return m.inner.OnStop()
}

var _ ITelemetryMiddleware = (*telemetryMiddleware)(nil)
