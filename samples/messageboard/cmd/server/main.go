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

	// 初始化引擎（注册路由、依赖注入等）
	if err := engine.Initialize(); err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// 初始化 HTML 模板服务
	initializeHTMLTemplateService(engine)

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

// initializeHTMLTemplateService 初始化 HTML 模板服务
func initializeHTMLTemplateService(engine *server.Engine) {
	services := engine.GetServices()
	for _, svc := range services {
		if templateSvc, ok := svc.(interface {
			SetGinEngine(*gin.Engine)
		}); ok {
			templateSvc.SetGinEngine(engine.GetGinEngine())
			log.Printf("[DEBUG] HTML templates initialized for %s", svc.ServiceName())
		}
	}
}
