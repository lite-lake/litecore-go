package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/manager/loggermgr"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litecontroller"
)

// ISysHealthController 健康检查控制器接口
type ISysHealthController interface {
	common.IBaseController
}

type sysHealthControllerImpl struct {
	componentController litecontroller.IHealthController // 内置健康检查控制器
	LoggerMgr           loggermgr.ILoggerManager         `inject:""` // 日志管理器
}

// NewSysHealthController 创建控制器实例
func NewSysHealthController() ISysHealthController {
	return &sysHealthControllerImpl{
		componentController: litecontroller.NewHealthController(),
	}
}

// ControllerName 返回控制器名称
func (c *sysHealthControllerImpl) ControllerName() string {
	return "sysHealthControllerImpl"
}

// GetRouter 返回路由信息
func (c *sysHealthControllerImpl) GetRouter() string {
	return c.componentController.GetRouter()
}

// Handle 处理健康检查请求
func (c *sysHealthControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ ISysHealthController = (*sysHealthControllerImpl)(nil)
