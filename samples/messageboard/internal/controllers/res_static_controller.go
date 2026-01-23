package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/manager/loggermgr"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litecontroller"
)

// IResStaticController 静态文件控制器接口
type IResStaticController interface {
	common.IBaseController
}

type resStaticControllerImpl struct {
	componentController *litecontroller.ResourceStaticController // 内置静态资源控制器
	LoggerMgr           loggermgr.ILoggerManager                 `inject:""` // 日志管理器
}

// NewResStaticController 创建控制器实例
func NewResStaticController() IResStaticController {
	return &resStaticControllerImpl{
		componentController: litecontroller.NewResourceStaticController("/static", "./static"),
	}
}

// ControllerName 返回控制器名称
func (c *resStaticControllerImpl) ControllerName() string {
	return "resStaticControllerImpl"
}

// GetRouter 返回路由信息
func (c *resStaticControllerImpl) GetRouter() string {
	return c.componentController.GetRouter()
}

// Handle 处理静态文件请求
func (c *resStaticControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ IResStaticController = (*resStaticControllerImpl)(nil)
