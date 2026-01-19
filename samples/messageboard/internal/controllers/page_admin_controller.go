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

type PageAdminController struct {
	HTMLTemplateService services.IHTMLTemplateService `inject:""`
}

func NewPageAdminController() IPageAdminController {
	return &PageAdminController{}
}

func (c *PageAdminController) ControllerName() string {
	return "PageAdminController"
}

func (c *PageAdminController) GetRouter() string {
	return "/admin [GET]"
}

func (c *PageAdminController) Handle(ctx *gin.Context) {
	c.HTMLTemplateService.Render(ctx, "admin.html", gin.H{
		"title": "留言管理",
	})
}

var _ IPageAdminController = (*PageAdminController)(nil)
