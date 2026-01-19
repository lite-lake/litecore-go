package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IHealthController 健康检查控制器接口
type IHealthController interface {
	common.BaseController
}

type HealthController struct {
	componentController componentControllers.IHealthController
}

func NewHealthController() IHealthController {
	return &HealthController{
		componentController: componentControllers.NewHealthController(),
	}
}

func (c *HealthController) ControllerName() string {
	return "HealthController"
}

func (c *HealthController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *HealthController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ common.BaseController = (*HealthController)(nil)
