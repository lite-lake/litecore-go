package server

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

// registerMetricsRoute 注册 Prometheus 指标路由
func (e *Engine) registerMetricsRoute() {
	e.ginEngine.GET("/metrics", e.metricsHandler)
}

// metricsHandler Prometheus 指标处理器
func (e *Engine) metricsHandler(c *gin.Context) {
	// 如果有 TelemetryManager，使用其指标收集功能
	// 否则返回默认的指标信息

	// 简单实现：返回基本的指标信息
	metrics := map[string]interface{}{
		"server":  "litecore-go",
		"status":  "running",
		"version": "1.0.0",
	}

	c.JSON(http.StatusOK, metrics)
}

// registerPprofRoutes 注册 pprof 性能分析路由
func (e *Engine) registerPprofRoutes() {
	// pprof 路由组
	pprofGroup := e.ginEngine.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
		pprofGroup.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
		pprofGroup.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
		pprofGroup.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
		pprofGroup.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
		pprofGroup.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
	}
}

// MetricsResponse 指标响应
type MetricsResponse struct {
	// 可以添加更多指标字段
}

// GetMetrics 获取当前指标
func (e *Engine) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 添加容器统计信息
	metrics["managers"] = e.Manager.Count()
	metrics["entities"] = e.Entity.Count()
	metrics["repositories"] = e.Repository.Count()
	metrics["services"] = e.Service.Count()
	metrics["controllers"] = e.Controller.Count()
	metrics["middlewares"] = e.Middleware.Count()

	return metrics
}
