# ConfigManager

配置管理器，支持 JSON 和 YAML 格式配置文件的加载与查询。

## 特性

- **多格式支持** - 支持 JSON 和 YAML 两种常见配置格式
- **路径查询** - 支持点分隔路径和数组索引语法
- **类型安全** - 泛型 API 支持自动类型转换
- **线程安全** - 配置数据不可变，可安全并发访问

## 快速开始

```go
import "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"

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
mgr.Get("port")                           // 返回 8080

// 嵌套路径
mgr.Get("server.host")                    // 返回 "localhost"

// 数组元素
mgr.Get("servers[0].port")                // 返回 8001
mgr.Get("servers[1].host")                // 返回 "s2.example.com"
```

## 类型安全获取

使用泛型 `Get` 函数进行类型安全的配置获取：

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

使用 `GetWithDefault` 在配置不存在时返回默认值：

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

创建配置管理器实例：

```go
// 从 YAML 文件创建
yamlMgr, err := configmgr.NewConfigManager("yaml", "config.yaml")

// 从 JSON 文件创建
jsonMgr, err := configmgr.NewConfigManager("json", "config.json")
```

支持的驱动类型：
- `"yaml"` - YAML 格式配置文件
- `"json"` - JSON 格式配置文件

## 接口方法

### IConfigManager

配置管理器基础接口，继承自 `common.IBaseManager`。

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
