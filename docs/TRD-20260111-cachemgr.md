# CacheManager 开发计划

## 1. 概述

CacheManager 是 LiteCore 框架的缓存管理组件，负责缓存的创建、管理和生命周期控制。支持 Redis、Memory 和 None 三种缓存驱动，提供统一的接口和配置方式。

## 2. 架构设计

### 2.1 整体结构

```
manager/cachemgr/
├── doc.go                    # 包文档
├── interface.go              # 缓存管理器接口定义
├── factory.go                # 工厂函数，用于创建管理器实例
├── cache_adapter.go          # 缓存适配器
├── README.md                 # 使用文档
└── internal/
    ├── config/               # 配置解析和验证
    │   ├── config.go         # 配置结构体和解析逻辑
    │   └── config_test.go    # 配置测试
    └── drivers/              # 驱动实现
        ├── base_manager.go   # 基础管理器（实现 common.Manager）
        ├── base_manager_test.go
        ├── none_manager.go   # 空实现（降级方案）
        ├── none_manager_test.go
        ├── redis_driver.go   # Redis 驱动实现
        ├── redis_driver_test.go
        └── memory_driver.go  # Memory 驱动实现
        └── memory_driver_test.go
```

### 2.2 接口设计

#### CacheManager 接口

```go
type CacheManager interface {
    // 继承 common.Manager 接口
    ManagerName() string
    Health() error
    OnStart() error
    OnStop() error

    // Get 获取缓存值
    Get(ctx context.Context, key string, dest any) error

    // Set 设置缓存值
    Set(ctx context.Context, key string, value any, expiration time.Duration) error

    // SetNX 仅当键不存在时才设置值（Set if Not eXists）
    // 返回值表示是否设置成功：true 表示设置成功，false 表示键已存在
    // 常用于分布式锁、幂等性控制等场景
    SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)

    // Delete 删除缓存值
    Delete(ctx context.Context, key string) error

    // Exists 检查键是否存在
    Exists(ctx context.Context, key string) (bool, error)

    // Expire 设置过期时间
    Expire(ctx context.Context, key string, expiration time.Duration) error

    // TTL 获取剩余过期时间
    TTL(ctx context.Context, key string) (time.Duration, error)

    // Clear 清空所有缓存（慎用）
    Clear(ctx context.Context) error

    // GetMultiple 批量获取
    GetMultiple(ctx context.Context, keys []string) (map[string]any, error)

    // SetMultiple 批量设置
    SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error

    // DeleteMultiple 批量删除
    DeleteMultiple(ctx context.Context, keys []string) error

    // Increment 自增
    Increment(ctx context.Context, key string, value int64) (int64, error)

    // Decrement 自减
    Decrement(ctx context.Context, key string, value int64) (int64, error)

    // Close 关闭缓存连接
    Close() error
}
```

### 2.3 配置设计

#### CacheConfig 结构

```go
type CacheConfig struct {
    Driver      string       `yaml:"driver"`      // 驱动类型: redis, memory, none
    RedisConfig *RedisConfig `yaml:"redis_config"` // Redis 配置
    MemoryConfig *MemoryConfig `yaml:"memory_config"` // Memory 配置
}

// RedisConfig Redis 缓存配置
type RedisConfig struct {
    Host            string `yaml:"host"`              // Redis 主机地址
    Port            int    `yaml:"port"`              // Redis 端口
    Password        string `yaml:"password"`          // Redis 密码
    DB              int    `yaml:"db"`                // Redis 数据库编号
    MaxIdleConns    int    `yaml:"max_idle_conns"`    // 最大空闲连接数
    MaxOpenConns    int    `yaml:"max_open_conns"`    // 最大打开连接数
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"` // 连接最大存活时间
}

// MemoryConfig 内存缓存配置
type MemoryConfig struct {
    MaxSize    int    `yaml:"max_size"`    // 最大缓存大小（MB）
    MaxAge     time.Duration `yaml:"max_age"`     // 最大缓存时间
    MaxBackups int    `yaml:"max_backups"` // 最大备份项数（清理策略相关）
    Compress   bool   `yaml:"compress"`    // 是否压缩
}
```

## 3. 依赖库选择

### 3.1 Redis 客户端库

**选择**: `github.com/redis/go-redis/v9`

**理由**:
- 官方推荐的 Redis Go 客户端
- 功能完整，支持 Redis 6+ 特性
- 支持 Pub/Sub、Streams、事务等
- 性能优秀，支持连接池
- 社区活跃，文档完善

### 3.2 内存缓存库

**选择**: `github.com/patrickmn/go-cache`

**理由**:
- 简单易用的内存缓存库
- 支持过期时间
- 线程安全
- 无需额外依赖
- 2025 年仍在广泛使用和维护

### 3.3 依赖声明

```go
import (
    "context"
    "time"

    // Redis 客户端
    "github.com/redis/go-redis/v9"

    // 内存缓存
    "github.com/patrickmn/go-cache"

    "com.litelake.litecore/common"
)
```

## 4. 开发任务

### 4.1 第一阶段：基础框架

#### 任务 1.1：创建目录结构
- [ ] 创建 `manager/cachemgr/` 目录
- [ ] 创建 `internal/config/` 目录
- [ ] 创建 `internal/drivers/` 目录

#### 任务 1.2：实现接口定义
- [ ] 创建 `interface.go`，定义 CacheManager 接口
- [ ] 创建 `doc.go`，编写包文档

#### 任务 1.3：实现配置解析
- [ ] 创建 `internal/config/config.go`
  - 定义 CacheConfig 结构体
  - 定义 RedisConfig、MemoryConfig 结构体
  - 实现 ParseCacheConfigFromMap() 函数
  - 实现 Validate() 方法
  - 设置默认值
- [ ] 创建 `internal/config/config_test.go`
  - 测试配置解析
  - 测试配置验证
  - 测试默认值

#### 任务 1.4：实现基础管理器
- [ ] 创建 `internal/drivers/base_manager.go`
  - 实现 common.Manager 接口
  - 提供公共方法
- [ ] 创建 `internal/drivers/base_manager_test.go`

### 4.2 第二阶段：驱动实现

#### 任务 2.1：实现 None 驱动
- [ ] 创建 `internal/drivers/none_manager.go`
  - 空实现，用于降级场景
  - 所有操作返回错误或空值（但不 panic）
- [ ] 创建 `internal/drivers/none_manager_test.go`

#### 任务 2.2：实现 Memory 驱动
- [ ] 创建 `internal/drivers/memory_driver.go`
  - 使用 go-cache 库
  - 实现基本 CRUD 操作
  - 实现批量操作
  - 实现计数器操作（Increment/Decrement）
  - 实现 health check
- [ ] 创建 `internal/drivers/memory_driver_test.go`
  - 测试基本操作
  - 测试过期时间
  - 测试批量操作
  - 测试并发安全

#### 任务 2.3：实现 Redis 驱动
- [ ] 创建 `internal/drivers/redis_driver.go`
  - 使用 go-redis v9 库
  - 实现连接池配置
  - 实现基本 CRUD 操作
  - 实现批量操作
  - 实现计数器操作
  - 实现 health check
  - 实现序列化/反序列化（使用 gob 或 json）
- [ ] 创建 `internal/drivers/redis_driver_test.go`
  - 测试基本操作（需要 Redis 实例或 miniredis）
  - 测试过期时间
  - 测试批量操作
  - 测试连接池

### 4.3 第三阶段：工厂和适配器

#### 任务 3.1：实现工厂函数
- [ ] 创建 `factory.go`
  - 实现 NewFactory() 构造函数
  - 实现 Build(driver string, cfg map[string]any) 方法
  - 实现 BuildWithConfig(config *CacheConfig) 方法
  - 实现降级逻辑（失败时返回 none 驱动）

#### 任务 3.2：实现适配器
- [ ] 创建 `cache_adapter.go`
  - 将内部驱动适配到 CacheManager 接口
  - 实现 common.Manager 接口方法
  - 实现 CacheManager 特有方法

### 4.4 第四阶段：文档和测试

#### 任务 4.1：编写使用文档
- [ ] 创建 `README.md`
  - 快速开始
  - 配置说明
  - 使用示例
  - 最佳实践

#### 任务 4.2：集成测试
- [ ] 创建 `integration_test.go`
  - 测试完整的初始化流程
  - 测试配置加载
  - 测试多驱动切换
  - 测试降级场景

#### 任务 4.3：性能测试（可选）
- [ ] 创建 `benchmark_test.go`
  - 测试读写性能
  - 测试批量操作性能
  - 测试并发性能
  - 对比 Redis vs Memory 性能

## 5. 技术细节

### 5.1 数据序列化

由于 Redis 只支持存储字符串和字节，需要进行序列化处理：

**方案**: 使用 gob 编码
- 优点：原生支持，性能较好
- 缺点：需要类型注册

**备选方案**: 使用 JSON
- 优点：跨语言，易调试
- 缺点：性能较差，不支持所有类型

```go
// 序列化
func serialize(value any) ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(value)
    return buf.Bytes(), err
}

// 反序列化
func deserialize(data []byte, dest any) error {
    buf := bytes.NewReader(data)
    dec := gob.NewDecoder(buf)
    return dec.Decode(dest)
}
```

### 5.2 配置默认值

#### Redis 默认值
```go
RedisConfig: &RedisConfig{
    Host:            "localhost",
    Port:            6379,
    Password:        "",
    DB:              0,
    MaxIdleConns:    10,
    MaxOpenConns:    100,
    ConnMaxLifetime: 30 * time.Second,
}
```

#### Memory 默认值
```go
MemoryConfig: &MemoryConfig{
    MaxSize:    100,  // MB
    MaxAge:     30 * 24 * time.Hour,  // 30 天
    MaxBackups: 1000,
    Compress:   false,
}
```

### 5.3 错误处理策略

1. **配置解析失败**：返回 none 驱动，记录错误日志
2. **连接创建失败**：返回 none 驱动，记录错误日志
3. **缓存操作失败**：返回原始错误，由调用方处理
4. **Health check 失败**：返回错误，但不影响管理器运行
5. **序列化失败**：返回序列化错误

### 5.4 降级策略

当以下情况发生时，自动降级到 none 驱动：
- 配置解析失败
- 驱动初始化失败
- Redis 连接创建失败

none 驱动行为：
- 所有 Get 操作返回 "cache not available" 错误
- 所有 Set/Delete 操作为空操作（不返回错误，但也不执行）
- Health() 返回错误
- OnStart() 和 OnStop() 无操作

### 5.5 并发安全

- **Memory 驱动**: go-cache 本身是并发安全的
- **Redis 驱动**: go-redis 客户端是并发安全的
- **None 驱动**: 无状态，天然并发安全

### 5.6 健康检查实现

#### Redis 健康检查
```go
func (m *RedisManager) Health() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return m.client.Ping(ctx).Err()
}
```

#### Memory 健康检查
```go
func (m *MemoryManager) Health() error {
    // 内存缓存一般不会失败，简单返回 nil
    return nil
}
```

## 6. 测试策略

### 6.1 单元测试

每个驱动都需要独立的单元测试：
- 配置解析测试
- 基本操作测试（Get/Set/Delete）
- 批量操作测试
- 计数器操作测试
- 过期时间测试
- 错误处理测试

### 6.2 集成测试

- 测试完整初始化流程
- 测试多驱动场景
- 测试降级机制
- 测试生命周期管理

### 6.3 Redis 测试策略

使用 `github.com/alicebob/miniredis` 进行 Redis 模拟测试：
- 轻量级，无需启动真实 Redis
- 支持大部分 Redis 命令
- 适合单元测试

可选：使用 testcontainers 运行真实 Redis 进行集成测试

## 7. 使用示例

### 7.1 配置示例

#### Redis 配置示例

```yaml
cache:
  driver: redis
  redis_config:
    host: localhost
    port: 6379
    password: password
    db: 0
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 30s
```

#### Memory 配置示例

```yaml
cache:
  driver: memory
  memory_config:
    max_size: 100MB
    max_age: 30d
    max_backups: 1000
    compress: false
```

#### None 配置示例

```yaml
cache:
  driver: none
```

### 7.2 代码示例

```go
// 从配置创建缓存管理器
cfg := loadConfig() // 从 YAML 加载配置
factory := cachemgr.NewFactory()
cacheMgr := factory.Build("redis", cfg["cache"].(map[string]any))

// 基本操作
ctx := context.Background()

// 设置缓存
err := cacheMgr.Set(ctx, "user:123", user, 10*time.Minute)

// 获取缓存
var user User
err = cacheMgr.Get(ctx, "user:123", &user)

// 删除缓存
err = cacheMgr.Delete(ctx, "user:123")

// 检查键是否存在
exists, err := cacheMgr.Exists(ctx, "user:123")

// 获取剩余过期时间
ttl, err := cacheMgr.TTL(ctx, "user:123")

// 批量操作
items := map[string]any{
    "user:123": user1,
    "user:456": user2,
}
err = cacheMgr.SetMultiple(ctx, items, 10*time.Minute)

values, err := cacheMgr.GetMultiple(ctx, []string{"user:123", "user:456"})

// 计数器
counter, err := cacheMgr.Increment(ctx, "views:page:1", 1)
counter, err := cacheMgr.Decrement(ctx, "views:page:1", 1)
```

## 8. 开发里程碑

| 阶段 | 任务 | 预计时间 |
|------|------|----------|
| 第一阶段 | 基础框架（接口、配置、基础管理器） | 1 天 |
| 第二阶段 | 驱动实现（None、Memory、Redis） | 2-3 天 |
| 第三阶段 | 工厂和适配器 | 0.5-1 天 |
| 第四阶段 | 文档和测试 | 1-2 天 |

**总计**：4.5-7 天

## 9. 注意事项

1. **序列化**:
   - 使用 gob 需要注册类型
   - 注意指针类型和值类型的区别
   - 复杂类型（如 channel、func）无法序列化

2. **安全性**:
   - 不要在日志中打印 Redis 密码
   - 使用环境变量管理敏感信息
   - 考虑使用 Redis ACL（Redis 6+）

3. **性能**:
   - Redis 使用连接池
   - 批量操作优先于单个操作
   - 合理设置过期时间避免内存泄漏

4. **兼容性**:
   - Memory 驱动仅限单机
   - Redis 支持 6.0+ 版本
   - 不同驱动的 TTL 行为可能有差异

5. **可观测性**:
   - 记录缓存命中率（可选）
   - 记录慢操作
   - 集成 telemetryMgr 进行追踪

## 10. 后续扩展

- [ ] 支持缓存穿透保护（空值缓存）
- [ ] 支持缓存击穿保护（互斥锁）
- [ ] 支持缓存雪崩保护（随机过期时间）
- [ ] 支持本地缓存 + Redis 的二级缓存
- [ ] 支持缓存预热
- [ ] 支持缓存统计指标（命中率、读写次数）
- [ ] 支持更灵活的序列化方式（JSON、MsgPack）
- [ ] 支持 Redis Cluster 模式
- [ ] 支持 Redis Sentinel 模式
- [ ] 支持缓存订阅/发布通知

## 11. 参考资料

### 第三方库

- **go-redis**: [https://github.com/redis/go-redis](https://github.com/redis/go-redis) - 官方 Redis Go 客户端
- **go-cache**: [https://github.com/patrickmn/go-cache](https://github.com/patrickmn/go-cache) - 内存缓存库
- **miniredis**: [https://github.com/alicebob/miniredis](https://github.com/alicebob/miniredis) - Redis 模拟器（测试用）

### 相关文档

- [Redis Go 客户端指南](https://redis.io/docs/clients/golang/)
- [go-redis v9 文档](https://redis.uptrace.dev/)
- [go-cache 使用指南](https://github.com/patrickmn/go-cache)
