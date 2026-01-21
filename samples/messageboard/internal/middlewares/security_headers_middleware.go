// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	componentMiddleware "github.com/lite-lake/litecore-go/component/middleware"
)

// ISecurityHeadersMiddleware 安全头中间件接口
type ISecurityHeadersMiddleware interface {
	common.IBaseMiddleware
}

type securityHeadersMiddleware struct {
	inner common.IBaseMiddleware
	order int
}

// NewSecurityHeadersMiddleware 创建安全头中间件
func NewSecurityHeadersMiddleware() ISecurityHeadersMiddleware {
	return &securityHeadersMiddleware{
		inner: componentMiddleware.NewSecurityHeadersMiddleware(),
		order: 40,
	}
}

func (m *securityHeadersMiddleware) MiddlewareName() string {
	return m.inner.MiddlewareName()
}

func (m *securityHeadersMiddleware) Order() int {
	return m.order
}

func (m *securityHeadersMiddleware) Wrapper() gin.HandlerFunc {
	return m.inner.Wrapper()
}

func (m *securityHeadersMiddleware) OnStart() error {
	return m.inner.OnStart()
}

func (m *securityHeadersMiddleware) OnStop() error {
	return m.inner.OnStop()
}

var _ ISecurityHeadersMiddleware = (*securityHeadersMiddleware)(nil)
