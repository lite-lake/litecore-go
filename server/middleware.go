package server

import (
	"github.com/lite-lake/litecore-go/common"
)

// registerMiddlewares 注册中间件
func (e *Engine) registerMiddlewares() error {
	middlewares := e.Middleware.GetAll()

	sortedMiddlewares := sortMiddlewares(middlewares)

	for _, mw := range sortedMiddlewares {
		e.ginEngine.Use(mw.Wrapper())
	}

	return nil
}

// sortMiddlewares 按 Order 排序中间件
func sortMiddlewares(middlewares []common.IBaseMiddleware) []common.IBaseMiddleware {
	sorted := make([]common.IBaseMiddleware, len(middlewares))
	copy(sorted, middlewares)

	n := len(sorted)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if sorted[j].Order() > sorted[j+1].Order() {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}
