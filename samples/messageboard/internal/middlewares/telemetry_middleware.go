package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litemiddleware"
)

// ITelemetryMiddleware 遥测中间件接口
type ITelemetryMiddleware interface {
	common.IBaseMiddleware
}

// NewTelemetryMiddleware 使用默认配置创建遥测中间件
func NewTelemetryMiddleware() ITelemetryMiddleware {
	return litemiddleware.NewTelemetryMiddlewareWithDefaults()
}
