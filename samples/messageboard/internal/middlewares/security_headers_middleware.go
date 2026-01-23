// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	componentMiddleware "github.com/lite-lake/litecore-go/component/middleware"
)

// ISecurityHeadersMiddleware 安全头中间件接口
type ISecurityHeadersMiddleware interface {
	common.IBaseMiddleware
}

// NewSecurityHeadersMiddleware 创建安全头中间件
func NewSecurityHeadersMiddleware() ISecurityHeadersMiddleware {
	// 直接返回 component 的中间件，让依赖注入自动处理
	return componentMiddleware.NewSecurityHeadersMiddleware()
}
