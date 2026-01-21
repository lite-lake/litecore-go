// Package middlewares 定义 HTTP 中间件
package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	componentMiddleware "github.com/lite-lake/litecore-go/component/middleware"
)

// IRecoveryMiddleware panic 恢复中间件接口
type IRecoveryMiddleware interface {
	common.IBaseMiddleware
}

type recoveryMiddleware struct {
	inner common.IBaseMiddleware
	order int
}

// NewRecoveryMiddleware 创建 panic 恢复中间件
func NewRecoveryMiddleware() IRecoveryMiddleware {
	return &recoveryMiddleware{
		inner: componentMiddleware.NewRecoveryMiddleware(),
		order: 10,
	}
}

func (m *recoveryMiddleware) MiddlewareName() string {
	return m.inner.MiddlewareName()
}

func (m *recoveryMiddleware) Order() int {
	return m.order
}

func (m *recoveryMiddleware) Wrapper() gin.HandlerFunc {
	return m.inner.Wrapper()
}

func (m *recoveryMiddleware) OnStart() error {
	return m.inner.OnStart()
}

func (m *recoveryMiddleware) OnStop() error {
	return m.inner.OnStop()
}

var _ IRecoveryMiddleware = (*recoveryMiddleware)(nil)
