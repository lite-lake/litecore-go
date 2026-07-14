package litecontroller

import (
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Managers  map[string]string `json:"managers,omitempty"`
}

// IHealthController 健康检查控制器接口
//
// Deprecated: 系统路由已由 Engine 自动注册 /api/health（liveness）和 /api/ready（readiness），
// 无需在各 app 中手动创建 HealthController。保留此组件仅为向后兼容。
// 新代码请直接使用 Engine 注册的系统路由。
type IHealthController interface {
	common.IBaseController
}

// HealthController 健康检查控制器实现
//
// Deprecated: 见 IHealthController 文档。
type HealthController struct {
	ManagerContainer common.IBaseManager      `inject:""`
	LoggerMgr        loggermgr.ILoggerManager `inject:""`
}

// NewHealthController 创建健康检查控制器
//
// Deprecated: 见 IHealthController 文档。
func NewHealthController() IHealthController {
	return &HealthController{}
}

func (c *HealthController) ControllerName() string {
	return "HealthController"
}

func (c *HealthController) GetRouter() string {
	return "/api/health [GET]"
}

func (c *HealthController) Handle(ctx *gin.Context) {
	managerStatus := make(map[string]string)
	allHealthy := true

	if c.ManagerContainer != nil {
		for _, mgr := range []common.IBaseManager{c.ManagerContainer} {
			if err := mgr.Health(); err != nil {
				managerStatus[mgr.ManagerName()] = "unhealthy: " + err.Error()
				allHealthy = false
			} else {
				managerStatus[mgr.ManagerName()] = "ok"
			}
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
		ctx.JSON(http.StatusOK, response)
	} else {
		ctx.JSON(http.StatusServiceUnavailable, response)
	}
}

var _ common.IBaseController = (*HealthController)(nil)
