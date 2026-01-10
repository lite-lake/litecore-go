package common

import (
	"github.com/gin-gonic/gin"
)

// BaseMiddleware 基础中间件接口
// 所有 Middleware 类必须继承此接口并实现相关方法
// 用于定义基础中间件的规范，包括名称、执行顺序和包装函数。
type BaseMiddleware interface {
	// GetMiddlewareName 返回中间件的名称
	GetMiddlewareName() string
	// Order 返回中间件的执行顺序
	Order() int
	// Wrapper 返回一个中间件函数，用于包装请求处理函数
	Wrapper() gin.HandlerFunc
}
