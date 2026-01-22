package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	componentControllers "github.com/lite-lake/litecore-go/component/controller"
)

// ISysHealthController 健康检查控制器接口
type ISysHealthController interface {
	common.IBaseController
}

type sysHealthControllerImpl struct {
	componentController componentControllers.IHealthController
	Logger              common.ILogger `inject:""`
}

func NewSysHealthController() ISysHealthController {
	return &sysHealthControllerImpl{
		componentController: componentControllers.NewHealthController(),
	}
}

func (c *sysHealthControllerImpl) ControllerName() string {
	return "sysHealthControllerImpl"
}

func (c *sysHealthControllerImpl) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *sysHealthControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ ISysHealthController = (*sysHealthControllerImpl)(nil)
