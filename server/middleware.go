package server

import (
	"sort"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/logger"
)

// registerMiddlewares 注册中间件
// 内置 system 中间件（Order 0-49，如 client host restore）由 Engine 直接注入，
// 与应用注册的中间件合并后统一按 Order 排序注册，早于 request-logger / CORS /
// authz 与全部应用中间件执行。
func (e *Engine) registerMiddlewares() error {
	appMiddlewares := e.Middleware.GetAll()
	middlewares := make([]common.IBaseMiddleware, 0, len(appMiddlewares)+1)
	middlewares = append(middlewares, NewClientHostRestoreMiddleware())
	middlewares = append(middlewares, appMiddlewares...)

	sortedMiddlewares := sortMiddlewares(middlewares)
	registeredCount := 0

	for _, mw := range sortedMiddlewares {
		e.ginEngine.Use(mw.Wrapper())
		mwType := "global"
		if _, ok := mw.(systemMiddleware); ok {
			mwType = "system"
		}
		e.logStartup(PhaseRouter, "Registering middleware",
			logger.F("middleware", mw.MiddlewareName()),
			logger.F("type", mwType))
		registeredCount++
	}

	e.logStartup(PhaseRouter, "Middleware registration completed",
		logger.F("middleware_count", registeredCount))

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
