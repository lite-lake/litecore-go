# util/rand

util/rand 是 litecore-go 的随机数生成工具包，提供了生成各种类型随机数和随机字符串的功能。

## 特性

- **加密安全的随机数** - 基于 `crypto/rand` 实现，适用于安全敏感场景
- **丰富的随机数类型** - 支持整数、浮点数、布尔值等多种数据类型
- **灵活的随机字符串生成** - 支持自定义字符集、字母、数字等多种组合
- **UUID v4 生成** - 符合 RFC 4122 标准的 UUID 生成
- **泛型随机选择** - 使用 Go 泛型实现的随机选择函数，支持任意类型
- **完善的错误处理** - 内置回退机制，确保在极端情况下也能正常工作

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/rand"
)

func main() {
    // 生成随机整数 [1, 100]
    randomInt := rand.Rand.RandomInt(1, 100)
    fmt.Printf("随机整数: %d\n", randomInt)

    // 生成随机浮点数 [0.0, 1.0)
    randomFloat := rand.Rand.RandomFloat(0.0, 1.0)
    fmt.Printf("随机浮点数: %.4f\n", randomFloat)

    // 生成随机字符串（长度32）
    randomString := rand.Rand.RandomString(32)
    fmt.Printf("随机字符串: %s\n", randomString)

    // 生成 UUID
    uuid := rand.Rand.RandomUUID()
    fmt.Printf("UUID: %s\n", uuid)

    // 从切片中随机选择一个元素
    fruits := []string{"苹果", "香蕉", "橙子", "葡萄"}
    chosen := rand.RandomChoice(fruits)
    fmt.Printf("随机选择的水果: %s\n", chosen)

    // 从切片中随机选择多个元素（不重复）
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    selected := rand.RandomChoices(numbers, 3)
    fmt.Printf("随机选择的数字: %v\n", selected)
}
```

### 完整示例

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/rand"
)

func main() {
    // 模拟生成用户数据
    userID := rand.Rand.RandomUUID()
    username := rand.Rand.RandomString(12)
    age := rand.Rand.RandomInt(18, 65)
    score := rand.Rand.RandomFloat(0.0, 100.0)
    isActive := rand.Rand.RandomBool()

    fmt.Printf("用户ID: %s\n", userID)
    fmt.Printf("用户名: %s\n", username)
    fmt.Printf("年龄: %d\n", age)
    fmt.Printf("分数: %.2f\n", score)
    fmt.Printf("是否激活: %t\n", isActive)

    // 随机分配用户角色
    roles := []string{"管理员", "编辑", "访客", "开发者"}
    userRole := rand.RandomChoice(roles)
    fmt.Printf("用户角色: %s\n", userRole)
}
```

## 功能详解

### 随机数生成

#### RandomInt

生成指定范围内的随机整数（包含边界值）。

```go
// 生成 [1, 100] 范围内的随机整数
num := rand.Rand.RandomInt(1, 100)
fmt.Println(num)  // 输出: 42

// 支持负数范围
num := rand.Rand.RandomInt(-100, 100)
fmt.Println(num)  // 输出: -37

// 自动处理反转的范围（min > max）
num := rand.Rand.RandomInt(100, 1)
fmt.Println(num)  // 输出: 23

// 相同值时返回该值
num := rand.Rand.RandomInt(50, 50)
fmt.Println(num)  // 输出: 50
```

#### RandomInt64

生成 int64 类型的随机整数，适用于大范围数值。

```go
// 生成大范围内的随机数
num := rand.Rand.RandomInt64(0, 1000000000)
fmt.Println(num)  // 输出: 542876193

// 生成负数范围
num := rand.Rand.RandomInt64(-1000000000, 1000000000)
fmt.Println(num)  // 输出: -287654321
```

#### RandomFloat

生成指定范围内的随机浮点数（左闭右开区间 [min, max)）。

```go
// 生成 [0.0, 1.0) 范围内的随机浮点数
value := rand.Rand.RandomFloat(0.0, 1.0)
fmt.Println(value)  // 输出: 0.6234

// 生成指定精度的小数
price := rand.Rand.RandomFloat(10.0, 100.0)
fmt.Printf("价格: %.2f\n", price)  // 输出: 价格: 56.78

// 生成负数范围
temperature := rand.Rand.RandomFloat(-10.0, 40.0)
fmt.Printf("温度: %.1f°C\n", temperature)  // 输出: 温度: 23.5°C
```

#### RandomBool

生成随机的布尔值。

```go
// 生成随机布尔值
isSuccess := rand.Rand.RandomBool()
fmt.Println(isSuccess)  // 输出: true

// 模拟抛硬币
coin := rand.Rand.RandomBool()
if coin {
    fmt.Println("正面")
} else {
    fmt.Println("反面")
}
```

### 随机字符串生成

#### RandomString

生成字母数字组合的随机字符串。

```go
// 生成 32 位随机字符串（默认字母数字）
token := rand.Rand.RandomString(32)
fmt.Println(token)  // 输出: aB3xY9mK2pL7qW4rT6uV8zX1cD5eF0gH

// 生成短 token
shortToken := rand.Rand.RandomString(16)
fmt.Println(shortToken)  // 输出: xY9mK2pL7qW4rT6u

// 生成验证码
captcha := rand.Rand.RandomString(6)
fmt.Println(captcha)  // 输出: 4aB7xY
```

#### RandomStringFromCharset

从自定义字符集生成随机字符串。

```go
// 自定义字符集
customCharset := "!@#$%^&*()_+-=[]{}|;:,.<>?"
password := rand.Rand.RandomStringFromCharset(16, customCharset)
fmt.Println(password)  // 输出: @#$%^&*()_+-=[]{}|

// 生成十六进制字符串
hexCharset := "0123456789abcdef"
hexString := rand.Rand.RandomStringFromCharset(32, hexCharset)
fmt.Println(hexString)  // 输出: 3a7f9c2b4e8d1f6a5c9b2e7d4f8a1c3b

// 生成 Base64 字符串（简化版）
base64Charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
base64String := rand.Rand.RandomStringFromCharset(24, base64Charset)
fmt.Println(base64String)  // 输出: YWJjZGVmZ2hpamtsbW5vcHFy
```

#### RandomLetters

生成纯字母的随机字符串。

```go
// 生成纯字母字符串
letters := rand.Rand.RandomLetters(20)
fmt.Println(letters)  // 输出: aBxYmKpLqWrTuVzXcD

// 生成随机名称
name := rand.Rand.RandomLetters(10)
fmt.Println(name)  // 输出: jKpQrStUvW
```

#### RandomDigits

生成纯数字的随机字符串。

```go
// 生成纯数字字符串
digits := rand.Rand.RandomDigits(12)
fmt.Println(digits)  // 输出: 528764913025

// 生成随机手机号（示例）
phoneNumber := "138" + rand.Rand.RandomDigits(8)
fmt.Println(phoneNumber)  // 输出: 13852876491

// 生成验证码
code := rand.Rand.RandomDigits(6)
fmt.Println(code)  // 输出: 528764
```

#### RandomLowercase / RandomUppercase

生成纯小写或纯大写的随机字符串。

```go
// 生成小写字母字符串
lower := rand.Rand.RandomLowercase(15)
fmt.Println(lower)  // 输出: abcdefghijklmno

// 生成大写字母字符串
upper := rand.Rand.RandomUppercase(15)
fmt.Println(upper)  // 输出: ABCDEFGHIJKLMNO

// 组合使用
username := rand.Rand.RandomLowercase(8)
fmt.Printf("用户名: %s\n", username)  // 输出: 用户名: abcdefgh
```

### UUID 生成

#### RandomUUID

生成符合 UUID v4 标准的唯一标识符。

```go
// 生成 UUID
uuid := rand.Rand.RandomUUID()
fmt.Println(uuid)  // 输出: 550e8400-e29b-41d4-a716-446655440000

// 用于生成唯一 ID
userID := rand.Rand.RandomUUID()
sessionID := rand.Rand.RandomUUID()
requestID := rand.Rand.RandomUUID()

fmt.Printf("用户ID: %s\n", userID)
fmt.Printf("会话ID: %s\n", sessionID)
fmt.Printf("请求ID: %s\n", requestID)
```

生成的 UUID 格式：`xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`

- x: 随机十六进制数字
- 4: UUID 版本号（v4）
- y: UUID 变体标识（8, 9, a, 或 b）

### 随机选择

#### RandomChoice

从切片中随机选择一个元素（泛型函数）。

```go
// 从字符串切片中选择
fruits := []string{"苹果", "香蕉", "橙子", "葡萄"}
selected := rand.RandomChoice(fruits)
fmt.Println(selected)  // 输出: 香蕉

// 从整数切片中选择
numbers := []int{10, 20, 30, 40, 50}
chosen := rand.RandomChoice(numbers)
fmt.Println(chosen)  // 输出: 30

// 从结构体切片中选择
type User struct {
    Name string
    Age  int
}
users := []User{
    {Name: "张三", Age: 25},
    {Name: "李四", Age: 30},
    {Name: "王五", Age: 35},
}
luckyUser := rand.RandomChoice(users)
fmt.Printf("幸运用户: %s (%d岁)\n", luckyUser.Name, luckyUser.Age)

// 随机选择颜色
colors := []string{"红色", "蓝色", "绿色", "黄色", "紫色"}
themeColor := rand.RandomChoice(colors)
fmt.Printf("主题颜色: %s\n", themeColor)
```

#### RandomChoices

从切片中随机选择多个不重复的元素（泛型函数）。

```go
// 从数字列表中选择 3 个不重复的数字
numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
selected := rand.RandomChoices(numbers, 3)
fmt.Println(selected)  // 输出: [3 7 1]

// 从用户列表中选择幸运用户
users := []string{"张三", "李四", "王五", "赵六", "钱七", "孙八"}
winners := rand.RandomChoices(users, 2)
fmt.Printf("获奖用户: %v\n", winners)  // 输出: 获奖用户: [李四 孙八]

// 选择所有元素（会打乱顺序）
items := []string{"A", "B", "C"}
shuffled := rand.RandomChoices(items, 3)
fmt.Println(shuffled)  // 输出: [B A C] 或其他排列

// 请求数量超过切片长度时返回所有元素
few := []int{1, 2, 3}
result := rand.RandomChoices(few, 10)
fmt.Println(result)  // 输出: [3 1 2] 或其他排列

// 生成随机抽奖序列
participants := []string{
    "参与者1", "参与者2", "参与者3", "参与者4",
    "参与者5", "参与者6", "参与者7", "参与者8",
}
// 一等奖 1 名
firstPrize := rand.RandomChoices(participants, 1)
// 二等奖 3 名
secondPrize := rand.RandomChoices(participants, 3)
fmt.Printf("一等奖: %v\n", firstPrize)
fmt.Printf("二等奖: %v\n", secondPrize)
```

## API 参考

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
    // 回退到简单的实现
    return min + int(float64(max-min)*0.5)
}
```

### Fisher-Yates 洗牌算法

`RandomChoices` 函数在需要选择所有元素时使用 Fisher-Yates 洗牌算法：

```go
// Fisher-Yates 洗牌算法
for i := len(result) - 1; i > 0; i-- {
    j := randOp.RandomInt(0, i)
    result[i], result[j] = result[j], result[i]
}
```

这保证了：
- 均匀的随机分布
- O(n) 时间复杂度
- 原地洗牌，无需额外空间

## 使用建议

### 安全性考虑

```go
// ✅ 推荐：用于生成 token、密钥等敏感数据
token := rand.Rand.RandomString(32)
apiKey := rand.Rand.RandomUUID()
sessionID := rand.Rand.RandomString(64)

// ✅ 推荐：用于生成验证码
captcha := rand.Rand.RandomDigits(6)

// ⚠️ 注意：如需更高性能且不要求密码学安全，可考虑 math/rand
```

### 性能考虑

```go
// 对于大量随机数生成的场景
for i := 0; i < 10000; i++ {
    // 使用实例方法，避免重复创建对象
    num := rand.Rand.RandomInt(1, 100)
    _ = num
}

// 预分配切片容量
results := make([]string, 0, 1000)
for i := 0; i < 1000; i++ {
    results = append(results, rand.Rand.RandomUUID())
}
```

### 错误处理

```go
// 处理空切片情况
empty := []int{}
result := rand.RandomChoice(empty)
fmt.Println(result)  // 输出: 0 (零值)

// 处理无效长度
invalidString := rand.Rand.RandomString(-1)
fmt.Println(invalidString)  // 输出: (空字符串)

// 处理空字符集
invalidCharset := rand.Rand.RandomStringFromCharset(10, "")
fmt.Println(invalidCharset)  // 输出: (空字符串)
```

## 常见用例

### 生成 API Token

```go
func GenerateAPIToken() string {
    return rand.Rand.RandomString(64)
}

token := GenerateAPIToken()
fmt.Printf("API Token: %s\n", token)
```

### 生成临时密码

```go
func GenerateTempPassword(length int) string {
    charset := "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789!@#$%"
    return rand.Rand.RandomStringFromCharset(length, charset)
}

password := GenerateTempPassword(16)
fmt.Printf("临时密码: %s\n", password)
```

### 随机抽样

```go
// 从大量数据中随机抽样
population := make([]int, 1000)
for i := range population {
    population[i] = i
}

// 随机选择 100 个样本
samples := rand.RandomChoices(population, 100)
fmt.Printf("样本数量: %d\n", len(samples))
```

### 随机分配

```go
// 随机分配任务
tasks := []string{"任务A", "任务B", "任务C", "任务D"}
workers := []string{"工人1", "工人2", "工人3", "工人4"}

assignments := make(map[string]string)
for _, task := range tasks {
    worker := rand.RandomChoice(workers)
    assignments[task] = worker
}

for task, worker := range assignments {
    fmt.Printf("%s -> %s\n", task, worker)
}
```

### 生成测试数据

```go
type User struct {
    ID       string
    Username string
    Email    string
    Age      int
    Active   bool
}

func GenerateRandomUser() User {
    domains := []string{"gmail.com", "yahoo.com", "hotmail.com"}
    return User{
        ID:       rand.Rand.RandomUUID(),
        Username: rand.Rand.RandomLowercase(10),
        Email:    rand.Rand.RandomLowercase(8) + "@" + rand.RandomChoice(domains),
        Age:      rand.Rand.RandomInt(18, 65),
        Active:   rand.Rand.RandomBool(),
    }
}

user := GenerateRandomUser()
fmt.Printf("用户: %+v\n", user)
```

## 运行测试

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

## 许可证

本包是 litecore-go 项目的一部分，遵循项目许可证。
