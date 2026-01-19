package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// ISysMetricsController 指标控制器接口
type ISysMetricsController interface {
	common.BaseController
}

type SysMetricsController struct {
	componentController componentControllers.IMetricsController
}

func NewSysMetricsController() ISysMetricsController {
	return &SysMetricsController{
		componentController: componentControllers.NewMetricsController(),
	}
}

func (c *SysMetricsController) ControllerName() string {
	return "SysMetricsController"
}

func (c *SysMetricsController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *SysMetricsController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ common.BaseController = (*SysMetricsController)(nil)
