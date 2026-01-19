package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofCmdlineController pprof 命令行参数控制器
type IPprofCmdlineController interface {
	common.BaseController
}

type PprofCmdlineController struct{}

func NewPprofCmdlineController() IPprofCmdlineController {
	return &PprofCmdlineController{}
}

func (c *PprofCmdlineController) ControllerName() string {
	return "PprofCmdlineController"
}

func (c *PprofCmdlineController) GetRouter() string {
	return "/debug/pprof/cmdline [GET]"
}

func (c *PprofCmdlineController) Handle(ctx *gin.Context) {
	pprof.Cmdline(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofCmdlineController)(nil)
