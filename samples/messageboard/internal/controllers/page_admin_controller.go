package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/samples/messageboard/internal/services"
)

// IPageAdminController 管理页面控制器接口
type IPageAdminController interface {
	common.IBaseController
}

type pageAdminControllerImpl struct {
	HTMLTemplateService services.IHTMLTemplateService `inject:""`
	Logger              common.ILogger                `inject:""`
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
