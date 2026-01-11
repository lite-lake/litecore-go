// Package config 提供配置管理功能，支持 JSON 和 YAML 格式。
//
// 核心特性：
//   - 多格式支持：JSON、YAML 配置文件
//   - 路径查询：支持点分隔路径（如 "database.host"）和数组索引（如 "servers[0].port"）
//   - 类型安全：泛型 Get 函数提供编译时类型检查
//   - 智能转换：自动处理 JSON 数字类型的 float64 到 int 的转换
//   - 线程安全：配置数据不可变，可安全地在多个 goroutine 间共享
//
// 基本用法：
//
//	// 创建配置提供者
//	provider, err := config.NewConfigProvider("yaml", "./config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 获取配置值
//	host := config.Get[string](provider, "database.host")
//	port := config.Get[int](provider, "database.port")
//
//	// 使用默认值
//	timeout := config.GetWithDefault(provider, "server.timeout", 30)
//
// 路径语法：
//
//	// 嵌套对象
//	provider.Get("database.connection.host")
//
//	// 数组元素
//	provider.Get("servers[0].port")
//	provider.Get("items[2]")
package config
