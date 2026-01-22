# util/json

JSON 操作工具库，提供便捷的 JSON 数据验证、格式化、转换和路径操作功能。

## 特性

- **完整的数据验证** - 支持 JSON 格式验证、类型检查和数据结构判断
- **灵活的格式化** - 提供美化打印、压缩输出和自定义缩进功能
- **强大的数据转换** - 支持与 Map、Struct 之间的双向转换
- **便捷的路径操作** - 使用简单的点号语法访问嵌套数据
- **智能的合并与比较** - 支持深度合并 JSON 对象和差异检测
- **丰富的工具函数** - 包含转义、类型检测、键值操作等实用功能

## 快速开始

### 基础使用

```go
package main

import (
    "fmt"
    "litecore-go/util/json"
)

func main() {
    // 获取 JSON 工具实例
    j := json.New()

    // 验证 JSON
    jsonStr := `{"name":"Alice","age":30,"city":"Beijing"}`
    if j.IsValid(jsonStr) {
        fmt.Println("有效的 JSON")
    }

    // 格式化输出
    formatted, _ := j.PrettyPrintWithIndent(jsonStr)
    fmt.Println(formatted)

    // 路径访问
    name, _ := j.GetString(jsonStr, "name")
    fmt.Printf("Name: %s\n", name)
}
```

### 数据转换示例

```go
// JSON 转 Map
jsonStr := `{"name":"Bob","age":25,"tags":["go","java"]}`
dataMap, err := j.ToMap(jsonStr)
if err == nil {
    fmt.Printf("Name: %v\n", dataMap["name"])
}

// JSON 转 Struct
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
var p Person
j.ToStruct(jsonStr, &p)
fmt.Printf("Person: %+v\n", p)

// Map 转 JSON
m := map[string]interface{}{
    "status": "success",
    "data": map[string]string{
        "message": "Hello",
    },
}
jsonResult, _ := j.FromMap(m)
fmt.Println(jsonResult)
```

## 功能详解

### 1. JSON 验证

使用 `IsValid` 方法验证 JSON 字符串的有效性。

```go
j := json.New()

// 有效的 JSON
validJSON := `{"name":"test","value":123}`
fmt.Println(j.IsValid(validJSON)) // true

// 无效的 JSON
invalidJSON := `{name:"test"}`
fmt.Println(j.IsValid(invalidJSON)) // false

// 检查 JSON 类型
objJSON := `{"data":"value"}`
arrJSON := `[1,2,3]`
fmt.Println(j.IsObject(objJSON))  // true
fmt.Println(j.IsArray(arrJSON))   // true
```

### 2. 格式化输出

提供多种 JSON 格式化选项，包括美化打印和压缩输出。

```go
j := json.New()

jsonStr := `{"name":"Alice","age":30,"address":{"city":"Beijing"}}`

// 自定义缩进（2 个空格）
formatted, _ := j.PrettyPrint(jsonStr, "  ")
fmt.Println(formatted)
// 输出：
// {
//   "name": "Alice",
//   "age": 30,
//   "address": {
//     "city": "Beijing"
//   }
// }

// 使用默认缩进（2 个空格）
formatted, _ = j.PrettyPrintWithIndent(jsonStr)

// 压缩 JSON（移除所有空白字符）
compacted, _ := j.Compact(jsonStr)
fmt.Println(compacted)
// 输出：{"name":"Alice","age":30,"address":{"city":"Beijing"}}
```

### 3. 数据转换

支持 JSON 字符串与 Go 数据结构之间的双向转换。

#### JSON 转 Map

```go
j := json.New()

jsonStr := `{
    "user": {
        "name": "Alice",
        "age": 30,
        "tags": ["developer", "golang"]
    }
}`

// 转换为 map[string]interface{}
dataMap, err := j.ToMap(jsonStr)
if err == nil {
    user := dataMap["user"].(map[string]interface{})
    fmt.Printf("Name: %v\n", user["name"])
    fmt.Printf("Tags: %v\n", user["tags"])
}

// 严格模式（必须是对象类型）
dataMap, err := j.ToMapStrict(jsonStr)
if err != nil {
    fmt.Println("转换失败:", err)
}
```

#### JSON 转 Struct

```go
j := json.New()

type User struct {
    Name    string   `json:"name"`
    Age     int      `json:"age"`
    Email   string   `json:"email"`
    Tags    []string `json:"tags"`
    Address struct {
        City    string `json:"city"`
        Country string `json:"country"`
    } `json:"address"`
}

jsonStr := `{
    "name": "Bob",
    "age": 25,
    "email": "bob@example.com",
    "tags": ["engineer", "python"],
    "address": {
        "city": "Shanghai",
        "country": "China"
    }
}`

var user User
err := j.ToStruct(jsonStr, &user)
if err == nil {
    fmt.Printf("User: %+v\n", user)
    fmt.Printf("Name: %s, City: %s\n", user.Name, user.Address.City)
}
```

#### Go 数据类型转 JSON

```go
j := json.New()

// Map 转 JSON
m := map[string]interface{}{
    "status": "success",
    "code":   200,
    "data": map[string]string{
        "message": "Operation completed",
    },
}
jsonStr, _ := j.FromMap(m)
fmt.Println(jsonStr)

// Struct 转 JSON（压缩格式）
type Config struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    SSL      bool   `json:"ssl"`
}
config := Config{Host: "localhost", Port: 8080, SSL: true}
jsonStr, _ = j.FromStruct(config)
fmt.Println(jsonStr)
// 输出：{"host":"localhost","port":8080,"ssl":true}

// Struct 转 JSON（格式化）
jsonStr, _ = j.FromStructWithIndent(config, "  ")
fmt.Println(jsonStr)
```

### 4. 路径操作

使用简单的点号语法访问 JSON 数据中的嵌套字段。

#### 路径语法说明

- **根对象**：使用空字符串 `""` 或点号 `"."`
- **对象字段**：使用 `fieldName` 访问
- **嵌套字段**：使用 `parent.child.grandchild` 访问
- **数组元素**：使用 `arrayName.index` 访问（索引从 0 开始）
- **数组中的对象**：使用 `users.0.name` 访问

#### 基本路径访问

```go
j := json.New()

jsonStr := `{
    "status": "success",
    "data": {
        "user": {
            "id": 123,
            "name": "Alice",
            "profile": {
                "age": 30,
                "city": "Beijing"
            }
        },
        "items": [1, 2, 3, 4, 5],
        "users": [
            {"name": "Alice", "age": 30},
            {"name": "Bob", "age": 25}
        ]
    }
}`

// 获取任意类型值
value, _ := j.GetValue(jsonStr, "data.user.name")
fmt.Printf("Name: %v\n", value)

// 获取字符串
name, _ := j.GetString(jsonStr, "data.user.name")
fmt.Printf("Name: %s\n", name)

// 获取数字
age, _ := j.GetFloat64(jsonStr, "data.user.profile.age")
fmt.Printf("Age: %.0f\n", age)

// 获取布尔值
// active, _ := j.GetBool(jsonStr, "data.user.active")

// 访问数组元素
firstItem, _ := j.GetValue(jsonStr, "data.items.0")
fmt.Printf("First item: %v\n", firstItem)

// 访问数组中的对象
secondUserName, _ := j.GetString(jsonStr, "data.users.1.name")
fmt.Printf("Second user: %s\n", secondUserName)

// 获取整个数组
items, _ := j.GetValue(jsonStr, "data.items")
fmt.Printf("Items: %v\n", items)
```

#### 高级路径操作

```go
j := json.New()

jsonStr := `{
    "configmgr": {
        "database": {
            "host": "localhost",
            "port": 3306,
            "credentials": {
                "username": "admin",
                "password": "secret"
            }
        },
        "logging": {
            "level": "info",
            "format": "json"
        }
    }
}`

// 获取类型信息
typeName, _ := j.GetType(jsonStr, "configmgr.database")
fmt.Printf("Type: %s\n", typeName) // "object"

// 获取对象的所有键
keys, _ := j.GetKeys(jsonStr, "configmgr.database")
fmt.Printf("Keys: %v\n", keys) // ["host", "port", "credentials"]

// 获取对象或数组的大小
size, _ := j.GetSize(jsonStr, "configmgr.database")
fmt.Printf("Size: %d\n", size) // 3

// 检查是否包含某个键
hasHost, _ := j.Contains(jsonStr, "configmgr.database", "host")
fmt.Printf("Has host: %v\n", hasHost) // true

hasTimeout, _ := j.Contains(jsonStr, "configmgr.database", "timeout")
fmt.Printf("Has timeout: %v\n", hasTimeout) // false
```

### 5. JSON 合并

支持深度合并两个 JSON 对象，后者的值会覆盖前者。

```go
j := json.New()

// 简单合并
json1 := `{"name":"Alice","age":25}`
json2 := `{"age":30,"city":"Beijing"}`

merged, _ := j.Merge(json1, json2)
fmt.Println(merged)
// 输出：{"name":"Alice","age":30,"city":"Beijing"}

// 嵌套对象合并（深度合并）
defaultConfig := `{
    "database": {
        "host": "localhost",
        "port": 3306,
        "ssl": false,
        "timeout": 30
    },
    "logging": {
        "level": "info",
        "format": "text"
    }
}`

userConfig := `{
    "database": {
        "host": "production.example.com",
        "password": "secret"
    },
    "logging": {
        "level": "debug"
    }
}`

merged, _ = j.Merge(defaultConfig, userConfig)
fmt.Println(merged)
// 输出：
// {
//     "database": {
//         "host": "production.example.com",  // 被覆盖
//         "port": 3306,                      // 保留
//         "ssl": false,                      // 保留
//         "timeout": 30,                     // 保留
//         "password": "secret"               // 新增
//     },
//     "logging": {
//         "level": "debug",                  // 被覆盖
//         "format": "text"                   // 保留
//     }
// }

// 多层嵌套合并
config1 := `{"a":{"b":{"c":1,"d":2}}}`
config2 := `{"a":{"b":{"d":3,"e":4}}}`

merged, _ = j.Merge(config1, config2)
fmt.Println(merged)
// 输出：{"a":{"b":{"c":1,"d":3,"e":4}}}
```

### 6. JSON 比较

比较两个 JSON 字符串是否有差异。

```go
j := json.New()

// 相同的对象
json1 := `{"name":"Alice","age":30}`
json2 := `{"name":"Alice","age":30}`
hasDiff, _ := j.Diff(json1, json2)
fmt.Printf("有差异: %v\n", hasDiff) // false

// 不同的值
json3 := `{"name":"Alice","age":30}`
json4 := `{"name":"Bob","age":30}`
hasDiff, _ = j.Diff(json3, json4)
fmt.Printf("有差异: %v\n", hasDiff) // true

// 不同的键
json5 := `{"name":"Alice"}`
json6 := `{"age":30}`
hasDiff, _ = j.Diff(json5, json6)
fmt.Printf("有差异: %v\n", hasDiff) // true

// 不同的结构
json7 := `{"name":"Alice"}`
json8 := `["Alice"]`
hasDiff, _ = j.Diff(json7, json8)
fmt.Printf("有差异: %v\n", hasDiff) // true

// 数组顺序不同
json9 := `[1,2,3]`
json10 := `[3,2,1]`
hasDiff, _ = j.Diff(json9, json10)
fmt.Printf("有差异: %v\n", hasDiff) // true

// 验证数据转换是否保持一致性
original := `{"name":"Test","value":123}`
m, _ := j.ToMap(original)
converted, _ := j.FromMap(m)
hasDiff, _ = j.Diff(original, converted)
fmt.Printf("转换后一致: %v\n", !hasDiff) // true
```

### 7. 字符串转义与反转义

处理 JSON 字符串中的特殊字符。

```go
j := json.New()

// 转义特殊字符
original := "Hello\nWorld\t!"
escaped := j.Escape(original)
fmt.Printf("Escaped: %s\n", escaped)
// 输出：Hello\nWorld\t!

// 转义引号
quote := `He said "Hello"`
escaped = j.Escape(quote)
fmt.Printf("Escaped: %s\n", escaped)
// 输出：He said \"Hello\"

// 转义反斜杠
path := "C:\\Users\\test"
escaped = j.Escape(path)
fmt.Printf("Escaped: %s\n", escaped)
// 输出：C:\\\\Users\\\\test

// 反转义
unescaped, err := j.Unescape("Hello\\nWorld\\t!")
if err == nil {
    fmt.Printf("Unescaped: %s\n", unescaped)
    // 输出：
    // Hello
    // World    !
}

// Unicode 转义
unicodeEscaped := "\\u4e2d\\u6587"
unescaped, err = j.Unescape(unicodeEscaped)
if err == nil {
    fmt.Printf("Unescaped: %s\n", unescaped)
    // 输出：中文
}
```

## API 参考

### 基础验证和格式化

| 函数 | 说明 |
|------|------|
| `IsValid(jsonStr string) bool` | 验证 JSON 字符串是否有效 |
| `PrettyPrint(jsonStr, indent string) (string, error)` | 使用指定缩进格式化 JSON |
| `PrettyPrintWithIndent(jsonStr string) (string, error)` | 使用默认缩进（2 空格）格式化 |
| `Compact(jsonStr string) (string, error)` | 压缩 JSON，移除所有空白字符 |
| `Escape(str string) string` | 转义 JSON 字符串中的特殊字符 |
| `Unescape(str string) (string, error)` | 反转义 JSON 字符串 |

### 数据转换

| 函数 | 说明 |
|------|------|
| `ToMap(jsonStr string) (map[string]any, error)` | 将 JSON 转换为 map |
| `ToMapStrict(jsonStr string) (map[string]any, error)` | 严格模式转换（必须是对象） |
| `ToStruct(jsonStr string, target any) error` | 将 JSON 转换为结构体 |
| `FromMap(data map[string]any) (string, error)` | 将 map 转换为 JSON 字符串 |
| `FromStruct(data any) (string, error)` | 将结构体转换为 JSON 字符串 |
| `FromStructWithIndent(data any, indent string) (string, error)` | 使用缩进将结构体转为 JSON |

### 路径操作

| 函数 | 说明 |
|------|------|
| `GetValue(jsonStr, path string) (any, error)` | 根据路径获取任意类型值 |
| `GetString(jsonStr, path string) (string, error)` | 根据路径获取字符串值 |
| `GetFloat64(jsonStr, path string) (float64, error)` | 根据路径获取数字值 |
| `GetBool(jsonStr, path string) (bool, error)` | 根据路径获取布尔值 |

### 高级操作

| 函数 | 说明 |
|------|------|
| `Merge(jsonStr1, jsonStr2 string) (string, error)` | 合并两个 JSON 对象 |
| `Diff(jsonStr1, jsonStr2 string) (bool, error)` | 比较两个 JSON 的差异 |

### 实用工具

| 函数 | 说明 |
|------|------|
| `IsObject(jsonStr string) bool` | 检查 JSON 是否为对象类型 |
| `IsArray(jsonStr string) bool` | 检查 JSON 是否为数组类型 |
| `GetType(jsonStr, path string) (string, error)` | 获取值的类型 |
| `GetKeys(jsonStr, path string) ([]string, error)` | 获取对象的所有键 |
| `GetSize(jsonStr, path string) (int, error)` | 获取数组或对象的长度 |
| `Contains(jsonStr, path, key string) (bool, error)` | 检查对象是否包含指定键 |

## 使用场景

### 配置文件管理

```go
j := json.New()

// 合并默认配置和用户配置
defaultConfig := `{
    "app": {
        "name": "MyApp",
        "version": "1.0.0",
        "debug": false
    },
    "server": {
        "port": 8080,
        "timeout": 30
    }
}`

userConfig := `{
    "app": {
        "debug": true
    },
    "server": {
        "port": 9000
    }
}`

finalConfig, _ := j.Merge(defaultConfig, userConfig)
```

### API 响应处理

```go
j := json.New()

apiResponse := `{
    "status": "success",
    "data": {
        "user": {
            "id": 123,
            "name": "Alice",
            "roles": ["admin", "user"]
        }
    }
}`

// 提取嵌套数据
userName, _ := j.GetString(apiResponse, "data.user.name")
userRoles, _ := j.GetValue(apiResponse, "data.user.roles")
```

### 数据验证

```go
j := json.New()

// 验证配置文件格式
configData := readFile("configmgr.json")
if !j.IsValid(configData) {
    log.Fatal("Invalid configmgr file")
}

// 检查必需字段
if j.Contains(configData, "", "database") {
    dbType, _ := j.GetType(configData, "database")
    if dbType != "object" {
        log.Fatal("Invalid database configuration")
    }
}
```

### 数据转换

```go
j := json.New()

// 数据库记录转换为 JSON
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
}

user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}

// 转换为 JSON 用于 API 响应
jsonStr, _ := j.FromStructWithIndent(user, "  ")

// 转换回 Map 进行动态处理
dataMap, _ := j.ToMap(jsonStr)
```

## 错误处理

所有可能失败的操作都返回 error，建议进行错误处理：

```go
j := json.New()

jsonStr := `{"name":"Alice","age":30}`

// 正确的错误处理
name, err := j.GetString(jsonStr, "name")
if err != nil {
    fmt.Printf("获取失败: %v\n", err)
    return
}
fmt.Printf("Name: %s\n", name)

// 处理路径不存在的情况
value, err := j.GetValue(jsonStr, "nonexistent.path")
if err != nil {
    // 路径不存在或其他错误
    fmt.Printf("路径访问失败: %v\n", err)
}

// 处理无效 JSON
invalidJSON := `{invalid}`
if !j.IsValid(invalidJSON) {
    fmt.Println("JSON 格式无效")
}
```

## 性能建议

1. **重复使用实例**：JSON 工具实例是可重用的，避免频繁创建
2. **缓存解析结果**：对于频繁访问的 JSON，考虑缓存为 Map
3. **选择合适的方法**：简单场景使用 `GetString`/`GetFloat64`，复杂场景使用 `GetValue`
4. **避免重复转换**：一次转换后重复使用，而不是多次转换

## 注意事项

1. **路径访问限制**：当前不支持数组通配符或复杂表达式，仅支持简单索引
2. **类型转换**：`GetString` 会尝试将非字符串值转换为字符串格式
3. **合并策略**：`Merge` 仅支持对象类型，数组合并会直接覆盖
4. **浮点数精度**：JSON 中的数字会被解析为 `float64`，大整数可能丢失精度
5. **键顺序**：Go 的 map 不保证键的顺序，多次序列化结果可能不同

## 相关模块

- `util/map` - Map 操作工具
- `util/string` - 字符串处理工具
- `util/convert` - 类型转换工具

## 许可证

本项目采用 MIT 许可证。
