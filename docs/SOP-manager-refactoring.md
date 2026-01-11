# Manager 包重构 SOP（标准操作流程）

本文档基于 `manager/cachemgr` 的简洁结构，定义了 manager 包重构的标准操作流程。

## 一、目录结构规范

### 1.1 扁平化结构（推荐）

```
manager/{name}/
├── doc.go              # 包文档
├── interface.go        # 核心接口定义
├── config.go           # 配置结构和解析函数
├── impl_base.go        # 基础实现（可观测性、工具函数）
├── {driver}_impl.go    # 各驱动实现
├── factory.go          # 工厂函数（依赖注入友好）
├── *_test.go           # 测试文件
```

示例（cachemgr）：
```
manager/cachemgr/
├── doc.go
├── interface.go
├── config.go
├── impl_base.go
├── redis_impl.go
├── memory_impl.go
├── none_impl.go
├── factory.go
└── *_test.go
```

### 1.2 避免的结构（不推荐）

```
manager/{name}/
├── internal/
│   ├── config/        # ❌ 避免嵌套过深
│   │   └── config.go
│   └── drivers/       # ❌ 避免过度封装
│       ├── base.go
│       ├── driver1.go
│       └── driver2.go
```

## 二、文件职责说明

### 2.1 `interface.go` - 接口定义

**职责**：定义管理器的核心接口

**内容**：
- 继承 `common.BaseManager` 接口（生命周期管理）
- 定义业务特定方法

```go
// 示例
type CacheManager interface {
    // 生命周期方法（来自 BaseManager）
    ManagerName() string
    Health() error
    OnStart() error
    OnStop() error

    // 业务方法
    Get(ctx, key, dest) error
    Set(ctx, key, value, exp) error
    // ...
}
```

### 2.2 `config.go` - 配置管理

**职责**：定义配置结构和提供解析函数

**内容**：
1. 配置结构体（使用 yaml 标签）
2. DefaultConfig() 函数
3. Validate() 方法
4. ParseXxxConfigFromMap() 函数（支持 ConfigProvider）

```go
// 示例
type CacheConfig struct {
    Driver       string       `yaml:"driver"`
    RedisConfig  *RedisConfig `yaml:"redis_config"`
    MemoryConfig *MemoryConfig `yaml:"memory_config"`
}

func DefaultConfig() *CacheConfig { ... }
func (c *CacheConfig) Validate() error { ... }
func ParseCacheConfigFromMap(cfg map[string]any) (*CacheConfig, error) { ... }
```

### 2.3 `impl_base.go` - 基础实现

**职责**：提供可观测性支持和公共工具函数

**内容**：
1. 基础结构体（包含依赖注入字段）
2. 初始化可观测性组件的方法
3. 操作记录函数（链路追踪、指标、日志）
4. 公共工具函数（验证、脱敏等）

```go
// 示例
type cacheManagerBaseImpl struct {
    loggerMgr    loggermgr.LoggerManager      `inject:""`
    telemetryMgr telemetrymgr.TelemetryManager `inject:""`
    logger       loggermgr.Logger
    tracer       trace.Tracer
    meter        metric.Meter
}

func (b *cacheManagerBaseImpl) initObservability() { ... }
func (b *cacheManagerBaseImpl) recordOperation(...) error { ... }
func ValidateContext(ctx) error { ... }
func ValidateKey(key) error { ... }
```

**设计原则**：
- 使用依赖注入标签 `inject:""`
- 提供统一的可观测性支持
- 封装公共逻辑，减少重复代码

### 2.4 `{driver}_impl.go` - 驱动实现

**职责**：实现特定驱动的业务逻辑

**内容**：
1. 驱动特定的结构体（嵌入 baseImpl）
2. 构造函数 NewXxxManagerImpl(cfg)
3. 实现所有接口方法

```go
// 示例
type cacheManagerRedisImpl struct {
    *cacheManagerBaseImpl
    client *redis.Client
    name   string
}

func NewCacheManagerRedisImpl(cfg *RedisConfig) (CacheManager, error) {
    impl := &cacheManagerRedisImpl{
        cacheManagerBaseImpl: newCacheManagerBaseImpl(),
        client:               client,
        name:                 "cacheManagerRedisImpl",
    }
    impl.initObservability()
    return impl, nil
}

func (r *cacheManagerRedisImpl) Get(...) error {
    return r.recordOperation(ctx, r.name, "get", key, func() error {
        // 实际业务逻辑
    })
}
```

**设计原则**：
- 通过嵌入 `*cacheManagerBaseImpl` 获得可观测性能力
- 使用 `recordOperation` 包装所有操作以自动记录
- 保持业务逻辑简洁

### 2.5 `factory.go` - 工厂函数

**职责**：提供统一的创建接口，支持依赖注入

**内容**：
1. `Build(driver, config) - 基础创建函数
2. `BuildWithConfigProvider(provider) - 从配置提供者创建

```go
// 示例
func Build(driverType string, driverConfig map[string]any) (CacheManager, error) {
    switch driverType {
    case "redis":
        cfg, _ := parseRedisConfig(driverConfig)
        return NewCacheManagerRedisImpl(cfg)
    case "memory":
        cfg, _ := parseMemoryConfig(driverConfig)
        return NewCacheManagerMemoryImpl(cfg)
    case "none":
        return NewCacheManagerNoneImpl(), nil
    default:
        return nil, fmt.Errorf("unsupported driver: %s", driverType)
    }
}

func BuildWithConfigProvider(configProvider common.BaseConfigProvider) (CacheManager, error) {
    if configProvider == nil {
        return nil, fmt.Errorf("configProvider cannot be nil")
    }

    driverType, _ := configProvider.Get("cache.driver")
    driverConfig, _ := configProvider.Get("cache.redis_config")
    return Build(driverType, driverConfig)
}
```

**配置路径规范**：
- 驱动类型：`{manager}.driver`（如 `cache.driver`、`telemetry.driver`）
- 驱动配置：`{manager}.{driver}_config`（如 `cache.redis_config`、`telemetry.otel_config`）
- 不使用实例名称，保持全局唯一性

**设计原则**：
- 简单直接，避免过度抽象
- 支持配置提供者（依赖注入）
- 返回接口类型，隐藏实现细节

## 三、命名规范

### 3.1 文件命名

- `interface.go` - 接口定义
- `config.go` - 配置结构
- `impl_base.go` - 基础实现
- `{driver}_impl.go` - 驱动实现（小写下划线）
- `factory.go` - 工厂函数
- `doc.go` - 包文档

### 3.2 类型命名

- 接口：`XxxManager`（如 `CacheManager`）
- 实现类：`xxxManager{Driver}Impl`（如 `cacheManagerRedisImpl`）
- 配置：`XxxConfig`（如 `RedisConfig`）

### 3.3 函数命名

- 构造函数：`New{Type}Impl(cfg)`（如 `NewCacheManagerRedisImpl`）
- 工厂函数：`Build(...)`、`BuildWithConfigProvider(...)`
- 工具函数：驼峰命名（如 `ValidateContext`、`sanitizeKey`）

## 四、可观测性集成规范

### 4.1 依赖注入模式

在 `impl_base.go` 中定义可观测性依赖：

```go
type xxxManagerBaseImpl struct {
    loggerMgr    loggermgr.LoggerManager      `inject:""`
    telemetryMgr telemetrymgr.TelemetryManager `inject:""`
    logger       loggermgr.Logger
    tracer       trace.Tracer
    meter        metric.Meter
}
```

### 4.2 初始化可观测性

```go
func (b *xxxManagerBaseImpl) initObservability() {
    if b.loggerMgr != nil {
        b.logger = b.loggerMgr.Logger("xxxmgr")
    }
    if b.telemetryMgr != nil {
        b.tracer = b.telemetryMgr.Tracer("xxxmgr")
        b.meter = b.telemetryMgr.Meter("xxxmgr")
    }
}
```

### 4.3 操作记录模式

```go
func (x *xxxImpl) Operation(...) error {
    return x.recordOperation(ctx, x.name, "operation", key, func() error {
        // 实际业务逻辑
    })
}
```

## 五、重构步骤

### 步骤 1：分析现有结构
- 列出当前文件和目录
- 识别需要移除的内部包

### 步骤 2：创建新文件
- 在根目录创建 `config.go`（合并 internal/config）
- 在根目录创建 `impl_base.go`（合并 internal/drivers/base）
- 创建 `{driver}_impl.go`（替换 internal/drivers/*.go）

### 步骤 3：更新工厂函数
- 修改 `factory.go` 以使用新的结构
- 添加 `BuildWithConfigProvider` 支持

### 步骤 4：更新导入路径
- 修复所有 import 语句
- 移除 `internal/` 前缀

### 步骤 5：运行测试
- 运行 `go test ./...`
- 确保所有测试通过

### 步骤 6：清理
- 删除旧的 `internal/` 目录
- 更新文档

## 六、检查清单

重构完成后，确保：

- [ ] 目录结构扁平化（无 internal 子目录）
- [ ] 所有配置集中在 `config.go`
- [ ] 所有实现在根目录的 `*_impl.go` 文件中
- [ ] `impl_base.go` 提供可观测性支持
- [ ] `factory.go` 支持依赖注入
- [ ] 所有测试通过
- [ ] 文档已更新

## 七、示例对比

### 重构前（telemetrymgr - 当前）

```
manager/telemetrymgr/
├── internal/
│   ├── config/
│   │   └── config.go
│   └── drivers/
│       ├── base_manager.go
│       ├── otel_manager.go
│       └── none_manager.go
├── interface.go
├── factory.go (已废弃)
└── manager.go
```

### 重构后（telemetrymgr - 目标）

```
manager/telemetrymgr/
├── doc.go
├── interface.go
├── config.go
├── impl_base.go
├── otel_impl.go
├── none_impl.go
├── factory.go
├── *_test.go
```

## 八、配置路径规范

### 8.1 标准配置路径格式

所有 manager 包的配置路径必须遵循统一的格式：

```
{manager}.driver           # 驱动类型
{manager}.{driver}_config  # 驱动配置
```

**示例**：

| Manager | 驱动类型配置 | 驱动配置 |
|---------|-------------|----------|
| cachemgr | `cache.driver` | `cache.redis_config`<br>`cache.memory_config` |
| telemetrymgr | `telemetry.driver` | `telemetry.otel_config` |
| databasemgr | `database.driver` | `database.mysql_config`<br>`database.postgres_config` |

### 8.2 配置示例

**cachemgr 配置示例**：
```yaml
cache:
  driver: redis                    # 驱动类型
  redis_config:                    # Redis 配置
    host: localhost
    port: 6379
    db: 0
  memory_config:                   # Memory 配置（备用）
    max_size: 100
    max_age: 24h
```

**telemetrymgr 配置示例**：
```yaml
telemetry:
  driver: otel                     # 驱动类型
  otel_config:                     # OTEL 配置
    endpoint: localhost:4317
    insecure: true
    traces:
      enabled: true
    metrics:
      enabled: false
```

### 8.3 错误的配置路径（避免）

❌ **错误**：使用实例名称前缀
```yaml
# 错误：使用了实例名称
telemetry.default.driver: otel
telemetry.default.otel_config: {...}
```

✅ **正确**：不使用实例名称
```yaml
# 正确：直接使用管理器名称
telemetry.driver: otel
telemetry.otel_config: {...}
```

### 8.4 BuildWithConfigProvider 实现规范

```go
func BuildWithConfigProvider(configProvider common.BaseConfigProvider) (XxxManager, error) {
    // 1. 检查 provider 是否为 nil
    if configProvider == nil {
        return nil, fmt.Errorf("configProvider cannot be nil")
    }

    // 2. 读取驱动类型 {manager}.driver
    driverType, err := configProvider.Get("{manager}.driver")
    if err != nil {
        return nil, fmt.Errorf("failed to get {manager}.driver: %w", err)
    }

    // 3. 根据驱动类型读取对应配置 {manager}.{driver}_config
    var driverConfig map[string]any
    switch driverType {
    case "driver1":
        driverConfig, _ = configProvider.Get("{manager}.driver1_config")
    case "driver2":
        driverConfig, _ = configProvider.Get("{manager}.driver2_config")
    }

    // 4. 调用 Build 函数
    return Build(driverType, driverConfig)
}
```

## 九、最佳实践

1. **保持简单**：避免过度抽象和过度封装
2. **依赖注入**：使用 `inject:""` 标签支持容器注入
3. **可观测性优先**：所有操作都应记录日志和指标
4. **接口优先**：工厂函数返回接口类型
5. **配置灵活**：支持多种配置方式（ConfigProvider、直接传参）
6. **配置统一**：遵循统一的配置路径规范，不使用实例名称前缀
