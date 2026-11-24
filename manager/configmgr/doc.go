// Package configmgr 提供配置管理功能，支持 JSON 和 YAML 格式。
//
// 核心特性：
//   - 多格式支持：支持 JSON 和 YAML 两种常见配置格式
//   - 路径查询：支持点分隔路径语法（如 server.host）和数组索引（如 servers[0].port）
//   - 类型安全：提供泛型 Get 函数，支持自动类型转换
//   - 线程安全：配置数据在加载后不可变，可安全地在多个 goroutine 之间共享
//
// 基本用法：
//
//	// 创建配置管理器
//	mgr, err := configmgr.NewConfigManager("yaml", "config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 获取配置值
//	name := configmgr.Ins[string](mgr, "app.name")
//	port := configmgr.Ins[int](mgr, "server.port")
//
//	// 带默认值的获取
//	timeout := configmgr.GetWithDefault(mgr, "server.timeout", 30)
//
// 路径语法：
//
// 配置路径使用点（.）分隔各层键名，支持数组索引语法 [n]：
//   - 简单键："port"
//   - 嵌套路径："server.host"
//   - 数组元素："servers[0]"
//   - 嵌套数组："items[0].name"
//
// 类型转换：
//
// Get 函数支持智能类型转换：
//   - JSON 中的 float64 自动转换为 int/int64（如果值是整数）
//   - 字符串 "true"/"false" 自动转换为 bool
//   - 数字字符串自动转换为数值类型
package configmgr
