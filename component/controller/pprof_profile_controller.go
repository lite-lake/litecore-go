package controller

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofProfileController pprof CPU profile 控制器
type IPprofProfileController interface {
	common.BaseController
}

type PprofProfileController struct{}

func NewPprofProfileController() IPprofProfileController {
	return &PprofProfileController{}
}

func (c *PprofProfileController) ControllerName() string {
	return "PprofProfileController"
}

func (c *PprofProfileController) GetRouter() string {
	return "/debug/pprof/profile [GET]"
}

func (c *PprofProfileController) Handle(ctx *gin.Context) {
	pprof.Profile(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.BaseController = (*PprofProfileController)(nil)
