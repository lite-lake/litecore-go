// Package litebloomsvc 提供布隆过滤器服务，支持高并发的元素存在性判断。
//
// 核心特性：
//   - 多过滤器管理：支持创建、获取、删除多个独立的布隆过滤器
//   - TTL 自动重建：支持配置过滤器生存时间，到期自动重建
//   - 并发安全：读写分离，支持高并发读写操作
//   - 统计信息：提供过滤器容量、元素数量、填充率等统计
//   - 依赖注入：通过 inject 标签注入 LoggerManager 组件
//
// 基本用法：
//
//	// 方式一：使用默认配置
//	service := litebloomsvc.NewService()
//
//	// 方式二：使用自定义配置
//	config := &litebloomsvc.Config{
//	    DefaultExpectedItems:    ptrUint(10000),
//	    DefaultFalsePositiveRate: ptrFloat64(0.01),
//	}
//	service := litebloomsvc.NewServiceWithConfig(config)
//
// 创建和管理过滤器：
//
//	// 创建默认配置的过滤器
//	filter, err := service.CreateFilter("user_ids")
//
//	// 创建自定义配置的过滤器
//	filterConfig := &litebloomsvc.FilterConfig{
//	    ExpectedItems:     ptrUint(100000),
//	    FalsePositiveRate: ptrFloat64(0.001),
//	}
//	filter, err := service.CreateFilterWithConfig("emails", filterConfig)
//
//	// 添加元素
//	service.Add("user_ids", []byte("user123"))
//	service.AddString("user_ids", "user456")
//
//	// 批量添加
//	service.AddBatch("user_ids", [][]byte{[]byte("a"), []byte("b")})
//
//	// 检查元素
//	exists := service.Contains("user_ids", []byte("user123"))
//	exists = service.ContainsString("user_ids", "user456")
//
//	// 获取统计信息
//	stats := service.GetStats("user_ids")
//	fmt.Printf("填充率: %.2f%%\n", stats.FillRatio*100)
//
// TTL 自动重建：
//
//	// 配置 TTL，过滤器到期后自动重建
//	config := &litebloomsvc.Config{
//	    DefaultTTL: ptrDuration(time.Hour),
//	}
//	service := litebloomsvc.NewServiceWithConfig(config)
//
// 性能建议：
//   - 预期元素数量：根据实际业务预估，宁大勿小
//   - 误判率：推荐 0.01 (1%) 或更低
//   - TTL：根据数据更新频率设置，频繁更新的数据设置较短 TTL
//   - 内存占用：与预期元素数量和误判率相关
package litebloomsvc
