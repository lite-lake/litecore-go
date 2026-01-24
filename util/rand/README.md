# util/rand

提供生成各种类型随机数和随机字符串的工具函数。

## 特性

- **加密安全的随机数** - 基于 `crypto/rand` 实现，适用于安全敏感场景
- **丰富的随机数类型** - 支持整数、浮点数、布尔值等多种数据类型
- **灵活的随机字符串生成** - 支持自定义字符集、字母、数字等多种组合
- **UUID v4 生成** - 符合 RFC 4122 标准的 UUID 生成
- **泛型随机选择** - 使用 Go 泛型实现的随机选择函数，支持任意类型
- **完善的错误处理** - 内置回退机制，确保在极端情况下也能正常工作

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/rand"
)

func main() {
    // 生成随机整数 [1, 100]
    num := rand.Rand.RandomInt(1, 100)
    fmt.Printf("随机整数: %d\n", num)

    // 生成随机浮点数 [0.0, 1.0)
    value := rand.Rand.RandomFloat(0.0, 1.0)
    fmt.Printf("随机浮点数: %.4f\n", value)

    // 生成随机字符串（长度32）
    token := rand.Rand.RandomString(32)
    fmt.Printf("随机字符串: %s\n", token)

    // 生成 UUID
    uuid := rand.Rand.RandomUUID()
    fmt.Printf("UUID: %s\n", uuid)

    // 从切片中随机选择一个元素
    fruits := []string{"apple", "banana", "orange"}
    selected := rand.RandomChoice(fruits)
    fmt.Printf("随机选择: %s\n", selected)

    // 从切片中随机选择多个元素（不重复）
    numbers := []int{1, 2, 3, 4, 5}
    picks := rand.RandomChoices(numbers, 3)
    fmt.Printf("随机选择多个: %v\n", picks)
}
```

## 随机数生成

### RandomInt

生成指定范围内的随机整数 [min, max]。

```go
// 正常范围
num := rand.Rand.RandomInt(1, 100)
fmt.Println(num)

// 负数范围
num := rand.Rand.RandomInt(-100, 100)
fmt.Println(num)

// 自动处理反转的范围
num := rand.Rand.RandomInt(100, 1)
fmt.Println(num)

// 相同值时返回该值
num := rand.Rand.RandomInt(50, 50)
fmt.Println(num)
```

### RandomInt64

生成 int64 类型的随机整数，适用于大范围数值。

```go
// 大范围
num := rand.Rand.RandomInt64(0, 1000000000)
fmt.Println(num)

// 负数范围
num := rand.Rand.RandomInt64(-1000000000, 1000000000)
fmt.Println(num)
```

### RandomFloat

生成指定范围内的随机浮点数 [min, max)。

```go
// [0.0, 1.0) 范围
value := rand.Rand.RandomFloat(0.0, 1.0)
fmt.Println(value)

// 负数范围
temperature := rand.Rand.RandomFloat(-10.0, 40.0)
fmt.Printf("温度: %.1f°C\n", temperature)
```

### RandomBool

生成随机的布尔值。

```go
isSuccess := rand.Rand.RandomBool()
fmt.Println(isSuccess)

// 模拟抛硬币
if rand.Rand.RandomBool() {
    fmt.Println("正面")
} else {
    fmt.Println("反面")
}
```

## 随机字符串生成

### RandomString

生成字母数字组合的随机字符串。

```go
// 生成 32 位随机字符串
token := rand.Rand.RandomString(32)
fmt.Println(token)

// 生成短 token
shortToken := rand.Rand.RandomString(16)
fmt.Println(shortToken)

// 生成验证码
captcha := rand.Rand.RandomString(6)
fmt.Println(captcha)
```

### RandomStringFromCharset

从自定义字符集生成随机字符串。

```go
// 自定义字符集
customCharset := "!@#$%^&*()"
password := rand.Rand.RandomStringFromCharset(16, customCharset)
fmt.Println(password)

// 生成十六进制字符串
hexCharset := "0123456789abcdef"
hexString := rand.Rand.RandomStringFromCharset(32, hexCharset)
fmt.Println(hexString)
```

### RandomLetters / RandomDigits / RandomLowercase / RandomUppercase

生成特定类型的随机字符串。

```go
// 纯字母字符串
letters := rand.Rand.RandomLetters(20)
fmt.Println(letters)

// 纯数字字符串
digits := rand.Rand.RandomDigits(12)
fmt.Println(digits)

// 纯小写字母字符串
lower := rand.Rand.RandomLowercase(15)
fmt.Println(lower)

// 纯大写字母字符串
upper := rand.Rand.RandomUppercase(15)
fmt.Println(upper)
```

## UUID 生成

### RandomUUID

生成符合 UUID v4 标准的唯一标识符。

```go
uuid := rand.Rand.RandomUUID()
fmt.Println(uuid)

// 用于生成唯一 ID
userID := rand.Rand.RandomUUID()
sessionID := rand.Rand.RandomUUID()
fmt.Printf("用户ID: %s\n", userID)
fmt.Printf("会话ID: %s\n", sessionID)
```

生成的 UUID 格式：`xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`

## 随机选择

### RandomChoice

从切片中随机选择一个元素（泛型函数）。

```go
// 字符串切片
fruits := []string{"apple", "banana", "orange"}
selected := rand.RandomChoice(fruits)
fmt.Println(selected)

// 整数切片
numbers := []int{10, 20, 30, 40, 50}
chosen := rand.RandomChoice(numbers)
fmt.Println(chosen)

// 结构体切片
type User struct {
    Name string
    Age  int
}
users := []User{
    {Name: "张三", Age: 25},
    {Name: "李四", Age: 30},
}
luckyUser := rand.RandomChoice(users)
fmt.Printf("幸运用户: %s\n", luckyUser.Name)
```

### RandomChoices

从切片中随机选择多个不重复的元素（泛型函数）。

```go
// 选择多个不重复元素
numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
selected := rand.RandomChoices(numbers, 3)
fmt.Println(selected)

// 选择所有元素（会打乱顺序）
items := []string{"A", "B", "C"}
shuffled := rand.RandomChoices(items, 3)
fmt.Println(shuffled)

// 请求数量超过切片长度时返回所有元素
result := rand.RandomChoices([]int{1, 2, 3}, 10)
fmt.Println(result)
```

## API

### 随机数生成

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `RandomInt(min, max int) int` | 生成 [min, max] 范围内的随机整数 | int |
| `RandomInt64(min, max int64) int64` | 生成 [min, max] 范围内的随机 int64 整数 | int64 |
| `RandomFloat(min, max float64) float64` | 生成 [min, max) 范围内的随机浮点数 | float64 |
| `RandomBool() bool` | 生成随机布尔值 | bool |

### 随机字符串生成

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `RandomString(length int) string` | 生成指定长度的字母数字随机字符串 | string |
| `RandomStringFromCharset(length int, charset string) string` | 从自定义字符集生成随机字符串 | string |
| `RandomLetters(length int) string` | 生成纯字母随机字符串 | string |
| `RandomDigits(length int) string` | 生成纯数字随机字符串 | string |
| `RandomLowercase(length int) string` | 生成纯小写字母随机字符串 | string |
| `RandomUppercase(length int) string` | 生成纯大写字母随机字符串 | string |

### UUID 生成

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `RandomUUID() string` | 生成 UUID v4 格式的唯一标识符 | string |

### 随机选择

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `RandomChoice[T any](options []T) T` | 从切片中随机选择一个元素 | T |
| `RandomChoices[T any](options []T, count int) []T` | 从切片中随机选择多个不重复元素 | []T |

### 预定义字符集常量

| 常量 | 值 |
|------|-----|
| `CharsetAlphanumeric` | `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789` |
| `CharsetLetters` | `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ` |
| `CharsetDigits` | `0123456789` |
| `CharsetLowercase` | `abcdefghijklmnopqrstuvwxyz` |
| `CharsetUppercase` | `ABCDEFGHIJKLMNOPQRSTUVWXYZ` |

## 随机性说明

### 安全性

本包使用 `crypto/rand` 作为随机数源，提供了密码学安全的随机数生成：

- **不可预测性**: 生成的随机数无法被预测
- **适用场景**: 适用于生成 token、密钥、会话 ID 等安全敏感的数据
- **性能考虑**: 相比 `math/rand` 性能稍低，但提供了更高的安全性

### 回退机制

当 `crypto/rand` 不可用时，包内实现了回退机制：

```go
// 示例：RandomInt 的回退逻辑
nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
if err != nil {
    return min + int(float64(max-min)*0.5)
}
```

## 测试

```bash
# 运行所有测试
go test ./util/rand

# 运行测试并显示覆盖率
go test ./util/rand -cover

# 运行性能测试
go test ./util/rand -bench=.

# 查看详细测试输出
go test ./util/rand -v
```
