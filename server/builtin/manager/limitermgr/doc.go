// Package limitermgr 提供统一的限流管理功能，支持 Redis、内存和无限流三种驱动。
//
// 核心特性：
//   - 多驱动支持：支持 Redis（分布式限流）、Memory（本地内存限流）、None（无限流）三种限流驱动
//   - 统一接口：提供统一的 ILimiterManager 接口，便于切换限流实现
//   - 可观测性：内置日志、指标和链路追踪支持
//   - 滑动窗口：支持时间窗口内的请求数限制
//   - 剩余查询：支持查询剩余可访问次数
//
// 基本用法：
//
//	// 使用内存限流
//	mgr := limitermgr.NewLimiterManagerMemoryImpl(&limitermgr.MemoryLimiterConfig{})
//	defer mgr.Close()
//
//	ctx := context.Background()
//
//	// 检查是否允许通过（100次/分钟）
//	allowed, err := mgr.Allow(ctx, "user:123", 100, time.Minute)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if !allowed {
//	    log.Println("请求被限流")
//	}
//
//	// 获取剩余次数
//	remaining, err := mgr.GetRemaining(ctx, "user:123", 100, time.Minute)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	log.Printf("剩余可访问次数: %d", remaining)
//
// 使用 Redis 限流：
//
//	cfg := &limitermgr.RedisLimiterConfig{
//	    Host:     "localhost",
//	    Port:     6379,
//	    Password: "",
//	    DB:       0,
//	}
//	mgr := limitermgr.NewLimiterManagerRedisImpl(cfg)
//	defer mgr.Close()
//
// 配置驱动类型：
//
//	// 通过 Build 函数创建
//	mgr, err := limitermgr.Build("memory", map[string]any{
//	    "max_backups": 1000,
//	})
//
//	// 通过配置提供者创建
//	mgr, err := limitermgr.BuildWithConfigProvider(configProvider)
//
// 使用场景：
//   - API 接口请求频率限制
//   - 用户行为频率控制（如点赞、评论）
//   - 防止恶意请求和爬虫
//   - 资源使用配额管理
package limitermgr
