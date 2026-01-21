package common

import (
	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/util/logger"
)

// IBaseController 基础控制器接口
// 所有 Controller 类必须继承此接口并实现相关方法
// 用于定义基础控制器的规范，包括路由和处理函数。
type IBaseController interface {
	// ControllerName 返回当前控制器的类名
	ControllerName() string
	// GetRouter 返回当前控制器的路由
	// 路由格式同OpenAPI @Router 规范
	// 如 `/aaa/bbb [GET]`; `/aaa/bbb [POST]`
	GetRouter() string
	// Handle 处理当前控制器的请求
	Handle(ctx *gin.Context)

	// Logger 获取日志实例
	Logger() logger.ILogger

	// SetLoggerManager 设置日志管理器
	SetLoggerManager(mgr logger.ILoggerManager)
}
