// Package main 是留言板应用的入口点
package main

import (
	"log"

	messageboardapp "github.com/lite-lake/litecore-go/samples/messageboard/internal/application"
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

	// 启动引擎（启动所有 Manager 和 HTTP 服务器）
	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}

	// 等待关闭信号
	engine.WaitForShutdown()

	log.Println("Server exited")
}
