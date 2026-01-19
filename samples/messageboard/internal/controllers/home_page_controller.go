package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IHomePageController 首页控制器接口
type IHomePageController interface {
	common.BaseController
}

type HomePageController struct {
	TemplateController componentControllers.HTMLTemplateController `inject:""`
}

func NewHomePageController() IHomePageController {
	return &HomePageController{}
}

func (c *HomePageController) ControllerName() string {
	return "HomePageController"
}

func (c *HomePageController) GetRouter() string {
	return "/ [GET]"
}

func (c *HomePageController) Handle(ctx *gin.Context) {
	c.TemplateController.Render(ctx, "index.html", gin.H{
		"title": "留言板",
	})
}

var _ IHomePageController = (*HomePageController)(nil)
