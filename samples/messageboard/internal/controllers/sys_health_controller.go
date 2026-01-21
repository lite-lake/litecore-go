package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	componentControllers "github.com/lite-lake/litecore-go/component/controller"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
)

// ISysHealthController 健康检查控制器接口
type ISysHealthController interface {
	common.IBaseController
}

type sysHealthControllerImpl struct {
	componentController componentControllers.IHealthController
	LoggerMgr           loggermgr.ILoggerManager `inject:""`
	logger              loggermgr.ILogger
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

func (c *sysHealthControllerImpl) Logger() loggermgr.ILogger {
	return c.logger
}

func (c *sysHealthControllerImpl) SetLoggerManager(mgr loggermgr.ILoggerManager) {
	c.LoggerMgr = mgr
	c.initLogger()
}

func (c *sysHealthControllerImpl) initLogger() {
	if c.LoggerMgr != nil {
		c.logger = c.LoggerMgr.Logger("SysHealthController")
	}
}

func (c *sysHealthControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ ISysHealthController = (*sysHealthControllerImpl)(nil)
