package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// HealthzController healthz路由检查控制器接口
type IHealthzController interface {
	common.BaseController
}

type HealthzController struct {
	ManagerContainer common.BaseManager `inject:""`
}

func NewHealthzController() IHealthzController {
	return &HealthzController{}
}

func (c *HealthzController) ControllerName() string {
	return "HealthzController"
}

func (c *HealthzController) GetRouter() string {
	return "/healthz [GET]"
}

func (c *HealthzController) Handle(ctx *gin.Context) {
	managerStatus := make(map[string]string)
	allHealthy := true

	if c.ManagerContainer != nil {
		for _, mgr := range []common.BaseManager{c.ManagerContainer} {
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

	ctx.JSON(http.StatusOK, gin.H{
		"status":   status,
		"managers": managerStatus,
	})
}

var _ common.BaseController = (*HealthzController)(nil)
