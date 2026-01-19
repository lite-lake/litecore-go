package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IResStaticController 静态文件控制器接口
type IResStaticController interface {
	common.IBaseController
}

type resStaticControllerImpl struct {
	componentController *componentControllers.ResourceStaticController
}

func NewResStaticController() IResStaticController {
	return &resStaticControllerImpl{
		componentController: componentControllers.NewResourceStaticController("/static", "./static"),
	}
}

func (c *resStaticControllerImpl) ControllerName() string {
	return "resStaticControllerImpl"
}

func (c *resStaticControllerImpl) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *resStaticControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ IResStaticController = (*resStaticControllerImpl)(nil)
