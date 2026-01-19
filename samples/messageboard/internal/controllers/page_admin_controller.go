package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/services"
)

// IPageAdminController 管理页面控制器接口
type IPageAdminController interface {
	common.BaseController
}

type pageAdminControllerImpl struct {
	HTMLTemplateService services.IHTMLTemplateService `inject:""`
}

func NewPageAdminController() IPageAdminController {
	return &pageAdminControllerImpl{}
}

func (c *pageAdminControllerImpl) ControllerName() string {
	return "pageAdminControllerImpl"
}

func (c *pageAdminControllerImpl) GetRouter() string {
	return "/admin [GET]"
}

func (c *pageAdminControllerImpl) Handle(ctx *gin.Context) {
	c.HTMLTemplateService.Render(ctx, "admin.html", gin.H{
		"title": "留言管理",
	})
}

var _ IPageAdminController = (*pageAdminControllerImpl)(nil)
