package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common/deployinfo"
)

// registerRoute 注册自定义路由
func (e *Engine) registerRoute(method, path string, handler gin.HandlerFunc) {
	switch strings.ToUpper(method) {
	case "GET":
		e.ginEngine.GET(path, handler)
	case "POST":
		e.ginEngine.POST(path, handler)
	case "PUT":
		e.ginEngine.PUT(path, handler)
	case "DELETE":
		e.ginEngine.DELETE(path, handler)
	case "PATCH":
		e.ginEngine.PATCH(path, handler)
	case "HEAD":
		e.ginEngine.HEAD(path, handler)
	case "OPTIONS":
		e.ginEngine.OPTIONS(path, handler)
	case "ANY":
		e.ginEngine.Any(path, handler)
	default:
		e.ginEngine.GET(path, handler)
	}
}

// registerSystemRoutes 注册系统路由（/api/health、/api/ready）
// 系统路由由 Engine 直接注册，不经过 Controller Container，所有 app 自动获得
func (e *Engine) registerSystemRoutes() {
	e.ginEngine.GET("/api/health", e.handleLiveness)
	e.ginEngine.GET("/api/ready", e.handleReadiness)
}

// handleLiveness 存活探针：进程存活即返回 ok
func (e *Engine) handleLiveness(c *gin.Context) {
	info := deployinfo.Get()
	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"lite_build_id": info.LiteBuildID,
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// handleReadiness 就绪探针：所有 Manager 健康且引擎启动完成后才返回 ok
func (e *Engine) handleReadiness(c *gin.Context) {
	e.mu.RLock()
	isReady := e.ready
	e.mu.RUnlock()

	info := deployinfo.Get()

	if !isReady {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":        "not_ready",
			"lite_build_id": info.LiteBuildID,
			"timestamp":     time.Now().Format(time.RFC3339),
		})
		return
	}

	managerStatus := make(map[string]string)
	allHealthy := true

	if e.Manager != nil {
		for _, mgr := range e.Manager.GetAll() {
			if err := mgr.Health(); err != nil {
				managerStatus[mgr.ManagerName()] = "unhealthy: " + err.Error()
				allHealthy = false
			} else {
				managerStatus[mgr.ManagerName()] = "ok"
			}
		}
	}

	if !allHealthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":        "degraded",
			"lite_build_id": info.LiteBuildID,
			"timestamp":     time.Now().Format(time.RFC3339),
			"managers":      managerStatus,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"lite_build_id": info.LiteBuildID,
		"timestamp":     time.Now().Format(time.RFC3339),
		"managers":      managerStatus,
	})
}
