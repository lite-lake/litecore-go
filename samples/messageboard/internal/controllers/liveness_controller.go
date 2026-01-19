package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// ILivenessController 存活检查控制器接口
type ILivenessController interface {
	common.BaseController
}

type LivenessController struct {
	componentController componentControllers.ILivenessController
}

func NewLivenessController() ILivenessController {
	return &LivenessController{
		componentController: componentControllers.NewLivenessController(),
	}
}

func (c *LivenessController) ControllerName() string {
	return "LivenessController"
}

func (c *LivenessController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *LivenessController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ common.BaseController = (*LivenessController)(nil)
