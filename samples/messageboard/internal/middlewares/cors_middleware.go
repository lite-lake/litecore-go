// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	componentMiddleware "github.com/lite-lake/litecore-go/component/middleware"
)

// ICorsMiddleware CORS 跨域中间件接口
type ICorsMiddleware interface {
	common.IBaseMiddleware
}

type corsMiddleware struct {
	inner common.IBaseMiddleware
	order int
}

// NewCorsMiddleware 创建 CORS 中间件
func NewCorsMiddleware() ICorsMiddleware {
	return &corsMiddleware{
		inner: componentMiddleware.NewCorsMiddleware(),
		order: 30,
	}
}

func (m *corsMiddleware) MiddlewareName() string {
	return m.inner.MiddlewareName()
}

func (m *corsMiddleware) Order() int {
	return m.order
}

func (m *corsMiddleware) Wrapper() gin.HandlerFunc {
	return m.inner.Wrapper()
}

func (m *corsMiddleware) OnStart() error {
	return m.inner.OnStart()
}

func (m *corsMiddleware) OnStop() error {
	return m.inner.OnStop()
}

var _ ICorsMiddleware = (*corsMiddleware)(nil)
