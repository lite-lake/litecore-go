package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IResourceStaticController 静态文件控制器接口
type IResourceStaticController interface {
	common.BaseController
}

type ResourceStaticController struct {
	componentController *componentControllers.ResourceStaticController
}

func NewResourceStaticController() IResourceStaticController {
	return &ResourceStaticController{
		componentController: componentControllers.NewResourceStaticController("/static", "./static"),
	}
}

func (c *ResourceStaticController) ControllerName() string {
	return "ResourceStaticController"
}

func (c *ResourceStaticController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *ResourceStaticController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ IResourceStaticController = (*ResourceStaticController)(nil)
