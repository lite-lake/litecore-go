// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	componentMiddleware "github.com/lite-lake/litecore-go/component/middleware"
)

// ICorsMiddleware CORS 跨域中间件接口
type ICorsMiddleware interface {
	common.IBaseMiddleware
}

// NewCorsMiddleware 创建 CORS 中间件
func NewCorsMiddleware() ICorsMiddleware {
	// 直接返回 component 的中间件
	return componentMiddleware.NewCorsMiddleware()
}
