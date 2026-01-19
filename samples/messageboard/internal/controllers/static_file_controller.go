package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IStaticFileController 静态文件控制器接口
type IStaticFileController interface {
	common.BaseController
}

type StaticFileController struct {
	componentController *componentControllers.StaticFileController
}

func NewStaticFileController() IStaticFileController {
	return &StaticFileController{
		componentController: componentControllers.NewStaticFileController("/static", "./static"),
	}
}

func (c *StaticFileController) ControllerName() string {
	return "StaticFileController"
}

func (c *StaticFileController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *StaticFileController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ IStaticFileController = (*StaticFileController)(nil)
