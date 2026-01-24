# ConfigManager

配置管理器，支持 JSON 和 YAML 格式配置文件的加载与查询。

## 特性

- **多格式支持** - 支持 JSON 和 YAML 两种常见配置格式
- **路径查询** - 支持点分隔路径和数组索引语法
- **类型安全** - 泛型 API 支持自动类型转换
- **线程安全** - 配置数据不可变，可安全并发访问
- **依赖注入** - 由 Engine 自动初始化并注入到各层组件

## 快速开始

### 通过 Engine 初始化（推荐）

在 5 层依赖注入架构中，ConfigManager 由 Engine 自动初始化：

```go
import (
    "github.com/lite-lake/litecore-go/server"
)

// 在创建 Engine 时指定配置文件
engine := server.NewEngine(
    &server.BuiltinConfig{
        Driver:   "yaml",  // 或 "json"
        FilePath: "configs/config.yaml",
    },
    entityContainer,
    repositoryContainer,
    serviceContainer,
    controllerContainer,
    middlewareContainer,
)

// Engine 会自动创建并注入 ConfigManager
```

### 在 Service/Repository 中使用

通过依赖注入获取 ConfigManager：

```go
type MyService struct {
    Config configmgr.IConfigManager `inject:""`
}

func (s *MyService) OnStart() error {
    // 获取字符串
    name, err := configmgr.Get[string](s.Config, "app.name")
    if err != nil {
        return err
    }

    // 获取整数
    port, err := configmgr.Get[int](s.Config, "server.port")
    if err != nil {
        return err
    }

    // 带默认值获取
    timeout := configmgr.GetWithDefault(s.Config, "server.timeout", 30)

    return nil
}
```

### 直接创建实例

```go
import "github.com/lite-lake/litecore-go/manager/configmgr"

// 创建配置管理器
mgr, err := configmgr.NewConfigManager("yaml", "config.yaml")
if err != nil {
    log.Fatal(err)
}

// 获取配置值
name := configmgr.Get[string](mgr, "app.name")
port := configmgr.Get[int](mgr, "server.port")

// 带默认值的获取
timeout := configmgr.GetWithDefault(mgr, "server.timeout", 30)
```

## 路径语法

配置路径使用点（.）分隔各层键名，支持数组索引语法：

```yaml
app:
  name: "myapp"
server:
  host: "localhost"
  port: 8080
servers:
  - host: "s1.example.com"
    port: 8001
  - host: "s2.example.com"
    port: 8002
```

```go
// 简单键
configmgr.Get[int](mgr, "port")                    // 返回 8080

// 嵌套路径
configmgr.Get[string](mgr, "server.host")         // 返回 "localhost"

// 数组元素
configmgr.Get[int](mgr, "servers[0].port")         // 返回 8001
configmgr.Get[string](mgr, "servers[1].host")      // 返回 "s2.example.com"
```

## 类型安全获取

### 基本类型

```go
// 获取字符串
name, err := configmgr.Get[string](mgr, "app.name")

// 获取整数
port, err := configmgr.Get[int](mgr, "server.port")

// 获取布尔值
enabled, err := configmgr.Get[bool](mgr, "debug")

// 获取浮点数
timeout, err := configmgr.Get[float64](mgr, "timeout")
```

### 默认值

```go
// 配置不存在时使用默认值
port := configmgr.GetWithDefault(mgr, "server.port", 8080)

// 获取失败时返回默认值
unknown := configmgr.GetWithDefault(mgr, "unknown.key", "default")
```

### 类型转换

`Get` 函数支持智能类型转换：

```go
// JSON 中的 float64 自动转换为 int
// JSON: {"count": 42.0}
count, err := configmgr.Get[int](mgr, "count")  // 42

// 字符串转 bool
// YAML: enabled: "true"
enabled, err := configmgr.Get[bool](mgr, "enabled")  // true

// 字符串转 int
port, err := configmgr.Get[int](mgr, "port")  // "8080" -> 8080
```

## 工厂函数

### NewConfigManager

```go
// 从 YAML 文件创建
yamlMgr, err := configmgr.NewConfigManager("yaml", "config.yaml")

// 从 JSON 文件创建
jsonMgr, err := configmgr.NewConfigManager("json", "config.json")
```

### Build

```go
// 通过 Build 函数创建
mgr, err := configmgr.Build("yaml", "config.yaml")
if err != nil {
    log.Fatal(err)
}
```

支持的驱动类型：
- `"yaml"` - YAML 格式配置文件
- `"json"` - JSON 格式配置文件

## 接口方法

### IConfigManager

```go
type IConfigManager interface {
    common.IBaseManager
    Get(key string) (any, error)
    Has(key string) bool
}
```

#### Get(key string) (any, error)

获取配置项，支持路径语法：

```go
value, err := mgr.Get("server.host")
```

#### Has(key string) bool

检查配置项是否存在：

```go
if mgr.Has("server.port") {
    // 配置存在
}
```

## 错误处理

```go
import "errors"

val, err := configmgr.Get[string](mgr, "key")
if err != nil {
    if configmgr.IsConfigKeyNotFound(err) {
        // 键不存在的错误
    }
    if errors.Is(err, configmgr.ErrTypeMismatch) {
        // 类型不匹配的错误
    }
}
```

## API

### 工厂函数

| 函数 | 说明 |
|------|------|
| `NewConfigManager(driver, filePath)` | 创建配置管理器实例 |
| `Build(driver, filePath)` | 根据驱动类型创建配置管理器 |

### 工具函数

| 函数 | 说明 |
|------|------|
| `Get[T](mgr, key)` | 类型安全获取配置值 |
| `GetWithDefault[T](mgr, key, defaultValue)` | 带默认值获取配置 |
| `IsConfigKeyNotFound(err)` | 判断是否为键不存在错误 |

### 加载函数

| 函数 | 说明 |
|------|------|
| `LoadJSON(filePath)` | 加载 JSON 配置文件 |
| `LoadYAML(filePath)` | 加载 YAML 配置文件 |

## 错误类型

| 变量 | 说明 |
|------|------|
| `ErrKeyNotFound` | 配置键不存在 |
| `ErrTypeMismatch` | 类型不匹配 |

## 性能特性

- 路径解析使用预编译正则表达式，提升性能
- 配置数据在加载后不可变，无需锁保护
- 支持高并发读取场景

## 使用示例

### 在 Service 中读取配置

```go
type AuthService struct {
    Config configmgr.IConfigManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (s *AuthService) VerifyPassword(password string) bool {
    storedPassword, err := configmgr.Get[string](s.Config, "app.admin.password")
    if err != nil {
        s.LoggerMgr.Ins().Error("获取管理员密码失败", "error", err)
        return false
    }
    return hash.BcryptVerify(password, storedPassword)
}
```

### 在 Repository 中初始化配置

```go
type MessageRepository struct {
    Config configmgr.IConfigManager `inject:""`
    pageSize int
}

func (r *MessageRepository) OnStart() error {
    // 从配置读取分页大小，使用默认值 20
    r.pageSize = configmgr.GetWithDefault(r.Config, "app.page_size", 20)
    return nil
}
```
