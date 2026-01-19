package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IHealthzController healthz路由检查控制器接口
type IHealthzController interface {
	common.BaseController
}

type HealthzController struct {
	componentController componentControllers.IHealthzController
}

func NewHealthzController() IHealthzController {
	return &HealthzController{
		componentController: componentControllers.NewHealthzController(),
	}
}

func (c *HealthzController) ControllerName() string {
	return "HealthzController"
}

func (c *HealthzController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *HealthzController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ common.BaseController = (*HealthzController)(nil)
