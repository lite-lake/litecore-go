package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofBlockController pprof 阻塞 profile 控制器
type IPprofBlockController interface {
	common.BaseController
}

type PprofBlockController struct{}

func NewPprofBlockController() IPprofBlockController {
	return &PprofBlockController{}
}

func (c *PprofBlockController) ControllerName() string {
	return "PprofBlockController"
}

func (c *PprofBlockController) GetRouter() string {
	return "/debug/pprof/block [GET]"
}

func (c *PprofBlockController) Handle(ctx *gin.Context) {
	pprof.Handler("block").ServeHTTP(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofBlockController)(nil)
