# util/json

全面的 JSON 操作工具集，提供 JSON 验证、格式化、数据转换、路径查询等功能。

## 特性

- **数据验证与格式化** - 快速验证 JSON 有效性，支持美化和压缩输出
- **灵活的数据转换** - 支持 JSON 字符串、Map 和结构体之间的双向转换
- **便捷的路径操作** - 使用点号语法访问嵌套字段，支持对象和数组索引
- **高级操作** - 支持深度合并和差异比较
- **转义处理** - 提供字符串的转义和反转义功能
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

## JSON 序列化

### 结构体转 JSON

```go
j := json.JSON

type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

cfg := Config{Host: "localhost", Port: 8080}

// 压缩格式
jsonStr, _ := j.FromStruct(cfg)
// 输出：{"host":"localhost","port":8080}

// 格式化输出（2 空格缩进）
formatted, _ := j.FromStructWithIndent(cfg, "  ")
// 输出：
// {
//   "host": "localhost",
//   "port": 8080
// }

// 使用默认缩进（2 空格）
formatted, _ := j.PrettyPrintWithIndent(jsonStr)
```

### Map 转 JSON

```go
j := json.JSON

data := map[string]interface{}{
    "name":  "Alice",
    "age":   30,
    "email": "alice@example.com",
}

jsonStr, _ := j.FromMap(data)
// 输出：{"name":"Alice","age":30,"email":"alice@example.com"}
```

## JSON 反序列化

### JSON 转 Map

```go
j := json.JSON

// 普通转换
data, err := j.ToMap(`{"name":"Alice","age":30}`)
if err != nil {
    // 处理错误
}
fmt.Printf("Name: %v\n", data["name"])

// 严格模式（必须是对象类型）
data, err := j.ToMapStrict(`{"name":"Alice"}`)
if err != nil {
    // 处理错误
}
```

### JSON 转 Struct

```go
j := json.JSON

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
fmt.Printf("User: %+v\n", user)
```

## JSON 格式化

### 美化 JSON

```go
j := json.JSON

// 使用自定义缩进
formatted, err := j.PrettyPrint(`{"name":"Alice","age":30}`, "  ")
if err != nil {
    // 处理错误
}
// 输出：
// {
//   "name": "Alice",
//   "age": 30
// }

// 使用默认缩进（2 空格）
formatted, err := j.PrettyPrintWithIndent(`{"name":"Alice"}`)
```

### 压缩 JSON

```go
j := json.JSON

compacted, err := j.Compact(`{ "name" : "Alice" , "age" : 30 }`)
if err != nil {
    // 处理错误
}
// 输出：{"name":"Alice","age":30}
```

## JSON 验证

```go
j := json.JSON

// 验证 JSON 有效性
isValid := j.IsValid(`{"name":"Alice"}`)
// true

isValid = j.IsValid(`{invalid}`)
// false
```

## 字符串转义

### 转义特殊字符

```go
j := json.JSON

// 转义字符串
escaped, err := j.Escape(`Hello\nWorld`)
if err != nil {
    // 处理错误
}
// 输出：Hello\nWorld

// 包含引号
escaped, _ = j.Escape(`say "hello"`)
// 输出：say \"hello\"

// 包含反斜杠
escaped, _ = j.Escape(`path\to\file`)
// 输出：path\\to\\file
```

### 反转义

```go
j := json.JSON

// 反转义
unescaped, err := j.Unescape(`Hello\\nWorld`)
if err != nil {
    // 处理错误
}
// 输出：Hello\n（实际换行符）

// Unicode 转义
unescaped, _ = j.Unescape(`\u4e2d\u6587`)
// 输出：中文
```

## 路径操作

使用点号语法访问嵌套字段：

- **对象字段**：`user.name`
- **嵌套字段**：`user.profile.age`
- **数组元素**：`items.0`（索引从 0 开始）
- **数组中的对象**：`users.0.name`

### 获取任意类型值

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
fmt.Println(value) // Alice

// 访问数组元素
firstItem, _ := j.GetValue(jsonStr, "items.0")
fmt.Println(firstItem) // 1

// 访问数组中的对象
secondUserName, _ := j.GetValue(jsonStr, "users.1.name")
fmt.Println(secondUserName) // Bob

// 获取根对象
root, _ := j.GetValue(jsonStr, "")
```

### 获取特定类型值

```go
j := json.JSON

jsonStr := `{
    "name": "Alice",
    "age": 30,
    "active": true,
    "price": 99.99
}`

// 获取字符串
name, _ := j.GetString(jsonStr, "name")
fmt.Println(name) // Alice

// 获取数字
age, _ := j.GetFloat64(jsonStr, "age")
fmt.Println(age) // 30

// 获取布尔值
active, _ := j.GetBool(jsonStr, "active")
fmt.Println(active) // true

// 类型自动转换
ageStr, _ := j.GetString(jsonStr, "age")
fmt.Println(ageStr) // "30"
```

## JSON 合并

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
// database.host 被覆盖为 "prod.example.com"
// database.port 保留默认值 3306
// database.ssl 保留默认值 false
// database.password 新增 "secret"
// logging.level 被覆盖为 "debug"
```

## JSON 比较

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

## 工具函数

### 类型检查

```go
j := json.JSON

// 检查 JSON 类型
j.IsObject(`{"name":"Alice"}`) // true
j.IsArray(`[1,2,3]`)           // true

// 获取值类型
typeName, _ := j.GetType(`{"name":"Alice"}`, "name")
// "string"

typeName, _ = j.GetType(`{"age":30}`, "age")
// "number"

typeName, _ = j.GetType(`{"active":true}`, "active")
// "boolean"

typeName, _ = j.GetType(`{"data":null}`, "data")
// "null"
```

### 键值操作

```go
j := json.JSON

jsonStr := `{
    "user": {
        "name": "Alice",
        "age": 30,
        "email": "alice@example.com"
    }
}`

// 获取对象的所有键
keys, _ := j.GetKeys(jsonStr, "user")
// ["name", "age", "email"]

// 检查是否包含指定键
exists, _ := j.Contains(jsonStr, "user", "email") // true
exists, _ = j.Contains(jsonStr, "user", "phone") // false

// 获取数组或对象的大小
size, _ := j.GetSize(jsonStr, "user")      // 3 (对象键数量)

jsonStr2 := `{"items": [1, 2, 3, 4, 5]}`
size, _ = j.GetSize(jsonStr2, "items")    // 5 (数组长度)
```

## API 说明

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

### 日志数据格式化

```go
// 美化日志中的 JSON 数据
logData := map[string]interface{}{
    "timestamp": time.Now().Unix(),
    "level":     "info",
    "message":   "User logged in",
    "user": map[string]interface{}{
        "id":    123,
        "name":  "Alice",
        "email": "alice@example.com",
    },
}

jsonStr, _ := j.FromMap(logData)
formatted, _ := j.PrettyPrint(jsonStr, "  ")
fmt.Println(formatted)
```

## 注意事项

1. **路径语法**：使用点号 `.` 分隔符，不支持通配符或复杂表达式
2. **数组索引**：使用非负整数，从 0 开始
3. **Merge 操作**：仅支持对象类型，数组合并会直接覆盖
4. **类型转换**：GetString/GetFloat64/GetBool 会尝试自动类型转换
5. **浮点数精度**：JSON 数字会被解析为 `float64`，大整数可能丢失精度
6. **键顺序**：Go 的 map 不保证键顺序，多次序列化结果可能不同
7. **转义处理**：Escape 和 Unescape 遵循 JSON 字符串转义规则
8. **空值处理**：null 值在路径操作中可以被获取，但在类型转换时可能返回错误

## 相关模块

- `util/map` - Map 操作工具
- `util/string` - 字符串处理工具
- `util/convert` - 类型转换工具
