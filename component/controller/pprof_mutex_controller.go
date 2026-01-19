package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofMutexController pprof 互斥锁控制器
type IPprofMutexController interface {
	common.BaseController
}

type PprofMutexController struct{}

func NewPprofMutexController() IPprofMutexController {
	return &PprofMutexController{}
}

func (c *PprofMutexController) ControllerName() string {
	return "PprofMutexController"
}

func (c *PprofMutexController) GetRouter() string {
	return "/debug/pprof/mutex [GET]"
}

func (c *PprofMutexController) Handle(ctx *gin.Context) {
	pprof.Handler("mutex").ServeHTTP(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofMutexController)(nil)
