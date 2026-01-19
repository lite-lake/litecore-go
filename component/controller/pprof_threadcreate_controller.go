package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofThreadcreateController pprof 线程创建控制器
type IPprofThreadcreateController interface {
	common.BaseController
}

type PprofThreadcreateController struct{}

func NewPprofThreadcreateController() IPprofThreadcreateController {
	return &PprofThreadcreateController{}
}

func (c *PprofThreadcreateController) ControllerName() string {
	return "PprofThreadcreateController"
}

func (c *PprofThreadcreateController) GetRouter() string {
	return "/debug/pprof/threadcreate [GET]"
}

func (c *PprofThreadcreateController) Handle(ctx *gin.Context) {
	pprof.Handler("threadcreate").ServeHTTP(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofThreadcreateController)(nil)
