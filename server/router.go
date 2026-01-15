package server

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// registerControllers 注册所有控制器
func (e *Engine) registerControllers() {
	controllers := e.Controller.GetAll()

	for _, ctrl := range controllers {
		route, method := parseRouter(ctrl.GetRouter())

		// 注册到 Gin
		switch method {
		case "GET":
			e.ginEngine.GET(route, ctrl.Handle)
		case "POST":
			e.ginEngine.POST(route, ctrl.Handle)
		case "PUT":
			e.ginEngine.PUT(route, ctrl.Handle)
		case "DELETE":
			e.ginEngine.DELETE(route, ctrl.Handle)
		case "PATCH":
			e.ginEngine.PATCH(route, ctrl.Handle)
		case "HEAD":
			e.ginEngine.HEAD(route, ctrl.Handle)
		case "OPTIONS":
			e.ginEngine.OPTIONS(route, ctrl.Handle)
		default:
			// 默认 GET
			e.ginEngine.GET(route, ctrl.Handle)
		}
	}
}

// parseRouter 解析路由字符串
// 支持格式：
// - "/api/users" (默认 GET)
// - "/api/users [GET]"
// - "/api/users [get]"
// - "/api/users GET"
func parseRouter(router string) (route string, method string) {
	router = strings.TrimSpace(router)

	// 尝试解析带方括号的格式: "/path [METHOD]"
	if strings.Contains(router, "[") && strings.Contains(router, "]") {
		parts := strings.SplitN(router, "[", 2)
		route = strings.TrimSpace(parts[0])
		methodPart := strings.TrimSuffix(strings.TrimSpace(parts[1]), "]")
		method = strings.ToUpper(methodPart)
		return route, method
	}

	// 尝试解析空格分隔的格式: "/path METHOD"
	parts := strings.Fields(router)
	if len(parts) >= 2 {
		route = parts[0]
		method = strings.ToUpper(parts[len(parts)-1])
		// 验证是否为有效的 HTTP 方法
		switch method {
		case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
			return route, method
		}
		// 无效方法，返回路由部分和默认 GET 方法
		return route, "GET"
	}

	// 默认 GET
	return router, "GET"
}

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
