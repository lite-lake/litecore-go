// Package services 定义业务逻辑层
package services

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/manager/loggermgr"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/liteservice"
)

// IHTMLTemplateService HTML 模板服务接口
type IHTMLTemplateService interface {
	common.IBaseService
	Render(ctx *gin.Context, name string, data interface{}) // 渲染 HTML 模板
	SetGinEngine(engine *gin.Engine)                        // 设置 Gin 引擎
}

type htmlTemplateServiceImpl struct {
	inner     *liteservice.HTMLTemplateService // 内置模板服务
	LoggerMgr loggermgr.ILoggerManager         `inject:""` // 日志管理器
}

// NewHTMLTemplateService 创建 HTML 模板服务实例
func NewHTMLTemplateService() IHTMLTemplateService {
	return &htmlTemplateServiceImpl{
		inner: liteservice.NewHTMLTemplateService("templates/*"),
	}
}

// ServiceName 返回服务名称
func (s *htmlTemplateServiceImpl) ServiceName() string {
	return "HTMLTemplateService"
}

// OnStart 启动时初始化模板引擎
func (s *htmlTemplateServiceImpl) OnStart() error {
	return s.inner.OnStart()
}

// OnStop 停止时清理
func (s *htmlTemplateServiceImpl) OnStop() error {
	return s.inner.OnStop()
}

// Render 渲染指定的 HTML 模板
func (s *htmlTemplateServiceImpl) Render(ctx *gin.Context, name string, data interface{}) {
	s.inner.Render(ctx, name, data)
}

// SetGinEngine 设置 Gin 引擎用于模板渲染
func (s *htmlTemplateServiceImpl) SetGinEngine(engine *gin.Engine) {
	s.inner.SetGinEngine(engine)
}

var _ IHTMLTemplateService = (*htmlTemplateServiceImpl)(nil)
