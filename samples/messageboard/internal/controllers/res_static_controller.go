package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	componentControllers "github.com/lite-lake/litecore-go/component/controller"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
)

// IResStaticController 静态文件控制器接口
type IResStaticController interface {
	common.IBaseController
}

type resStaticControllerImpl struct {
	componentController *componentControllers.ResourceStaticController
	LoggerMgr           loggermgr.ILoggerManager `inject:""`
	logger              loggermgr.ILogger
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

func (c *resStaticControllerImpl) Logger() loggermgr.ILogger {
	return c.logger
}

func (c *resStaticControllerImpl) SetLoggerManager(mgr loggermgr.ILoggerManager) {
	c.LoggerMgr = mgr
	c.initLogger()
}

func (c *resStaticControllerImpl) initLogger() {
	if c.LoggerMgr != nil {
		c.logger = c.LoggerMgr.Logger("ResStaticController")
	}
}

func (c *resStaticControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ IResStaticController = (*resStaticControllerImpl)(nil)
