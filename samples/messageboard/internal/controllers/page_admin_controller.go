package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IPageAdminController 管理页面控制器接口
type IPageAdminController interface {
	common.IBaseController
}

type pageAdminControllerImpl struct {
	HTMLTemplateService services.IHTMLTemplateService `inject:""` // HTML 模板服务
	LoggerMgr           loggermgr.ILoggerManager      `inject:""` // 日志管理器
}

// NewPageAdminController 创建控制器实例
func NewPageAdminController() IPageAdminController {
	return &pageAdminControllerImpl{}
}

// ControllerName 返回控制器名称
func (c *pageAdminControllerImpl) ControllerName() string {
	return "pageAdminControllerImpl"
}

// GetRouter 返回路由信息
func (c *pageAdminControllerImpl) GetRouter() string {
	return "/admin [GET]"
}

// Handle 处理管理页面请求，渲染 admin.html 模板
func (c *pageAdminControllerImpl) Handle(ctx *gin.Context) {
	c.HTMLTemplateService.Render(ctx, "admin.html", gin.H{
		"title": "留言管理",
	})
}

var _ IPageAdminController = (*pageAdminControllerImpl)(nil)
