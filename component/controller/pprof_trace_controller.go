package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofTraceController pprof goroutine trace 控制器
type IPprofTraceController interface {
	common.BaseController
}

type PprofTraceController struct{}

func NewPprofTraceController() IPprofTraceController {
	return &PprofTraceController{}
}

func (c *PprofTraceController) ControllerName() string {
	return "PprofTraceController"
}

func (c *PprofTraceController) GetRouter() string {
	return "/debug/pprof/trace [GET]"
}

func (c *PprofTraceController) Handle(ctx *gin.Context) {
	pprof.Trace(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofTraceController)(nil)
