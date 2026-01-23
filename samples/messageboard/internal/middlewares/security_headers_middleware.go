// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litemiddleware"
)

// ISecurityHeadersMiddleware 安全头中间件接口
type ISecurityHeadersMiddleware interface {
	common.IBaseMiddleware
}

// NewSecurityHeadersMiddleware 使用默认配置创建安全头中间件
func NewSecurityHeadersMiddleware() ISecurityHeadersMiddleware {
	return litemiddleware.NewSecurityHeadersMiddlewareWithDefaults()
}
