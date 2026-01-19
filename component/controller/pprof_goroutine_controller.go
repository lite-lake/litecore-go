package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofGoroutineController pprof goroutine 控制器
type IPprofGoroutineController interface {
	common.BaseController
}

type PprofGoroutineController struct{}

func NewPprofGoroutineController() IPprofGoroutineController {
	return &PprofGoroutineController{}
}

func (c *PprofGoroutineController) ControllerName() string {
	return "PprofGoroutineController"
}

func (c *PprofGoroutineController) GetRouter() string {
	return "/debug/pprof/goroutine [GET]"
}

func (c *PprofGoroutineController) Handle(ctx *gin.Context) {
	pprof.Handler("goroutine").ServeHTTP(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofGoroutineController)(nil)
