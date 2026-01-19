package controller

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// HTMLTemplateConfig HTML模板配置
type HTMLTemplateConfig struct {
	TemplatePath string // 模板文件路径模式，如 templates/*
}

// HTMLTemplateController HTML模板控制器
// 用于处理HTML模板渲染
type HTMLTemplateController struct {
	config    *HTMLTemplateConfig
	ginEngine *gin.Engine
}

// NewHTMLTemplateController 创建HTML模板控制器
func NewHTMLTemplateController(templatePath string) *HTMLTemplateController {
	return &HTMLTemplateController{
		config: &HTMLTemplateConfig{
			TemplatePath: templatePath,
		},
	}
}

func (c *HTMLTemplateController) ControllerName() string {
	return "HTMLTemplateController"
}

func (c *HTMLTemplateController) GetRouter() string {
	return ""
}

func (c *HTMLTemplateController) Handle(ctx *gin.Context) {
	ctx.JSON(500, gin.H{"error": "HTMLTemplateController should not be registered as a route"})
}

// LoadTemplates 加载HTML模板
func (c *HTMLTemplateController) LoadTemplates(engine *gin.Engine) {
	c.ginEngine = engine
	engine.LoadHTMLGlob(c.config.TemplatePath)
}

// Render 渲染HTML模板
func (c *HTMLTemplateController) Render(ctx *gin.Context, name string, data interface{}) {
	if c.ginEngine == nil {
		ctx.JSON(500, gin.H{"error": "HTML templates not loaded"})
		return
	}
	ctx.HTML(200, name, data)
}

// GetConfig 获取HTML模板配置
func (c *HTMLTemplateController) GetConfig() *HTMLTemplateConfig {
	return c.config
}

var _ common.BaseController = (*HTMLTemplateController)(nil)
