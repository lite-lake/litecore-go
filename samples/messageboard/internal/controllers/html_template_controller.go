package controllers

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	componentControllers "com.litelake.litecore/component/controller"
)

// IHTMLTemplateController HTML模板控制器接口
type IHTMLTemplateController interface {
	common.BaseController
	InitializeTemplates(engine *gin.Engine)
}

type HTMLTemplateController struct {
	componentController *componentControllers.HTMLTemplateController
}

func NewHTMLTemplateController() IHTMLTemplateController {
	return &HTMLTemplateController{
		componentController: componentControllers.NewHTMLTemplateController("templates/*"),
	}
}

func (c *HTMLTemplateController) ControllerName() string {
	return "HTMLTemplateController"
}

func (c *HTMLTemplateController) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *HTMLTemplateController) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

func (c *HTMLTemplateController) InitializeTemplates(engine *gin.Engine) {
	c.componentController.LoadTemplates(engine)
}

var _ IHTMLTemplateController = (*HTMLTemplateController)(nil)
