package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofHeapController pprof 堆内存控制器
type IPprofHeapController interface {
	common.IBaseController
}

type PprofHeapController struct{}

func NewPprofHeapController() IPprofHeapController {
	return &PprofHeapController{}
}

func (c *PprofHeapController) ControllerName() string {
	return "PprofHeapController"
}

func (c *PprofHeapController) GetRouter() string {
	return "/debug/pprof/heap [GET]"
}

func (c *PprofHeapController) Handle(ctx *gin.Context) {
	pprof.Handler("heap").ServeHTTP(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.IBaseController = (*PprofHeapController)(nil)
