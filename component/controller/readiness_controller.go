package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IReadinessController 就绪检查控制器接口
type IReadinessController interface {
	common.BaseController
}

type ReadinessController struct {
	ManagerContainer common.BaseManager `inject:""`
}

func NewReadinessController() IReadinessController {
	return &ReadinessController{}
}

func (c *ReadinessController) ControllerName() string {
	return "ReadinessController"
}

func (c *ReadinessController) GetRouter() string {
	return "/ready [GET]"
}

func (c *ReadinessController) Handle(ctx *gin.Context) {
	allReady := true

	if c.ManagerContainer != nil {
		if err := c.ManagerContainer.Health(); err != nil {
			allReady = false
		}
	}

	if allReady {
		ctx.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

var _ common.BaseController = (*ReadinessController)(nil)
