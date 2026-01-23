// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"
)

// ITelemetryMiddleware 遥测中间件接口
type ITelemetryMiddleware interface {
	common.IBaseMiddleware
}

type telemetryMiddleware struct {
	wrapperFunc      func(c *gin.Context)
	order            int
	TelemetryManager telemetrymgr.ITelemetryManager `inject:""`
}

// NewTelemetryMiddleware 创建遥测中间件
func NewTelemetryMiddleware() ITelemetryMiddleware {
	return &telemetryMiddleware{
		order: 50,
	}
}

func (m *telemetryMiddleware) MiddlewareName() string {
	return "TelemetryMiddleware"
}

func (m *telemetryMiddleware) Order() int {
	return m.order
}

func (m *telemetryMiddleware) Wrapper() gin.HandlerFunc {
	if m.wrapperFunc != nil {
		return m.wrapperFunc
	}
	// 如果 TelemetryManager 为空，返回空中间件
	if m.TelemetryManager == nil {
		m.wrapperFunc = func(c *gin.Context) {
			c.Next()
		}
		return m.wrapperFunc
	}
	// 使用 TelemetryManager 的 GinMiddleware
	if ginMiddleware, ok := m.TelemetryManager.(interface{ GinMiddleware() gin.HandlerFunc }); ok {
		m.wrapperFunc = ginMiddleware.GinMiddleware()
	} else {
		m.wrapperFunc = func(c *gin.Context) {
			c.Next()
		}
	}
	return m.wrapperFunc
}

func (m *telemetryMiddleware) OnStart() error {
	return nil
}

func (m *telemetryMiddleware) OnStop() error {
	return nil
}

var _ ITelemetryMiddleware = (*telemetryMiddleware)(nil)
