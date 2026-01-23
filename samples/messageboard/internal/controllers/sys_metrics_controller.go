package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litecontroller"
)

// ISysMetricsController 指标控制器接口
type ISysMetricsController interface {
	common.IBaseController
}

type sysMetricsControllerImpl struct {
	componentController litecontroller.IMetricsController // 内置指标控制器
	LoggerMgr           loggermgr.ILoggerManager          `inject:""` // 日志管理器
}

// NewSysMetricsController 创建控制器实例
func NewSysMetricsController() ISysMetricsController {
	return &sysMetricsControllerImpl{
		componentController: litecontroller.NewMetricsController(),
	}
}

// ControllerName 返回控制器名称
func (c *sysMetricsControllerImpl) ControllerName() string {
	return "sysMetricsControllerImpl"
}

// GetRouter 返回路由信息
func (c *sysMetricsControllerImpl) GetRouter() string {
	return c.componentController.GetRouter()
}

// Handle 处理指标请求
func (c *sysMetricsControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ ISysMetricsController = (*sysMetricsControllerImpl)(nil)
