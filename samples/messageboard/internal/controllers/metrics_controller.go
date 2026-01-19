package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IMetricsController 指标控制器接口
type IMetricsController interface {
	common.BaseController
}

type MetricsController struct {
	componentController componentControllers.IMetricsController
}

func NewMetricsController() IMetricsController {
	return &MetricsController{
		componentController: componentControllers.NewMetricsController(),
	}
}

func (c *MetricsController) ControllerName() string {
	return "MetricsController"
}

func (c *MetricsController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *MetricsController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ common.BaseController = (*MetricsController)(nil)
