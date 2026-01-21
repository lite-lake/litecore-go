package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IPageHomeController 首页控制器接口
type IPageHomeController interface {
	common.IBaseController
}

type pageHomeControllerImpl struct {
	HTMLTemplateService services.IHTMLTemplateService `inject:""`
	LoggerMgr           loggermgr.ILoggerManager      `inject:""`
	logger              loggermgr.ILogger
}

func NewPageHomeController() IPageHomeController {
	return &pageHomeControllerImpl{}
}

func (c *pageHomeControllerImpl) ControllerName() string {
	return "pageHomeControllerImpl"
}

func (c *pageHomeControllerImpl) GetRouter() string {
	return "/ [GET]"
}

func (c *pageHomeControllerImpl) Logger() loggermgr.ILogger {
	return c.logger
}

func (c *pageHomeControllerImpl) SetLoggerManager(mgr loggermgr.ILoggerManager) {
	c.LoggerMgr = mgr
	c.initLogger()
}

func (c *pageHomeControllerImpl) initLogger() {
	if c.LoggerMgr != nil {
		c.logger = c.LoggerMgr.Logger("PageHomeController")
	}
}

func (c *pageHomeControllerImpl) Handle(ctx *gin.Context) {
	c.HTMLTemplateService.Render(ctx, "index.html", gin.H{
		"title": "留言板",
	})
}

var _ IPageHomeController = (*pageHomeControllerImpl)(nil)
