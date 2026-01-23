// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	componentMiddleware "github.com/lite-lake/litecore-go/component/middleware"
)

// IRequestLoggerMiddleware 请求日志中间件接口
type IRequestLoggerMiddleware interface {
	common.IBaseMiddleware
}

// NewRequestLoggerMiddleware 使用默认配置创建请求日志中间件
func NewRequestLoggerMiddleware() IRequestLoggerMiddleware {
	return componentMiddleware.NewRequestLoggerMiddlewareWithDefaults()
}
