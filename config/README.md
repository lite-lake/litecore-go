# Config

配置管理模块，支持 JSON 和 YAML 格式的配置文件读取与查询。

## 特性

- **多格式支持** - JSON、YAML 配置文件
- **路径查询** - 点分隔路径和数组索引语法
- **类型安全** - 泛型 API 提供编译时类型检查
- **智能转换** - 自动处理数字类型转换
- **线程安全** - 配置数据不可变

## 快速开始

```go
// 创建配置提供者
provider, err := config.NewConfigProvider("yaml", "./config.yaml")
if err != nil {
    log.Fatal(err)
}

// 获取配置值
host, err := config.Get[string](provider, "database.host")
port, err := config.Get[int](provider, "database.port")

// 使用默认值
timeout := config.GetWithDefault(provider, "server.timeout", 30)

// 检查键是否存在
if provider.Has("feature.enabled") {
    // ...
}
```

## 路径语法

### 对象嵌套

```go
// 配置: { "database": { "host": "localhost", "port": 3306 } }
provider.Get("database.host")  // "localhost"
provider.Get("database.port")  // 3306
```

### 数组索引

```go
// 配置: { "servers": [{ "host": "s1", "port": 8080 }, { "host": "s2", "port": 8081 }] }
provider.Get("servers[0].host")  // "s1"
provider.Get("servers[1].port")  // 8081
provider.Get("servers[2]")       // 错误：索引越界
```

## API

### 工厂函数

```go
func NewConfigProvider(driver string, filePath string) (ConfigProvider, error)
```

支持的 driver: `"json"`, `"yaml"`

### 泛型获取函数

```go
func Get[T any](provider ConfigProvider, key string) (T, error)
func GetWithDefault[T any](provider ConfigProvider, key string, defaultValue T) T
```

### ConfigProvider 接口

```go
type ConfigProvider interface {
    Get(key string) (any, error)
    Has(key string) bool
}
```

## 错误处理

```go
import "com.litelake.litecore/config"

val, err := config.Get[string](provider, "some.key")
if config.IsConfigKeyNotFound(err) {
    // 键不存在
} else if err != nil {
    // 其他错误（类型不匹配等）
}
```

## 线程安全

配置提供者在创建后不可变，可以安全地在多个 goroutine 之间共享使用，无需额外的同步机制。
