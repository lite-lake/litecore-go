// Package services 定义业务逻辑层
package services

import (
	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	"com.litelake.litecore/component/service"
)

// IHTMLTemplateService HTML模板服务接口
type IHTMLTemplateService interface {
	common.BaseService
	Render(ctx *gin.Context, name string, data interface{})
}

type htmlTemplateService struct {
	inner *service.HTMLTemplateService
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
	return s.inner.OnStart()
}

func (s *htmlTemplateService) OnStop() error {
	return s.inner.OnStop()
}

func (s *htmlTemplateService) Render(ctx *gin.Context, name string, data interface{}) {
	s.inner.Render(ctx, name, data)
}

func (s *htmlTemplateService) SetGinEngine(engine *gin.Engine) {
	s.inner.SetGinEngine(engine)
}

var _ IHTMLTemplateService = (*htmlTemplateService)(nil)
