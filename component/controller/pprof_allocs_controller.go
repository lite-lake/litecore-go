package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofAllocsController pprof 内存分配控制器
type IPprofAllocsController interface {
	common.BaseController
}

type PprofAllocsController struct{}

func NewPprofAllocsController() IPprofAllocsController {
	return &PprofAllocsController{}
}

func (c *PprofAllocsController) ControllerName() string {
	return "PprofAllocsController"
}

func (c *PprofAllocsController) GetRouter() string {
	return "/debug/pprof/allocs [GET]"
}

func (c *PprofAllocsController) Handle(ctx *gin.Context) {
	pprof.Handler("allocs").ServeHTTP(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofAllocsController)(nil)
