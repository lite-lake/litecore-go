package controller

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
)

// IPprofIndexController pprof 首页控制器
type IPprofIndexController interface {
	common.BaseController
}

type PprofIndexController struct{}

func NewPprofIndexController() IPprofIndexController {
	return &PprofIndexController{}
}

func (c *PprofIndexController) ControllerName() string {
	return "PprofIndexController"
}

func (c *PprofIndexController) GetRouter() string {
	return "/debug/pprof [GET]"
}

func (c *PprofIndexController) Handle(ctx *gin.Context) {
	pprof.Index(wrapResponseWriter(ctx.Writer), ctx.Request)
}

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

// IPprofSymbolPostController pprof 符号表 POST 控制器
type IPprofSymbolPostController interface {
	common.BaseController
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

// IPprofAllocsController pprof 内存分配控制器
type IPprofAllocsController interface {
	common.BaseController
}

type PprofAllocsController struct{}

func NewPprofAllocsController() IPprofAllocsController {
	return &PprofAllocsController{}
}

func (c *PprofAllocsController) ControllerName() string {
	return "PprofAllocsController"
}

func (c *PprofAllocsController) GetRouter() string {
	return "/debug/pprof/allocs [GET]"
}

func (c *PprofAllocsController) Handle(ctx *gin.Context) {
	pprof.Handler("allocs").ServeHTTP(wrapResponseWriter(ctx.Writer), ctx.Request)
}

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

// IPprofHeapController pprof 堆内存控制器
type IPprofHeapController interface {
	common.BaseController
}

type PprofHeapController struct{}

func NewPprofHeapController() IPprofHeapController {
	return &PprofHeapController{}
}

func (c *PprofHeapController) ControllerName() string {
	return "PprofHeapController"
}

func (c *PprofHeapController) GetRouter() string {
	return "/debug/pprof/heap [GET]"
}

func (c *PprofHeapController) Handle(ctx *gin.Context) {
	pprof.Handler("heap").ServeHTTP(wrapResponseWriter(ctx.Writer), ctx.Request)
}

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

// wrapResponseWriter 包装 gin.ResponseWriter 实现 http.ResponseWriter 接口
type responseWriterWrapper struct {
	gin.ResponseWriter
}

func wrapResponseWriter(w gin.ResponseWriter) http.ResponseWriter {
	return &responseWriterWrapper{ResponseWriter: w}
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

var (
	_ common.BaseController = (*PprofIndexController)(nil)
	_ common.BaseController = (*PprofCmdlineController)(nil)
	_ common.BaseController = (*PprofProfileController)(nil)
	_ common.BaseController = (*PprofSymbolController)(nil)
	_ common.BaseController = (*PprofSymbolPostController)(nil)
	_ common.BaseController = (*PprofTraceController)(nil)
	_ common.BaseController = (*PprofAllocsController)(nil)
	_ common.BaseController = (*PprofBlockController)(nil)
	_ common.BaseController = (*PprofGoroutineController)(nil)
	_ common.BaseController = (*PprofHeapController)(nil)
	_ common.BaseController = (*PprofMutexController)(nil)
	_ common.BaseController = (*PprofThreadcreateController)(nil)
)
