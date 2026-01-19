package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IResStaticController 静态文件控制器接口
type IResStaticController interface {
	common.BaseController
}

type ResStaticController struct {
	componentController *componentControllers.ResourceStaticController
}

func NewResStaticController() IResStaticController {
	return &ResStaticController{
		componentController: componentControllers.NewResourceStaticController("/static", "./static"),
	}
}

func (c *ResStaticController) ControllerName() string {
	return "ResStaticController"
}

func (c *ResStaticController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *ResStaticController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ IResStaticController = (*ResStaticController)(nil)
