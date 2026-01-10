// Package cachemgr 提供 LiteCore 框架的缓存管理功能
//
// 缓存管理器（CacheManager）是 LiteCore 框架的核心组件之一，负责缓存的创建、管理和生命周期控制。
// 支持多种缓存驱动：Redis、Memory 和 None（降级方案），提供统一的缓存操作接口。
//
// 核心接口：
//   - CacheManager: 缓存管理器接口，提供所有缓存操作方法
//
// 支持的驱动：
//   - Redis: 使用 go-redis v9 客户端连接 Redis 服务器
//   - Memory: 使用 go-cache 库实现内存缓存
//   - None: 空实现，用于降级场景
//
// 基本用法：
//
//	// 创建缓存管理器
//	cfg := loadConfig() // 从配置加载
//	mgr := cachemgr.Build(cfg["cache"].(map[string]any))
//
//	// 基本操作
//	ctx := context.Background()
//	mgr.Set(ctx, "key", "value", 10*time.Minute)
//	var value string
//	mgr.Get(ctx, "key", &value)
//
// 配置示例：
//
//	cache:
//	  driver: redis
//	  redis_config:
//	    host: localhost
//	    port: 6379
//	    password: ""
//	    db: 0
package cachemgr
