// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litemiddleware"
)

// ICorsMiddleware CORS 跨域中间件接口
type ICorsMiddleware interface {
	common.IBaseMiddleware
}

// NewCorsMiddleware 使用默认配置创建 CORS 中间件
func NewCorsMiddleware() ICorsMiddleware {
	return litemiddleware.NewCorsMiddlewareWithDefaults()
}
