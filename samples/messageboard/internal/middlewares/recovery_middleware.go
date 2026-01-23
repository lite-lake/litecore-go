// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	componentMiddleware "github.com/lite-lake/litecore-go/component/middleware"
)

// IRecoveryMiddleware panic 恢复中间件接口
type IRecoveryMiddleware interface {
	common.IBaseMiddleware
}

// NewRecoveryMiddleware 使用默认配置创建 panic 恢复中间件
func NewRecoveryMiddleware() IRecoveryMiddleware {
	return componentMiddleware.NewRecoveryMiddlewareWithDefaults()
}
