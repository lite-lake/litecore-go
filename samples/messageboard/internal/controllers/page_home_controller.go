package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/manager/loggermgr"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IPageHomeController 首页控制器接口
type IPageHomeController interface {
	common.IBaseController
}

type pageHomeControllerImpl struct {
	HTMLTemplateService services.IHTMLTemplateService `inject:""` // HTML 模板服务
	LoggerMgr           loggermgr.ILoggerManager      `inject:""` // 日志管理器
}

// NewPageHomeController 创建控制器实例
func NewPageHomeController() IPageHomeController {
	return &pageHomeControllerImpl{}
}

// ControllerName 返回控制器名称
func (c *pageHomeControllerImpl) ControllerName() string {
	return "pageHomeControllerImpl"
}

// GetRouter 返回路由信息
func (c *pageHomeControllerImpl) GetRouter() string {
	return "/ [GET]"
}

// Handle 处理首页请求，渲染 index.html 模板
func (c *pageHomeControllerImpl) Handle(ctx *gin.Context) {
	c.HTMLTemplateService.Render(ctx, "index.html", gin.H{
		"title": "留言板",
	})
}

var _ IPageHomeController = (*pageHomeControllerImpl)(nil)
