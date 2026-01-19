// Package main 是留言板应用的入口点
package main

import (
	"log"

	messageboardapp "com.litelake.litecore/samples/messageboard/internal/application"
	"com.litelake.litecore/server"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建应用引擎
	engine, err := messageboardapp.NewEngine()
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	// 初始化 HTML 模板
	initializeHTMLTemplates(engine)

	// 初始化引擎（注册路由、依赖注入等）
	if err := engine.Initialize(); err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// 调试：打印所有路由
	ginEngine := engine.GetGinEngine()
	log.Printf("[DEBUG] Total routes: %d", len(ginEngine.Routes()))
	for _, route := range ginEngine.Routes() {
		log.Printf("[DEBUG] Route: %s %s", route.Method, route.Path)
	}

	// 启动引擎（启动所有 Manager 和 HTTP 服务器）
	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}

	// 等待关闭信号
	engine.WaitForShutdown()

	log.Println("Server exited")
}

// initializeHTMLTemplates 初始化 HTML 模板
func initializeHTMLTemplates(engine *server.Engine) {
	controllers := engine.GetControllers()
	for _, ctrl := range controllers {
		if templateCtrl, ok := ctrl.(interface {
			InitializeTemplates(*gin.Engine)
		}); ok {
			templateCtrl.InitializeTemplates(engine.GetGinEngine())
			log.Printf("[DEBUG] HTML templates initialized for %s", ctrl.ControllerName())
		}
	}
}
