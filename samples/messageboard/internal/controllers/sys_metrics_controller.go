package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	componentControllers "github.com/lite-lake/litecore-go/component/controller"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
)

// ISysMetricsController 指标控制器接口
type ISysMetricsController interface {
	common.IBaseController
}

type sysMetricsControllerImpl struct {
	componentController componentControllers.IMetricsController
	LoggerMgr           loggermgr.ILoggerManager `inject:""`
	logger              loggermgr.ILogger
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

func (c *sysMetricsControllerImpl) Logger() loggermgr.ILogger {
	return c.logger
}

func (c *sysMetricsControllerImpl) SetLoggerManager(mgr loggermgr.ILoggerManager) {
	c.LoggerMgr = mgr
	c.initLogger()
}

func (c *sysMetricsControllerImpl) initLogger() {
	if c.LoggerMgr != nil {
		c.logger = c.LoggerMgr.Logger("SysMetricsController")
	}
}

func (c *sysMetricsControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ ISysMetricsController = (*sysMetricsControllerImpl)(nil)
