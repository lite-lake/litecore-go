package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Managers  map[string]string `json:"managers,omitempty"`
}

// IHealthController 健康检查控制器接口
type IHealthController interface {
	common.IBaseController
}

type HealthController struct {
	ManagerContainer common.IBaseManager `inject:""`
}

func NewHealthController() IHealthController {
	return &HealthController{}
}

func (c *HealthController) ControllerName() string {
	return "HealthController"
}

func (c *HealthController) GetRouter() string {
	return "/health [GET]"
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
