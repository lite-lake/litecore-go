// Package main 是留言板应用的入口点
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"com.litelake.litecore/server"
	messageboardapp "com.litelake.litecore/samples/messageboard/internal/application"
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

	// 设置自定义路由（静态文件、HTML 模板）
	setupRoutes(engine)

	// 调试：打印所有路由
	ginEngine := engine.GetGinEngine()
	log.Printf("[DEBUG] Total routes after setupRoutes: %d", len(ginEngine.Routes()))
	for _, route := range ginEngine.Routes() {
		log.Printf("[DEBUG] Route: %s %s", route.Method, route.Path)
	}

	// 启动引擎（启动所有 Manager 和 HTTP 服务器）
	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 优雅关闭
	if err := engine.Stop(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

// setupRoutes 设置自定义路由
func setupRoutes(engine *server.Engine) {
	router := engine.GetGinEngine()

	// 静态文件服务
	router.Static("/static", "./static")

	// 加载 HTML 模板
	router.LoadHTMLGlob("templates/*")

	// 页面路由
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "留言板",
		})
	})

	router.GET("/admin.html", func(c *gin.Context) {
		c.HTML(200, "admin.html", gin.H{
			"title": "留言管理",
		})
	})
}
