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

// NewRecoveryMiddleware 创建 panic 恢复中间件
func NewRecoveryMiddleware() IRecoveryMiddleware {
	// 直接返回 component 的中间件，让依赖注入自动处理
	return componentMiddleware.NewRecoveryMiddleware()
}
