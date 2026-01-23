// Package lockmgr 提供统一的锁管理功能，支持 Redis、内存和空锁三种驱动。
//
// 核心特性：
//   - 多驱动支持：支持 Redis（分布式锁）、Memory（本地内存锁）、None（无锁）三种锁驱动
//   - 统一接口：提供统一的 ILockManager 接口，便于切换锁实现
//   - 可观测性：内置日志、指标和链路追踪支持
//   - 自动过期：支持锁的自动过期机制，防止死锁
//   - 非阻塞模式：TryLock 支持非阻塞获取锁
//
// 基本用法：
//
//	// 使用内存锁
//	mgr := lockmgr.NewLockManagerMemoryImpl(&lockmgr.MemoryLockConfig{})
//	defer mgr.Close()
//
//	ctx := context.Background()
//
//	// 获取锁（阻塞）
//	err := mgr.Lock(ctx, "resource:123", 10*time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 执行需要加锁的操作
//	// ...
//
//	// 释放锁
//	mgr.Unlock(ctx, "resource:123")
//
// 使用 Redis 分布式锁：
//
//	cfg := &lockmgr.RedisLockConfig{
//	    Host:     "localhost",
//	    Port:     6379,
//	    Password: "",
//	    DB:       0,
//	}
//	mgr := lockmgr.NewLockManagerRedisImpl(cfg)
//	defer mgr.Close()
//
// 非阻塞获取锁：
//
//	// 尝试获取锁（非阻塞）
//	locked, err := mgr.TryLock(ctx, "resource:123", 10*time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if locked {
//	    // 成功获取锁
//	    // 执行需要加锁的操作
//	    // ...
//
//	    // 释放锁
//	    mgr.Unlock(ctx, "resource:123")
//	} else {
//	    // 锁已被占用
//	    log.Println("锁已被占用")
//	}
//
// 配置驱动类型：
//
//	// 通过 Build 函数创建
//	mgr, err := lockmgr.Build("memory", map[string]any{
//	    "max_backups": 1000,
//	})
//
//	// 通过配置提供者创建
//	mgr, err := lockmgr.BuildWithConfigProvider(configProvider)
//
// 使用场景：
//   - 分布式环境下的资源互斥访问
//   - 并发控制，防止重复操作
//   - 任务队列的任务消费
//   - 限流控制
package lockmgr
