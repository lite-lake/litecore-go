package liteservice

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/common/cachebust"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
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
	LoggerMgr loggermgr.ILoggerManager `inject:""`
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
	matches, err := filepath.Glob(s.config.TemplatePath)
	if err != nil {
		return fmt.Errorf("匹配模板文件失败: %w", err)
	}
	if len(matches) == 0 {
		s.LoggerMgr.Ins().Warn("未找到模板文件")
		return nil
	}
	// 解析期重写：把 /static/ 资源引用烘焙为版本化路径，主资源直连免 302
	tmpl := template.New("")
	for _, path := range matches {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取模板文件失败 %s: %w", path, err)
		}
		content = cachebust.RewriteHTML(content)
		name := filepath.Base(path)
		if _, err := tmpl.New(name).Parse(string(content)); err != nil {
			return fmt.Errorf("解析模板失败 %s: %w", path, err)
		}
	}
	s.ginEngine.SetHTMLTemplate(tmpl)
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
	ctx.Header("Cache-Control", cachebust.HeaderHTMLNoCache)
	ctx.HTML(common.HTTPStatusOK, name, data)
}

// GetConfig 获取HTML模板配置
func (s *HTMLTemplateService) GetConfig() *HTMLTemplateConfig {
	return s.config
}

var _ IHTMLTemplateService = (*HTMLTemplateService)(nil)
