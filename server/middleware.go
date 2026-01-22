package server

import (
	"sort"

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

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Order() < sorted[j].Order()
	})

	return sorted
}
