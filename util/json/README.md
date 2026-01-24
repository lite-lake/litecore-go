# util/json

JSON 工具包，提供 JSON 验证、格式化、数据转换和路径查询等功能。

## 特性

- **数据验证与格式化** - 快速验证 JSON 有效性，支持美化和压缩输出
- **灵活的数据转换** - 支持 JSON 字符串、Map 和结构体之间的双向转换
- **便捷的路径操作** - 使用点号语法访问嵌套字段，支持对象和数组索引
- **高级操作** - 支持深度合并和差异比较
- **丰富的工具函数** - 类型检测、键值查询、大小获取等辅助功能

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/json"
)

func main() {
    j := json.JSON

    // JSON 验证
    if j.IsValid(`{"name":"Alice"}`) {
        fmt.Println("JSON 有效")
    }

    // JSON 转 Map
    data, _ := j.ToMap(`{"name":"Alice","age":30}`)
    fmt.Printf("Name: %v\n", data["name"])

    // JSON 转 Struct
    type User struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }
    var user User
    j.ToStruct(`{"name":"Bob","age":25}`, &user)
    fmt.Printf("User: %+v\n", user)

    // 路径操作
    value, _ := j.GetValue(`{"user":{"name":"Alice"}}`, "user.name")
    fmt.Printf("Value: %v\n", value)
}
```

## 数据转换

### JSON 转 Map

```go
j := json.JSON

// 普通转换
data, err := j.ToMap(`{"name":"Alice","age":30}`)
if err != nil {
    // 处理错误
}

// 严格模式（必须是对象类型）
data, err := j.ToMapStrict(`{"name":"Alice"}`)
if err != nil {
    // 处理错误
}
```

### JSON 转 Struct

```go
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

var user User
err := j.ToStruct(`{"name":"Alice","age":30}`, &user)
if err != nil {
    // 处理错误
}
```

### Go 类型转 JSON

```go
// Map 转 JSON
m := map[string]interface{}{"name": "Alice", "age": 30}
jsonStr, _ := j.FromMap(m)

// Struct 转 JSON（压缩）
type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}
cfg := Config{Host: "localhost", Port: 8080}
jsonStr, _ := j.FromStruct(cfg)
// 输出：{"host":"localhost","port":8080}

// Struct 转 JSON（格式化）
jsonStr, _ := j.FromStructWithIndent(cfg, "  ")
```

## 路径操作

使用点号语法访问嵌套字段：

- **对象字段**：`user.name`
- **嵌套字段**：`user.profile.age`
- **数组元素**：`items.0`（索引从 0 开始）
- **数组中的对象**：`users.0.name`

```go
j := json.JSON

jsonStr := `{
    "user": {
        "name": "Alice",
        "profile": {"age": 30, "city": "Beijing"}
    },
    "items": [1, 2, 3],
    "users": [{"name": "Alice"}, {"name": "Bob"}]
}`

// 获取任意类型值
value, _ := j.GetValue(jsonStr, "user.name")

// 获取字符串
name, _ := j.GetString(jsonStr, "user.name")

// 获取数字
age, _ := j.GetFloat64(jsonStr, "user.profile.age")

// 获取布尔值
active, _ := j.GetBool(jsonStr, "user.active")

// 访问数组元素
firstItem, _ := j.GetValue(jsonStr, "items.0")
secondUserName, _ := j.GetString(jsonStr, "users.1.name")
```

## 高级操作

### JSON 合并

```go
j := json.JSON

// 简单合并
merged, _ := j.Merge(
    `{"name":"Alice","age":25}`,
    `{"age":30,"city":"Beijing"}`,
)
// 输出：{"name":"Alice","age":30,"city":"Beijing"}

// 深度合并（嵌套对象）
defaultConfig := `{
    "database": {"host": "localhost", "port": 3306, "ssl": false},
    "logging": {"level": "info"}
}`

userConfig := `{
    "database": {"host": "prod.example.com", "password": "secret"},
    "logging": {"level": "debug"}
}`

merged, _ = j.Merge(defaultConfig, userConfig)
// database.host 被覆盖，database.port 和 database.ssl 保留，database.password 新增
// logging.level 被覆盖
```

### JSON 比较

```go
j := json.JSON

// 比较两个 JSON 是否不同
hasDiff, _ := j.Diff(
    `{"name":"Alice","age":30}`,
    `{"name":"Bob","age":30}`,
)
if hasDiff {
    fmt.Println("JSON 存在差异")
}

// 验证转换一致性
original := `{"name":"Test","value":123}`
m, _ := j.ToMap(original)
converted, _ := j.FromMap(m)
hasDiff, _ = j.Diff(original, converted)
fmt.Printf("转换后一致: %v\n", !hasDiff) // true
```

## 格式化与转义

```go
j := json.JSON

// 格式化 JSON（美化）
formatted, _ := j.PrettyPrint(`{"name":"Alice"}`, "  ")
// 输出：
// {
//   "name": "Alice"
// }

// 默认缩进（2 空格）
formatted, _ := j.PrettyPrintWithIndent(`{"name":"Alice"}`)

// 压缩 JSON
compacted, _ := j.Compact(`{ "name" : "Alice" }`)
// 输出：{"name":"Alice"}

// 转义特殊字符
escaped, _ := j.Escape(`Hello\nWorld`)
// 输出：Hello\nWorld

// 反转义
unescaped, _ := j.Unescape(`Hello\\nWorld`)
// 输出：Hello\n（实际换行符）
```

## 工具函数

```go
j := json.JSON

// 类型检查
j.IsObject(`{"name":"Alice"}`) // true
j.IsArray(`[1,2,3]`)           // true

// 获取类型
typeName, _ := j.GetType(jsonStr, "user.profile")
// "object" | "array" | "string" | "number" | "boolean" | "null"

// 获取对象的所有键
keys, _ := j.GetKeys(jsonStr, "user")
// ["name", "age", "email"]

// 获取数组或对象的大小
size, _ := j.GetSize(jsonStr, "items")    // 数组长度
size, _ = j.GetSize(jsonStr, "user")      // 对象键数量

// 检查是否包含指定键
exists, _ := j.Contains(jsonStr, "user", "email") // true
```

## API 参考

### 验证与格式化

| 函数 | 说明 |
|------|------|
| `IsValid(jsonStr string) bool` | 验证 JSON 是否有效 |
| `PrettyPrint(jsonStr, indent string) (string, error)` | 使用指定缩进格式化 |
| `PrettyPrintWithIndent(jsonStr string) (string, error)` | 使用默认缩进（2 空格）格式化 |
| `Compact(jsonStr string) (string, error)` | 压缩 JSON |
| `Escape(str string) (string, error)` | 转义特殊字符 |
| `Unescape(str string) (string, error)` | 反转义 |

### 数据转换

| 函数 | 说明 |
|------|------|
| `ToMap(jsonStr string) (map[string]any, error)` | JSON 转 Map |
| `ToMapStrict(jsonStr string) (map[string]any, error)` | 严格模式（必须是对象） |
| `ToStruct(jsonStr string, target any) error` | JSON 转 Struct |
| `FromMap(data map[string]any) (string, error)` | Map 转 JSON |
| `FromStruct(data any) (string, error)` | Struct 转 JSON |
| `FromStructWithIndent(data any, indent string) (string, error)` | Struct 转 JSON（带缩进） |

### 路径操作

| 函数 | 说明 |
|------|------|
| `GetValue(jsonStr, path string) (any, error)` | 获取任意类型值 |
| `GetString(jsonStr, path string) (string, error)` | 获取字符串值 |
| `GetFloat64(jsonStr, path string) (float64, error)` | 获取数字值 |
| `GetBool(jsonStr, path string) (bool, error)` | 获取布尔值 |

### 高级操作

| 函数 | 说明 |
|------|------|
| `Merge(jsonStr1, jsonStr2 string) (string, error)` | 合并两个 JSON 对象 |
| `Diff(jsonStr1, jsonStr2 string) (bool, error)` | 比较差异 |

### 工具函数

| 函数 | 说明 |
|------|------|
| `IsObject(jsonStr string) bool` | 检查是否为对象 |
| `IsArray(jsonStr string) bool` | 检查是否为数组 |
| `GetType(jsonStr, path string) (string, error)` | 获取值类型 |
| `GetKeys(jsonStr, path string) ([]string, error)` | 获取对象的所有键 |
| `GetSize(jsonStr, path string) (int, error)` | 获取数组或对象大小 |
| `Contains(jsonStr, path, key string) (bool, error)` | 检查对象是否包含指定键 |

## 使用场景

### 配置文件管理

```go
// 合并默认配置和用户配置
defaultConfig := `{
    "app": {"name": "MyApp", "version": "1.0.0"},
    "server": {"port": 8080, "timeout": 30}
}`

userConfig := `{
    "app": {"debug": true},
    "server": {"port": 9000}
}`

finalConfig, _ := j.Merge(defaultConfig, userConfig)
```

### API 响应处理

```go
apiResponse := `{
    "status": "success",
    "data": {"user": {"id": 123, "name": "Alice"}}
}`

// 提取嵌套数据
userName, _ := j.GetString(apiResponse, "data.user.name")
```

### 数据验证

```go
// 验证配置文件格式
configData := readFile("config.json")
if !j.IsValid(configData) {
    log.Fatal("无效的配置文件")
}

// 检查必需字段
if j.Contains(configData, "", "database") {
    dbType, _ := j.GetType(configData, "database")
    if dbType != "object" {
        log.Fatal("数据库配置无效")
    }
}
```

## 注意事项

1. **路径语法**：使用点号 `.` 分隔符，不支持通配符或复杂表达式
2. **数组索引**：使用非负整数，从 0 开始
3. **Merge 操作**：仅支持对象类型，数组合并会直接覆盖
4. **类型转换**：GetString/GetFloat64/GetBool 会尝试自动类型转换
5. **浮点数精度**：JSON 数字会被解析为 `float64`，大整数可能丢失精度
6. **键顺序**：Go 的 map 不保证键顺序，多次序列化结果可能不同

## 相关模块

- `util/map` - Map 操作工具
- `util/string` - 字符串处理工具
- `util/convert` - 类型转换工具
