package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"
)

// TelemetryConfig 遥测配置
type TelemetryConfig struct {
	Name  string // 中间件名称
	Order *int   // 执行顺序（指针类型用于判断是否设置）
}

// DefaultTelemetryConfig 默认遥测配置
func DefaultTelemetryConfig() *TelemetryConfig {
	defaultOrder := OrderTelemetry
	return &TelemetryConfig{
		Name:  "TelemetryMiddleware",
		Order: &defaultOrder,
	}
}

// telemetryMiddleware 遥测中间件
type telemetryMiddleware struct {
	TelemetryManager telemetrymgr.ITelemetryManager `inject:""`
	cfg              *TelemetryConfig
}

// NewTelemetryMiddleware 创建遥测中间件
func NewTelemetryMiddleware(config *TelemetryConfig) common.IBaseMiddleware {
	if config == nil {
		config = DefaultTelemetryConfig()
	}
	return &telemetryMiddleware{cfg: config}
}

// NewTelemetryMiddlewareWithDefaults 使用默认配置创建遥测中间件
func NewTelemetryMiddlewareWithDefaults() common.IBaseMiddleware {
	return NewTelemetryMiddleware(nil)
}

// MiddlewareName 返回中间件名称
func (m *telemetryMiddleware) MiddlewareName() string {
	if m.cfg.Name != "" {
		return m.cfg.Name
	}
	return "TelemetryMiddleware"
}

// Order 返回执行顺序
func (m *telemetryMiddleware) Order() int {
	if m.cfg.Order != nil {
		return *m.cfg.Order
	}
	return OrderTelemetry
}

// Wrapper 返回 Gin 中间件函数
func (m *telemetryMiddleware) Wrapper() gin.HandlerFunc {
	if m.TelemetryManager != nil {
		if otelMiddleware, ok := m.TelemetryManager.(interface {
			GinMiddleware() gin.HandlerFunc
		}); ok {
			return otelMiddleware.GinMiddleware()
		}
	}
	return func(c *gin.Context) {
		c.Next()
	}
}

// OnStart 服务器启动时触发
func (m *telemetryMiddleware) OnStart() error {
	return nil
}

// OnStop 服务器停止时触发
func (m *telemetryMiddleware) OnStop() error {
	return nil
}
