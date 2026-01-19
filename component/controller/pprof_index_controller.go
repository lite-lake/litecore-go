package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofIndexController pprof 首页控制器
type IPprofIndexController interface {
	common.IBaseController
}

type PprofIndexController struct{}

func NewPprofIndexController() IPprofIndexController {
	return &PprofIndexController{}
}

func (c *PprofIndexController) ControllerName() string {
	return "PprofIndexController"
}

func (c *PprofIndexController) GetRouter() string {
	return "/debug/pprof [GET]"
}

func (c *PprofIndexController) Handle(ctx *gin.Context) {
	pprof.Index(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.IBaseController = (*PprofIndexController)(nil)
