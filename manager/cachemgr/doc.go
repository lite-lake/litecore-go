// Package cachemgr 提供统一的缓存管理功能，支持 Redis、内存和空缓存三种驱动。
//
// 核心特性：
//   - 多驱动支持：支持 Redis（分布式）、Memory（高性能内存）、None（降级）三种缓存驱动
//   - 统一接口：提供统一的 ICacheManager 接口，便于切换缓存实现
//   - 可观测性：内置日志、指标和链路追踪支持
//   - 连接池管理：Redis 驱动支持连接池配置和自动管理
//   - 批量操作：支持批量获取、设置和删除操作
//   - 原子操作：支持 SetNX、Increment、Decrement 等原子操作
//
// 基本用法：
//
//	// 使用内存缓存
//	mgr := cachemgr.NewCacheManagerMemoryImpl(10*time.Minute, 5*time.Minute)
//	defer mgr.Close()
//
//	ctx := context.Background()
//
//	// 设置缓存
//	err := mgr.Set(ctx, "user:123", userData, 5*time.Minute)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 获取缓存
//	var data User
//	err = mgr.Get(ctx, "user:123", &data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 使用 Redis 缓存：
//
//	cfg := &cachemgr.RedisConfig{
//	    Host:     "localhost",
//	    Port:     6379,
//	    Password: "",
//	    DB:       0,
//	}
//	mgr, err := cachemgr.NewCacheManagerRedisImpl(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mgr.Close()
//
// 配置驱动类型：
//
//	// 通过 Build 函数创建
//	mgr, err := cachemgr.Build("memory", map[string]any{
//	    "max_age": "1h",
//	})
//
//	// 通过配置提供者创建
//	mgr, err := cachemgr.BuildWithConfigProvider(configProvider)
//
// 过期时间管理：
//
//	// 设置过期时间
//	mgr.Expire(ctx, "key", 10*time.Minute)
//
//	// 查看剩余时间
//	ttl, err := mgr.TTL(ctx, "key")
//
// 分布式锁（使用 SetNX）：
//
//	// 尝试获取锁
//	locked, err := mgr.SetNX(ctx, "lock:resource", "owner", 10*time.Second)
//	if locked {
//	    // 执行需要加锁的操作
//	    // ...
//
//	    // 释放锁
//	    mgr.Delete(ctx, "lock:resource")
//	}
package cachemgr
