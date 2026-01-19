package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IAdminPageController 管理页面控制器接口
type IAdminPageController interface {
	common.BaseController
}

type AdminPageController struct {
	TemplateController componentControllers.HTMLTemplateController `inject:""`
}

func NewAdminPageController() IAdminPageController {
	return &AdminPageController{}
}

func (c *AdminPageController) ControllerName() string {
	return "AdminPageController"
}

func (c *AdminPageController) GetRouter() string {
	return "/admin [GET]"
}

func (c *AdminPageController) Handle(ctx *gin.Context) {
	c.TemplateController.Render(ctx, "admin.html", gin.H{
		"title": "留言管理",
	})
}

var _ IAdminPageController = (*AdminPageController)(nil)
