package server

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// registerRoute 注册自定义路由
func (e *Engine) registerRoute(method, path string, handler gin.HandlerFunc) {
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
