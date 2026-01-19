package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IReadinessController 就绪检查控制器接口
type IReadinessController interface {
	common.BaseController
}

type ReadinessController struct {
	componentController componentControllers.IReadinessController
}

func NewReadinessController() IReadinessController {
	return &ReadinessController{
		componentController: componentControllers.NewReadinessController(),
	}
}

func (c *ReadinessController) ControllerName() string {
	return "ReadinessController"
}

func (c *ReadinessController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *ReadinessController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ common.BaseController = (*ReadinessController)(nil)
