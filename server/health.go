package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Managers  map[string]string `json:"managers,omitempty"`
}

// ManagerHealthStatus 管理器健康状态
type ManagerHealthStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// registerHealthRoute 注册健康检查路由
func (e *Engine) registerHealthRoute() {
	e.ginEngine.GET("/health", e.healthHandler)
	e.ginEngine.GET("/healthz", e.healthHandler)
	e.ginEngine.GET("/live", e.livenessHandler)
	e.ginEngine.GET("/ready", e.readinessHandler)
}

// healthHandler 健康检查处理器
func (e *Engine) healthHandler(c *gin.Context) {
	managers := e.manager.GetAll()
	managerStatus := make(map[string]string)

	allHealthy := true
	for _, mgr := range managers {
		if err := mgr.Health(); err != nil {
			managerStatus[mgr.ManagerName()] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			managerStatus[mgr.ManagerName()] = "ok"
		}
	}

	status := "ok"
	if !allHealthy {
		status = "degraded"
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Managers:  managerStatus,
	}

	if allHealthy {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// livenessHandler 存活检查处理器
// 用于判断服务是否正在运行
func (e *Engine) livenessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// readinessHandler 就绪检查处理器
// 用于判断服务是否准备好接收流量
func (e *Engine) readinessHandler(c *gin.Context) {
	managers := e.manager.GetAll()

	// 检查所有管理器是否就绪
	allReady := true
	for _, mgr := range managers {
		if err := mgr.Health(); err != nil {
			allReady = false
			break
		}
	}

	if allReady {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// DetailedHealthResponse 详细健康检查响应
type DetailedHealthResponse struct {
	Status    string                `json:"status"`
	Timestamp string                `json:"timestamp"`
	Managers  []ManagerHealthStatus `json:"managers"`
}

// DetailedHealthHandler 详细健康检查处理器
func (e *Engine) DetailedHealthHandler(c *gin.Context) {
	managers := e.manager.GetAll()
	managerStatus := make([]ManagerHealthStatus, 0, len(managers))

	allHealthy := true
	for _, mgr := range managers {
		status := ManagerHealthStatus{
			Name: mgr.ManagerName(),
		}

		if err := mgr.Health(); err != nil {
			status.Status = "unhealthy"
			status.Error = err.Error()
			allHealthy = false
		} else {
			status.Status = "healthy"
		}

		managerStatus = append(managerStatus, status)
	}

	overallStatus := "healthy"
	if !allHealthy {
		overallStatus = "unhealthy"
	}

	response := DetailedHealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Managers:  managerStatus,
	}

	if allHealthy {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// RegisterDetailedHealthRoute 注册详细健康检查路由
func (e *Engine) RegisterDetailedHealthRoute(path string) {
	e.ginEngine.GET(path, e.DetailedHealthHandler)
}
