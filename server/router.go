package server

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterRoute 注册自定义路由
func (e *Engine) RegisterRoute(method, path string, handler gin.HandlerFunc) {
	switch strings.ToUpper(method) {
	case "GET":
		e.ginEngine.GET(path, handler)
	case "POST":
		e.ginEngine.POST(path, handler)
	case "PUT":
		e.ginEngine.PUT(path, handler)
	case "DELETE":
		e.ginEngine.DELETE(path, handler)
	case "PATCH":
		e.ginEngine.PATCH(path, handler)
	case "HEAD":
		e.ginEngine.HEAD(path, handler)
	case "OPTIONS":
		e.ginEngine.OPTIONS(path, handler)
	default:
		e.ginEngine.GET(path, handler)
	}
}

// RegisterGroup 注册路由组
func (e *Engine) RegisterGroup(groupPath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return e.ginEngine.Group(groupPath, handlers...)
}

// GetRouteInfo 获取所有已注册的路由信息
func (e *Engine) GetRouteInfo() []RouteInfo {
	routes := e.ginEngine.Routes()
	info := make([]RouteInfo, 0, len(routes))

	for _, route := range routes {
		info = append(info, RouteInfo{
			Method:  route.Method,
			Path:    route.Path,
			Handler: route.Handler,
		})
	}

	return info
}

// RouteInfo 路由信息
type RouteInfo struct {
	Method  string
	Path    string
	Handler string
}
