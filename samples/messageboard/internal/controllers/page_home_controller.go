package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	"com.litelake.litecore/samples/messageboard/internal/services"
)

// IPageHomeController 首页控制器接口
type IPageHomeController interface {
	common.BaseController
}

type PageHomeController struct {
	HTMLTemplateService services.IHTMLTemplateService `inject:""`
}

func NewPageHomeController() IPageHomeController {
	return &PageHomeController{}
}

func (c *PageHomeController) ControllerName() string {
	return "PageHomeController"
}

func (c *PageHomeController) GetRouter() string {
	return "/ [GET]"
}

func (c *PageHomeController) Handle(ctx *gin.Context) {
	c.HTMLTemplateService.Render(ctx, "index.html", gin.H{
		"title": "留言板",
	})
}

var _ IPageHomeController = (*PageHomeController)(nil)
