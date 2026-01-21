// Package config 提供配置管理功能，支持 JSON 和 YAML 格式的配置文件。
//
// 核心特性：
//   - 多格式支持：JSON、YAML 配置文件解析
//   - 路径查询：支持点分隔路径（如 "database.host"）和数组索引（如 "servers[0].port"）
//   - 类型安全：泛型 Get 函数提供编译时类型检查
//   - 智能转换：自动处理 JSON 数字类型的 float64 到 int 的转换
//   - 线程安全：配置数据不可变，可安全地在多个 goroutine 间共享
//
// 基本用法：
//
//	package main
//
//	import (
//	    loggermgr "github.com/lite-lake/litecore-go/component/manager/loggermgr"
//
//	    "github.com/lite-lake/litecore-go/config"
//	)
//
//	func main() {
//	    loggerMgr := loggermgr.GetLoggerManager()
//	    logger := loggerMgr.Logger("main")
//
//	    // 创建配置提供者
//	    provider, err := config.NewConfigProvider("yaml", "./config.yaml")
//	    if err != nil {
//	        logger.Fatal("创建配置提供者失败", "error", err)
//	    }
//
//	    // 获取配置值（类型安全）
//	    host, err := config.Get[string](provider, "database.host")
//	    if err != nil {
//	        logger.Fatal("获取配置失败", "error", err)
//	    }
//
//	    port, err := config.Get[int](provider, "database.port")
//	    if err != nil {
//	        logger.Fatal("获取配置失败", "error", err)
//	    }
//
//	    // 使用默认值
//	    timeout := config.GetWithDefault(provider, "server.timeout", 30)
//
//	    // 检查键是否存在
//	    if provider.Has("feature.enabled") {
//	        // 功能已启用
//	    }
//	}
//
// 路径语法：
//
// 配置路径使用点号分隔对象层级，使用方括号访问数组元素：
//
//	// 嵌套对象访问
//	provider.Get("database.connection.host")  // 访问 database.connection.host
//
//	// 数组索引访问
//	provider.Get("servers[0].port")           // 访问 servers 数组第一个元素的 port
//	provider.Get("items[2]")                  // 访问 items 数组第三个元素
//
//	// 混合访问
//	provider.Get("app.servers[0].ports[0]")   // 嵌套对象+数组
//
// 错误处理：
//
// 包提供两个错误变量用于错误判断：
//
//	import "github.com/lite-lake/litecore-go/config"
//
//	val, err := config.Get[string](provider, "some.key")
//	if config.IsConfigKeyNotFound(err) {
//	    // 键不存在
//	} else if err != nil {
//	    // 其他错误（类型不匹配等）
//	    logger.Error("获取配置失败", "error", err)
//	}
//
// 线程安全：
//
// 配置提供者在创建后不可变，可以安全地在多个 goroutine 之间共享使用。
package config
