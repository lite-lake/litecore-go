package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// ISysMetricsController 指标控制器接口
type ISysMetricsController interface {
	common.IBaseController
}

type sysMetricsControllerImpl struct {
	componentController componentControllers.IMetricsController
}

func NewSysMetricsController() ISysMetricsController {
	return &sysMetricsControllerImpl{
		componentController: componentControllers.NewMetricsController(),
	}
}

func (c *sysMetricsControllerImpl) ControllerName() string {
	return "sysMetricsControllerImpl"
}

func (c *sysMetricsControllerImpl) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *sysMetricsControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ ISysMetricsController = (*sysMetricsControllerImpl)(nil)
