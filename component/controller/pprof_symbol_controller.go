package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofSymbolController pprof 符号表控制器
type IPprofSymbolController interface {
	common.BaseController
}

type PprofSymbolController struct{}

func NewPprofSymbolController() IPprofSymbolController {
	return &PprofSymbolController{}
}

func (c *PprofSymbolController) ControllerName() string {
	return "PprofSymbolController"
}

func (c *PprofSymbolController) GetRouter() string {
	return "/debug/pprof/symbol [GET]"
}

func (c *PprofSymbolController) Handle(ctx *gin.Context) {
	pprof.Symbol(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofSymbolController)(nil)
