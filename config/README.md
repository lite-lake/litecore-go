# Config - 配置管理

统一的配置管理模块，支持 JSON 和 YAML 格式的配置文件读取与查询。

## 特性

- **多格式支持** - JSON、YAML 配置文件解析
- **类型安全** - 泛型 API 提供编译时类型检查
- **灵活查询** - 点分隔路径和数组索引语法
- **智能转换** - 自动处理数字类型转换（float64 ↔ int）
- **线程安全** - 配置数据不可变，支持并发访问
- **默认值** - 内置默认值支持，简化错误处理

## 快速开始

```go
package main

import (
    "log"

    "github.com/lite-lake/litecore-go/config"
)

func main() {
    // 创建配置提供者
    provider, err := config.NewConfigProvider("yaml", "./config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 获取配置值（类型安全）
    host, err := config.Get[string](provider, "database.host")
    if err != nil {
        log.Fatal(err)
    }

    port, err := config.Get[int](provider, "database.port")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("连接到 %s:%d", host, port)

    // 使用默认值
    timeout := config.GetWithDefault(provider, "server.timeout", 30)
    log.Printf("超时设置: %d 秒", timeout)

    // 检查键是否存在
    if provider.Has("feature.enabled") {
        log.Println("功能已启用")
    }
}
```

**配置文件示例（config.yaml）：**
```yaml
app:
  name: "我的应用"
  version: "1.0.0"

database:
  host: "localhost"
  port: 3306
  credentials:
    username: "admin"
    password: "secret"

server:
  host: "0.0.0.0"
  port: 8080
  timeout: 30

servers:
  - host: "server1.example.com"
    port: 8001
    ssl: true
  - host: "server2.example.com"
    port: 8002
    ssl: true

features:
  caching: true
  logging: false
```

## 路径语法

### 对象嵌套访问

使用点号（`.`）分隔嵌套对象的层级：

```go
// 访问 database.host
host, _ := config.Get[string](provider, "database.host")
// 返回: "localhost"

// 访问 database.credentials.username
username, _ := config.Get[string](provider, "database.credentials.username")
// 返回: "admin"

// 访问 server.timeout
timeout, _ := config.Get[int](provider, "server.timeout")
// 返回: 30
```

### 数组索引访问

使用方括号（`[index]`）访问数组元素：

```go
// 访问 servers 数组第一个元素的 host
host, _ := config.Get[string](provider, "servers[0].host")
// 返回: "server1.example.com"

// 访问 servers 数组第二个元素的 port
port, _ := config.Get[int](provider, "servers[1].port")
// 返回: 8002

// 访问整个数组元素
server, _ := provider.Get("servers[0]")
// 返回: map[string]any{"host": "server1.example.com", "port": 8001, "ssl": true}
```

### 混合访问

结合对象嵌套和数组索引：

```go
// 深层嵌套访问
ssl, _ := config.Get[bool](provider, "servers[0].ssl")
// 返回: true

// 获取所有配置数据
all, _ := provider.Get("")
// 返回: 完整的配置 map[string]any
```

## API 参考

### 工厂函数

#### `NewConfigProvider`

```go
func NewConfigProvider(driver string, filePath string) (common.BaseConfigProvider, error)
```

创建配置提供者实例。

**参数：**
- `driver`: 配置文件类型，支持 `"json"` 或 `"yaml"`
- `filePath`: 配置文件路径

**返回：**
- `common.BaseConfigProvider`: 配置提供者接口
- `error`: 错误信息（文件不存在、解析失败等）

**示例：**
```go
// YAML 配置
provider, err := config.NewConfigProvider("yaml", "./config.yaml")

// JSON 配置
provider, err := config.NewConfigProvider("json", "./config.json")
```

### 泛型获取函数

#### `Get`

```go
func Get[T any](provider common.BaseConfigProvider, key string) (T, error)
```

获取指定类型的配置值。

**类型参数：**
- `T`: 目标类型（支持 `string`, `int`, `int32`, `int64`, `float64`, `bool`）

**参数：**
- `provider`: 配置提供者
- `key`: 配置键路径

**返回：**
- `T`: 配置值
- `error`: 错误信息（键不存在、类型不匹配等）

**示例：**
```go
// 获取字符串
name, err := config.Get[string](provider, "app.name")

// 获取整数
port, err := config.Get[int](provider, "server.port")

// 获取浮点数
rate, err := config.Get[float64](provider, "tax.rate")

// 获取布尔值
enabled, err := config.Get[bool](provider, "features.caching")
```

#### `GetWithDefault`

```go
func GetWithDefault[T any](provider common.BaseConfigProvider, key string, defaultValue T) T
```

获取配置值，如果不存在或出错则返回默认值。

**参数：**
- `provider`: 配置提供者
- `key`: 配置键路径
- `defaultValue`: 默认值

**返回：**
- `T`: 配置值或默认值

**示例：**
```go
// 键不存在时返回默认值
timeout := config.GetWithDefault(provider, "server.timeout", 30)
maxRetries := config.GetWithDefault(provider, "server.maxRetries", 3)
```

### 工具函数

#### `IsConfigKeyNotFound`

```go
func IsConfigKeyNotFound(err error) bool
```

判断错误是否为"键不存在"错误。

**参数：**
- `err`: 错误信息

**返回：**
- `bool`: 是否为键不存在错误

**示例：**
```go
val, err := config.Get[string](provider, "some.key")
if config.IsConfigKeyNotFound(err) {
    // 处理键不存在的情况
    log.Println("配置项不存在，使用默认值")
} else if err != nil {
    // 处理其他错误（类型不匹配等）
    log.Printf("获取配置失败: %v", err)
}
```

### ConfigProvider 接口

```go
type BaseConfigProvider interface {
    Get(key string) (any, error)
    Has(key string) bool
    ConfigProviderName() string
}
```

**方法：**
- `Get(key string) (any, error)`: 获取配置值（返回 any 类型）
- `Has(key string) bool`: 检查键是否存在
- `ConfigProviderName() string`: 返回提供者名称

## 错误处理

### 错误类型

包定义了三种错误类型：

```go
var (
    ErrKeyNotFound  = errors.New("config key not found")
    ErrTypeMismatch = errors.New("type mismatch")
    ErrInvalidValue = errors.New("invalid value")
)
```

### 错误判断

使用 `IsConfigKeyNotFound` 函数判断错误类型：

```go
import "github.com/lite-lake/litecore-go/config"

val, err := config.Get[string](provider, "database.host")
if err != nil {
    if config.IsConfigKeyNotFound(err) {
        // 键不存在
        log.Println("配置项 database.host 不存在")
    } else {
        // 其他错误（类型不匹配、无效值等）
        log.Printf("获取配置失败: %v", err)
    }
}
```

### 类型转换

包会自动尝试类型转换：

```go
// JSON 中的整数是 float64
// 可以自动转换为 int
port, err := config.Get[int](provider, "server.port")
// JSON: {"port": 8080} → port = 8080 (int)

// 字符串转整数
count, err := config.Get[int](provider, "item.count")
// JSON: {"count": "42"} → count = 42 (int)

// 数字转字符串
version, err := config.Get[string](provider, "api.version")
// JSON: {"version": 1.0} → version = "1" (string)
```

### 使用默认值简化错误处理

```go
// 不需要错误检查，键不存在时自动使用默认值
timeout := config.GetWithDefault(provider, "server.timeout", 30)
maxConn := config.GetWithDefault(provider, "database.maxConnections", 10)
```

## 线程安全

配置提供者在创建后不可变，可以安全地在多个 goroutine 之间共享使用：

```go
// 全局共享配置提供者
var globalConfig common.BaseConfigProvider

func init() {
    var err error
    globalConfig, err = config.NewConfigProvider("yaml", "./config.yaml")
    if err != nil {
        log.Fatal(err)
    }
}

// 多个 goroutine 可以安全地并发访问
func worker() {
    timeout := config.GetWithDefault(globalConfig, "worker.timeout", 30)
    // 使用 timeout...
}

func main() {
    for i := 0; i < 10; i++ {
        go worker()
    }
}
```

## 最佳实践

### 1. 配置文件组织

```go
// 推荐：使用环境前缀区分不同环境
configFile := fmt.Sprintf("./config.%s.yaml", env)
provider, err := config.NewConfigProvider("yaml", configFile)

// 示例文件：
// - config.dev.yaml   (开发环境)
// - config.test.yaml  (测试环境)
// - config.prod.yaml  (生产环境)
```

### 2. 配置结构设计

```yaml
# 推荐：扁平化配置，避免过深嵌套
app:
  name: "myapp"
  version: "1.0.0"

database:
  host: "localhost"
  port: 3306

# 避免超过 3 层嵌套
```

### 3. 使用默认值

```go
// 推荐：为可选配置提供默认值
timeout := config.GetWithDefault(provider, "server.timeout", 30)
debug := config.GetWithDefault(provider, "app.debug", false)

// 而不是：
timeout, err := config.Get[int](provider, "server.timeout")
if err != nil {
    timeout = 30  // 手动设置默认值
}
```

### 4. 配置验证

```go
// 推荐：创建配置验证函数
func validateConfig(provider common.BaseConfigProvider) error {
    requiredKeys := []string{
        "app.name",
        "database.host",
        "database.port",
    }

    for _, key := range requiredKeys {
        if !provider.Has(key) {
            return fmt.Errorf("缺少必需的配置: %s", key)
        }
    }

    return nil
}
```

### 5. 类型断言

```go
// 推荐：使用泛型 Get 函数，自动类型转换
port, err := config.Get[int](provider, "server.port")

// 而不是：
val, _ := provider.Get("server.port")
port := val.(int)  // 手动类型断言，容易 panic
```

## 常见问题

**Q: 为什么 JSON 中的整数被解析为 float64？**

A: 这是 JSON 标准库的默认行为。config 包会自动将整数类型的 float64 转换为 int，无需手动处理。

**Q: 如何处理环境变量覆盖配置？**

A: 可以在配置文件中设置默认值，然后在代码中检查环境变量：

```go
// 配置文件设置默认值
debug := config.GetWithDefault(provider, "app.debug", false)

// 环境变量覆盖
if envDebug := os.Getenv("APP_DEBUG"); envDebug != "" {
    debug = envDebug == "true"
}
```

**Q: 支持配置热更新吗？**

A: 当前版本不支持。配置在创建后不可变，如需更新需要重新创建 provider。

**Q: 如何处理敏感信息（如密码）？**

A: 建议使用环境变量或专门的密钥管理系统，不要将敏感信息硬编码在配置文件中。
