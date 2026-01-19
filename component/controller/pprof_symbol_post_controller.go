package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofSymbolPostController pprof 符号表 POST 控制器
type IPprofSymbolPostController interface {
	common.IBaseController
}

type PprofSymbolPostController struct{}

func NewPprofSymbolPostController() IPprofSymbolPostController {
	return &PprofSymbolPostController{}
}

func (c *PprofSymbolPostController) ControllerName() string {
	return "PprofSymbolPostController"
}

func (c *PprofSymbolPostController) GetRouter() string {
	return "/debug/pprof/symbol [POST]"
}

func (c *PprofSymbolPostController) Handle(ctx *gin.Context) {
	pprof.Symbol(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.IBaseController = (*PprofSymbolPostController)(nil)
