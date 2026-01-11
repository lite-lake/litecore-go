// Package cachemgr 提供 LiteCore 框架的缓存管理功能。
//
// 核心特性：
//   - 多驱动支持：支持 Redis、Memory 和 None 三种缓存驱动
//   - 统一接口：提供一致的缓存操作 API，方便切换驱动
//   - 分布式功能：支持分布式锁、计数器等高级功能（Redis 驱动）
//   - 优雅降级：配置失败时自动降级到 None 驱动
//   - 生命周期管理：集成服务启停接口，支持健康检查
//
// 缓存管理器（CacheManager）是 LiteCore 框架的核心组件之一，负责缓存的创建、管理和生命周期控制。
//
// 核心接口：
//   - CacheManager: 缓存管理器接口，提供所有缓存操作方法
//
// 基本用法：
//
//	// 创建缓存管理器
//	cfg := map[string]any{
//	    "driver": "redis",
//	    "redis_config": map[string]any{
//	        "host":     "localhost",
//	        "port":     6379,
//	        "password": "",
//	        "db":       0,
//	    },
//	}
//	mgr := cachemgr.Build(cfg)
//
//	// 基本操作
//	ctx := context.Background()
//	mgr.Set(ctx, "key", "value", 10*time.Minute)
//	var value string
//	mgr.Get(ctx, "key", &value)
//
// 支持的驱动类型：
//
//   - "redis": Redis 缓存，使用 go-redis v9 客户端
//   - "memory": 内存缓存，使用 go-cache 库
//   - "none": 空实现，用于降级场景
package cachemgr
