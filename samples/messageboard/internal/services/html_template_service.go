// Package services 定义业务逻辑层
package services

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/loggermgr"
	"github.com/lite-lake/litecore-go/component/service"
)

// IHTMLTemplateService HTML模板服务接口
type IHTMLTemplateService interface {
	common.IBaseService
	Render(ctx *gin.Context, name string, data interface{})
	SetGinEngine(engine *gin.Engine)
}

type htmlTemplateService struct {
	inner     *service.HTMLTemplateService
	LoggerMgr loggermgr.ILoggerManager `inject:""`
	logger    loggermgr.ILogger
}

// NewHTMLTemplateService 创建HTML模板服务
func NewHTMLTemplateService() IHTMLTemplateService {
	return &htmlTemplateService{
		inner: service.NewHTMLTemplateService("templates/*"),
	}
}

func (s *htmlTemplateService) ServiceName() string {
	return "HTMLTemplateService"
}

func (s *htmlTemplateService) OnStart() error {
	s.initLogger()
	return s.inner.OnStart()
}

func (s *htmlTemplateService) OnStop() error {
	return s.inner.OnStop()
}

func (s *htmlTemplateService) Logger() loggermgr.ILogger {
	return s.logger
}

func (s *htmlTemplateService) SetLoggerManager(mgr loggermgr.ILoggerManager) {
	s.LoggerMgr = mgr
	s.initLogger()
}

func (s *htmlTemplateService) initLogger() {
	if s.LoggerMgr != nil {
		s.logger = s.LoggerMgr.Logger("HTMLTemplateService")
	}
}

func (s *htmlTemplateService) Render(ctx *gin.Context, name string, data interface{}) {
	s.inner.Render(ctx, name, data)
}

func (s *htmlTemplateService) SetGinEngine(engine *gin.Engine) {
	s.inner.SetGinEngine(engine)
}

var _ IHTMLTemplateService = (*htmlTemplateService)(nil)
