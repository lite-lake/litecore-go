package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// ISysHealthController 健康检查控制器接口
type ISysHealthController interface {
	common.BaseController
}

type SysHealthController struct {
	componentController componentControllers.IHealthController
}

func NewSysHealthController() ISysHealthController {
	return &SysHealthController{
		componentController: componentControllers.NewHealthController(),
	}
}

func (c *SysHealthController) ControllerName() string {
	return "SysHealthController"
}

func (c *SysHealthController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *SysHealthController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ common.BaseController = (*SysHealthController)(nil)
