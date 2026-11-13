# util/rand

提供生成各种类型随机数和随机字符串的工具函数。

## 特性

- **加密安全的随机数** - 基于 `crypto/rand` 实现，适用于安全敏感场景
- **丰富的随机数类型** - 支持整数（int/int64）、浮点数、布尔值等多种数据类型
- **灵活的随机字符串生成** - 支持自定义字符集、字母、数字、大小写字母等多种组合
- **UUID v4 生成** - 符合 RFC 4122 标准的 UUID 生成
- **泛型随机选择** - 使用 Go 泛型实现的随机选择函数，支持任意类型切片
- **完善的错误处理** - 内置回退机制，确保在极端情况下也能正常工作
- **自动范围处理** - 自动处理 min > max 的反转范围情况

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

	// 生成随机字符串（长度32，字母数字混合）
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

生成指定范围内的随机整数 [min, max]，自动处理范围反转。

```go
// 正常范围
num := rand.Rand.RandomInt(1, 100)
fmt.Println(num)

// 负数范围
num := rand.Rand.RandomInt(-100, 100)
fmt.Println(num)

// 自动处理反转的范围（min > max）
num := rand.Rand.RandomInt(100, 1) // 等同于 RandomInt(1, 100)
fmt.Println(num)

// 相同值时返回该值
num := rand.Rand.RandomInt(50, 50)
fmt.Println(num) // 输出: 50
```

### RandomInt64

生成 int64 类型的随机整数 [min, max]，适用于大范围数值。

```go
// 大范围
num := rand.Rand.RandomInt64(0, 1000000000)
fmt.Println(num)

// 负数范围
num := rand.Rand.RandomInt64(-1000000000, 1000000000)
fmt.Println(num)

// 极大范围（接近 int64 边界）
bigNum := rand.Rand.RandomInt64(-9223372036854775807, 9223372036854775807)
fmt.Println(bigNum)
```

### RandomFloat

生成指定范围内的随机浮点数 [min, max)，左闭右开区间。

```go
// [0.0, 1.0) 范围
value := rand.Rand.RandomFloat(0.0, 1.0)
fmt.Println(value)

// 负数范围
temperature := rand.Rand.RandomFloat(-10.0, 40.0)
fmt.Printf("温度: %.1f°C\n", temperature)

// 百分比生成
percent := rand.Rand.RandomFloat(0.0, 100.0)
fmt.Printf("完成度: %.2f%%\n", percent)
```

### RandomBool

生成随机的布尔值，理论上 true 和 false 的出现概率各为 50%。

```go
isSuccess := rand.Rand.RandomBool()
fmt.Println(isSuccess)

// 模拟抛硬币
if rand.Rand.RandomBool() {
	fmt.Println("正面")
} else {
	fmt.Println("反面")
}

// 随机决定某个行为
if rand.Rand.RandomBool() {
	fmt.Println("执行方案A")
} else {
	fmt.Println("执行方案B")
}
```

## 随机字符串生成

### RandomString

生成字母数字组合的随机字符串，字符集包含大小写字母和数字。

```go
// 生成 32 位随机字符串
token := rand.Rand.RandomString(32)
fmt.Println(token)

// 生成短 token（适用于 session ID）
shortToken := rand.Rand.RandomString(16)
fmt.Println(shortToken)

// 生成验证码（包含字母和数字）
captcha := rand.Rand.RandomString(6)
fmt.Println(captcha)

// 生成临时密码
tempPassword := rand.Rand.RandomString(12)
fmt.Println(tempPassword)
```

### RandomStringFromCharset

从自定义字符集生成随机字符串，灵活性最高。

```go
// 自定义字符集（特殊符号）
customCharset := "!@#$%^&*()"
password := rand.Rand.RandomStringFromCharset(16, customCharset)
fmt.Println(password)

// 生成十六进制字符串（用于颜色、哈希等）
hexCharset := "0123456789abcdef"
hexString := rand.Rand.RandomStringFromCharset(32, hexCharset)
fmt.Println(hexString)

// 生成 Base64 字符（排除 + 和 /）
base64Charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
base64String := rand.Rand.RandomStringFromCharset(24, base64Charset)
fmt.Println(base64String)

// 使用预定义字符集常量
hexString = rand.Rand.RandomStringFromCharset(16, rand.CharsetDigits)
fmt.Println(hexString)
```

### RandomLetters

生成纯字母的随机字符串，包含大小写字母。

```go
// 纯字母字符串（适合随机 ID、命名等）
letters := rand.Rand.RandomLetters(20)
fmt.Println(letters)

// 生成短码（用于优惠码等）
promoCode := rand.Rand.RandomLetters(8)
fmt.Println(promoCode)
```

### RandomDigits

生成纯数字的随机字符串，适用于验证码、随机数等场景。

```go
// 纯数字字符串（适合验证码）
digits := rand.Rand.RandomDigits(6)
fmt.Println(digits)

// 生成随机手机号（示例）
prefix := "138"
suffix := rand.Rand.RandomDigits(8)
phoneNumber := prefix + suffix
fmt.Println(phoneNumber)

// 生成随机订单号
orderNum := rand.Rand.RandomDigits(12)
fmt.Println(orderNum)
```

### RandomLowercase

生成纯小写字母的随机字符串。

```go
// 纯小写字母字符串
lower := rand.Rand.RandomLowercase(15)
fmt.Println(lower)

// 生成随机用户名
username := rand.Rand.RandomLowercase(10)
fmt.Println(username)
```

### RandomUppercase

生成纯大写字母的随机字符串。

```go
// 纯大写字母字符串
upper := rand.Rand.RandomUppercase(15)
fmt.Println(upper)

// 生成随机产品代码
productCode := "PROD-" + rand.Rand.RandomUppercase(6)
fmt.Println(productCode)
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

从切片中随机选择一个元素（泛型函数），支持任意类型。

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

// 浮点数切片
prices := []float64{9.99, 19.99, 29.99, 39.99}
randomPrice := rand.RandomChoice(prices)
fmt.Printf("随机价格: %.2f\n", randomPrice)

// 空切片返回零值
emptyResult := rand.RandomChoice([]int{})
fmt.Println(emptyResult) // 输出: 0
```

### RandomChoices

从切片中随机选择多个不重复的元素（泛型函数），使用 Fisher-Yates 洗牌算法。

```go
// 选择多个不重复元素
numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
selected := rand.RandomChoices(numbers, 3)
fmt.Println(selected) // 例如输出: [5 2 9]

// 选择所有元素（会打乱顺序，实现洗牌效果）
items := []string{"A", "B", "C"}
shuffled := rand.RandomChoices(items, 3)
fmt.Println(shuffled)

// 请求数量超过切片长度时返回所有元素（打乱后）
result := rand.RandomChoices([]int{1, 2, 3}, 10)
fmt.Println(result) // 输出: [1,2,3] 的某种排列

// 字符串切片选择
colors := []string{"red", "green", "blue", "yellow", "purple"}
selectedColors := rand.RandomChoices(colors, 2)
fmt.Println(selectedColors)

// 空切片返回 nil
emptyResult := rand.RandomChoices([]int{}, 3)
fmt.Println(emptyResult) // 输出: <nil>

// count 为 0 或负数时返回 nil
result = rand.RandomChoices([]int{1, 2, 3}, 0)
fmt.Println(result) // 输出: <nil>
```

## API

### 随机数生成

| 函数 | 说明 | 返回值 | 闭区间 |
|------|------|--------|--------|
| `RandomInt(min, max int) int` | 生成 [min, max] 范围内的随机整数 | int | 是 |
| `RandomInt64(min, max int64) int64` | 生成 [min, max] 范围内的随机 int64 整数 | int64 | 是 |
| `RandomFloat(min, max float64) float64` | 生成 [min, max) 范围内的随机浮点数 | float64 | 左闭右开 |
| `RandomBool() bool` | 生成随机布尔值（50% 概率） | bool | - |

**特性：**
- 自动处理范围反转（min > max 时自动交换）
- min == max 时返回该值
- 使用 `crypto/rand` 确保密码学安全

### 随机字符串生成

| 函数 | 说明 | 返回值 | 字符集 |
|------|------|--------|--------|
| `RandomString(length int) string` | 生成指定长度的字母数字随机字符串 | string | `CharsetAlphanumeric` |
| `RandomStringFromCharset(length int, charset string) string` | 从自定义字符集生成随机字符串 | string | 自定义 |
| `RandomLetters(length int) string` | 生成纯字母随机字符串 | string | `CharsetLetters` |
| `RandomDigits(length int) string` | 生成纯数字随机字符串 | string | `CharsetDigits` |
| `RandomLowercase(length int) string` | 生成纯小写字母随机字符串 | string | `CharsetLowercase` |
| `RandomUppercase(length int) string` | 生成纯大写字母随机字符串 | string | `CharsetUppercase` |

**特性：**
- length ≤ 0 或 charset 为空时返回空字符串
- 使用 `crypto/rand` 确保字符选择的密码学安全

### UUID 生成

| 函数 | 说明 | 返回值 | 格式 |
|------|------|--------|------|
| `RandomUUID() string` | 生成 UUID v4 格式的唯一标识符 | string | `xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx` |

**特性：**
- 符合 RFC 4122 标准的 UUID v4
- 16 字节随机生成 + 版本/变体标记
- 格式：`8-4-4-4-12`，共 36 个字符（含 4 个连字符）

### 随机选择（泛型）

| 函数 | 说明 | 返回值 | 特性 |
|------|------|--------|------|
| `RandomChoice[T any](options []T) T` | 从切片中随机选择一个元素 | T | 空切片返回零值 |
| `RandomChoices[T any](options []T, count int) []T` | 从切片中随机选择多个不重复元素 | []T | Fisher-Yates 洗牌算法 |

**特性：**
- 支持任意类型 T（int、string、struct、指针等）
- `RandomChoices` 确保结果中元素不重复
- `RandomChoices` 当 count ≥ len(options) 时返回所有元素（打乱顺序）
- 空切片或 count ≤ 0 时返回 nil（对于 `RandomChoices`）或零值（对于 `RandomChoice`）

### 预定义字符集常量

| 常量 | 值 | 说明 |
|------|-----|------|
| `CharsetAlphanumeric` | `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789` | 字母数字混合 |
| `CharsetLetters` | `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ` | 大小写字母 |
| `CharsetDigits` | `0123456789` | 数字 0-9 |
| `CharsetLowercase` | `abcdefghijklmnopqrstuvwxyz` | 小写字母 a-z |
| `CharsetUppercase` | `ABCDEFGHIJKLMNOPQRSTUVWXYZ` | 大写字母 A-Z |

### 全局实例

| 变量 | 类型 | 说明 |
|------|------|------|
| `Rand` | `*randEngine` | 默认的随机数操作实例（单例） |

## 随机性说明

### 安全性

本包使用 `crypto/rand` 作为随机数源，提供了密码学安全的随机数生成：

- **不可预测性**: 生成的随机数无法被预测，符合密码学安全标准
- **适用场景**: 适用于生成 token、密钥、会话 ID、验证码等安全敏感的数据
- **性能考虑**: 相比 `math/rand` 性能稍低，但提供了更高的安全性
- **使用建议**: 非安全敏感的随机数（如测试数据、UI 随机效果）可考虑使用 `math/rand`

### 回退机制

当 `crypto/rand` 不可用时（极端情况下），包内实现了回退机制以确保程序继续运行：

```go
// 示例：RandomInt 的回退逻辑
nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
if err != nil {
	// 如果加密随机数失败，回退到简单的伪随机数
	return min + int(float64(max-min)*0.5)
}
```

**回退策略：**
- `RandomInt`: 返回范围中点值
- `RandomInt64`: 返回 min 值
- `RandomFloat`: 返回 min 值
- `RandomBool`: 通过 `RandomInt(0, 1)` 回退
- `RandomStringFromCharset`: 使用字符集顺序循环
- `RandomUUID`: 使用 `RandomInt` 逐字节生成

## 性能

本包的性能基准测试结果（仅供参考，实际性能取决于硬件）：

| 函数 | 操作 | 说明 |
|------|------|------|
| `BenchmarkRandomInt` | 随机整数 | 单次操作约几微秒级 |
| `BenchmarkRandomInt64` | 随机 int64 整数 | 与 `RandomInt` 相当 |
| `BenchmarkRandomFloat` | 随机浮点数 | 单次操作约几微秒级 |
| `BenchmarkRandomString` | 随机字符串（20字符） | 与长度成正比 |
| `BenchmarkRandomUUID` | UUID 生成 | 单次操作约几微秒级 |
| `BenchmarkRandomChoice` | 随机选择一个 | 与切片长度关系较小 |
| `BenchmarkRandomChoices` | 随机选择多个 | 与 count 成正比 |

**性能优化建议：**
- 批量生成随机数时考虑一次性生成多次调用
- 对于大规模随机字符串生成，可考虑预分配 buffer
- 非安全敏感场景且对性能要求极高时，可考虑使用 `math/rand`

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

# 运行指定测试
go test ./util/rand -run TestRandomInt

# 运行指定性能测试
go test ./util/rand -bench=BenchmarkRandomInt
```

**测试覆盖：**
- 范围测试（正常范围、反转范围、负数范围、相同值）
- 边界测试（空切片、零长度、极大值）
- 随机性测试（多次调用验证结果多样性）
- 唯一性测试（UUID 唯一性、RandomChoices 不重复）
- 格式验证（UUID 格式、字符集验证）
