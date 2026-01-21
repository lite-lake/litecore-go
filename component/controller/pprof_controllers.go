package controller

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
)

type pprofHandlerFunc func(http.ResponseWriter, *http.Request)

type PprofController struct {
	name   string
	route  string
	method string
	handle pprofHandlerFunc
}

func (c *PprofController) ControllerName() string {
	return c.name
}

func (c *PprofController) GetRouter() string {
	return c.route
}

func (c *PprofController) Handle(ctx *gin.Context) {
	c.handle(wrapResponseWriter(ctx.Writer), ctx.Request)
}

var _ common.IBaseController = (*PprofController)(nil)

type IPprofIndexController interface {
	common.IBaseController
}

func NewPprofIndexController() IPprofIndexController {
	return &PprofController{
		name:   "PprofIndexController",
		route:  "/debug/pprof [GET]",
		method: "GET",
		handle: pprof.Index,
	}
}

type IPprofHeapController interface {
	common.IBaseController
}

func NewPprofHeapController() IPprofHeapController {
	return &PprofController{
		name:   "PprofHeapController",
		route:  "/debug/pprof/heap [GET]",
		method: "GET",
		handle: func(w http.ResponseWriter, r *http.Request) {
			pprof.Handler("heap").ServeHTTP(w, r)
		},
	}
}

type IPprofGoroutineController interface {
	common.IBaseController
}

func NewPprofGoroutineController() IPprofGoroutineController {
	return &PprofController{
		name:   "PprofGoroutineController",
		route:  "/debug/pprof/goroutine [GET]",
		method: "GET",
		handle: func(w http.ResponseWriter, r *http.Request) {
			pprof.Handler("goroutine").ServeHTTP(w, r)
		},
	}
}

type IPprofAllocsController interface {
	common.IBaseController
}

func NewPprofAllocsController() IPprofAllocsController {
	return &PprofController{
		name:   "PprofAllocsController",
		route:  "/debug/pprof/allocs [GET]",
		method: "GET",
		handle: func(w http.ResponseWriter, r *http.Request) {
			pprof.Handler("allocs").ServeHTTP(w, r)
		},
	}
}

type IPprofBlockController interface {
	common.IBaseController
}

func NewPprofBlockController() IPprofBlockController {
	return &PprofController{
		name:   "PprofBlockController",
		route:  "/debug/pprof/block [GET]",
		method: "GET",
		handle: func(w http.ResponseWriter, r *http.Request) {
			pprof.Handler("block").ServeHTTP(w, r)
		},
	}
}

type IPprofMutexController interface {
	common.IBaseController
}

func NewPprofMutexController() IPprofMutexController {
	return &PprofController{
		name:   "PprofMutexController",
		route:  "/debug/pprof/mutex [GET]",
		method: "GET",
		handle: func(w http.ResponseWriter, r *http.Request) {
			pprof.Handler("mutex").ServeHTTP(w, r)
		},
	}
}

type IPprofProfileController interface {
	common.IBaseController
}

func NewPprofProfileController() IPprofProfileController {
	return &PprofController{
		name:   "PprofProfileController",
		route:  "/debug/pprof/profile [GET]",
		method: "GET",
		handle: pprof.Profile,
	}
}

type IPprofTraceController interface {
	common.IBaseController
}

func NewPprofTraceController() IPprofTraceController {
	return &PprofController{
		name:   "PprofTraceController",
		route:  "/debug/pprof/trace [GET]",
		method: "GET",
		handle: pprof.Trace,
	}
}

type IPprofSymbolController interface {
	common.IBaseController
}

func NewPprofSymbolController() IPprofSymbolController {
	return &PprofController{
		name:   "PprofSymbolController",
		route:  "/debug/pprof/symbol [GET]",
		method: "GET",
		handle: pprof.Symbol,
	}
}

type IPprofSymbolPostController interface {
	common.IBaseController
}

func NewPprofSymbolPostController() IPprofSymbolPostController {
	return &PprofController{
		name:   "PprofSymbolPostController",
		route:  "/debug/pprof/symbol [POST]",
		method: "POST",
		handle: pprof.Symbol,
	}
}

type IPprofCmdlineController interface {
	common.IBaseController
}

func NewPprofCmdlineController() IPprofCmdlineController {
	return &PprofController{
		name:   "PprofCmdlineController",
		route:  "/debug/pprof/cmdline [GET]",
		method: "GET",
		handle: pprof.Cmdline,
	}
}

type IPprofThreadcreateController interface {
	common.IBaseController
}

func NewPprofThreadcreateController() IPprofThreadcreateController {
	return &PprofController{
		name:   "PprofThreadcreateController",
		route:  "/debug/pprof/threadcreate [GET]",
		method: "GET",
		handle: func(w http.ResponseWriter, r *http.Request) {
			pprof.Handler("threadcreate").ServeHTTP(w, r)
		},
	}
}
