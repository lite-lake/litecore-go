package service

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// HTMLTemplateConfig HTML模板配置
type HTMLTemplateConfig struct {
	TemplatePath string // 模板文件路径模式，如 templates/*
}

// IHTMLTemplateService HTML模板服务接口
type IHTMLTemplateService interface {
	common.IBaseService
	// Render 渲染HTML模板
	Render(ctx *gin.Context, name string, data interface{})
}

// HTMLTemplateService HTML模板服务
// 用于处理HTML模板渲染
type HTMLTemplateService struct {
	config    *HTMLTemplateConfig
	ginEngine *gin.Engine
}

// NewHTMLTemplateService 创建HTML模板服务
func NewHTMLTemplateService(templatePath string) *HTMLTemplateService {
	return &HTMLTemplateService{
		config: &HTMLTemplateConfig{
			TemplatePath: templatePath,
		},
	}
}

// ServiceName 服务名称
func (s *HTMLTemplateService) ServiceName() string {
	return "HTMLTemplateService"
}

// OnStart 启动服务，加载HTML模板
func (s *HTMLTemplateService) OnStart() error {
	if s.ginEngine == nil {
		return nil
	}
	s.ginEngine.LoadHTMLGlob(s.config.TemplatePath)
	return nil
}

// OnStop 停止服务
func (s *HTMLTemplateService) OnStop() error {
	s.ginEngine = nil
	return nil
}

// SetGinEngine 设置Gin引擎
func (s *HTMLTemplateService) SetGinEngine(engine *gin.Engine) {
	s.ginEngine = engine
}

// Render 渲染HTML模板
func (s *HTMLTemplateService) Render(ctx *gin.Context, name string, data interface{}) {
	if s.ginEngine == nil {
		ctx.JSON(common.HTTPStatusInternalServerError, gin.H{"error": "HTML templates not loaded"})
		return
	}
	ctx.HTML(common.HTTPStatusOK, name, data)
}

// GetConfig 获取HTML模板配置
func (s *HTMLTemplateService) GetConfig() *HTMLTemplateConfig {
	return s.config
}

var _ IHTMLTemplateService = (*HTMLTemplateService)(nil)
