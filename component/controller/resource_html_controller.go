package controller

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// ResourceHTMLConfig HTML模板配置
type ResourceHTMLConfig struct {
	TemplatePath string // 模板文件路径模式，如 templates/*
}

// ResourceHTMLController HTML模板控制器
// 用于处理HTML模板渲染
type ResourceHTMLController struct {
	config    *ResourceHTMLConfig
	ginEngine *gin.Engine
}

// NewResourceHTMLController 创建HTML模板控制器
func NewResourceHTMLController(templatePath string) *ResourceHTMLController {
	return &ResourceHTMLController{
		config: &ResourceHTMLConfig{
			TemplatePath: templatePath,
		},
	}
}

func (c *ResourceHTMLController) ControllerName() string {
	return "ResourceHTMLController"
}

func (c *ResourceHTMLController) GetRouter() string {
	return ""
}

func (c *ResourceHTMLController) Handle(ctx *gin.Context) {
	ctx.JSON(500, gin.H{"error": "ResourceHTMLController should not be registered as a route"})
}

// LoadTemplates 加载HTML模板
func (c *ResourceHTMLController) LoadTemplates(engine *gin.Engine) {
	c.ginEngine = engine
	engine.LoadHTMLGlob(c.config.TemplatePath)
}

// Render 渲染HTML模板
func (c *ResourceHTMLController) Render(ctx *gin.Context, name string, data interface{}) {
	if c.ginEngine == nil {
		ctx.JSON(500, gin.H{"error": "HTML templates not loaded"})
		return
	}
	ctx.HTML(200, name, data)
}

// GetConfig 获取HTML模板配置
func (c *ResourceHTMLController) GetConfig() *ResourceHTMLConfig {
	return c.config
}

var _ common.BaseController = (*ResourceHTMLController)(nil)
