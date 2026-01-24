# util/id

分布式唯一 ID 生成工具，提供 CUID2 风格的高性能、高可读性 ID 生成方案。

## 特性

- **时间有序性**：ID 前缀包含毫秒级时间戳，保证生成结果大致按时间排序，便于数据库索引优化
- **高唯一性**：结合时间戳和加密级随机数（16 字节），碰撞概率极低，适合大规模分布式环境
- **高可读性**：仅使用小写字母和数字（0-9, a-z），25 位固定长度，易于人类识别和传输
- **分布式安全**：无需中央协调机制，各节点可独立生成唯一 ID，避免单点故障和性能瓶颈
- **高性能**：使用 Base36 编码和位运算，单次生成耗时极低，支持高并发场景
- **URL 安全**：不包含特殊字符，可直接用于 URL 路径和参数

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/id"
)

func main() {
    // 生成 CUID2
    uid, err := id.NewCUID2()
    if err != nil {
        panic(err)
    }
    fmt.Println("生成的 ID:", uid)
    // 输出示例: 2j8c3qf9g4k5m6p7r8s9t0v1x

    // 批量生成多个 ID
    ids := make([]string, 5)
    for i := 0; i < 5; i++ {
        uid, err := id.NewCUID2()
        if err != nil {
            panic(err)
        }
        ids[i] = uid
        fmt.Printf("ID %d: %s\n", i+1, uid)
    }

    // 验证唯一性
    uniqueIDs := make(map[string]bool)
    for _, uid := range ids {
        uniqueIDs[uid] = true
    }
    fmt.Printf("生成了 %d 个唯一 ID\n", len(uniqueIDs))
}
```

## 核心功能

### 数据库主键

CUID2 的时间有序性使其非常适合作为数据库主键，特别是在 B+ 树索引中可减少页分裂。

```go
type User struct {
    ID        string `gorm:"primaryKey;type:varchar(25)"`
    Name      string `gorm:"type:varchar(100);not null"`
    Email     string `gorm:"type:varchar(255);uniqueIndex"`
    CreatedAt int64  `gorm:"index"`
}

func CreateUser(name, email string) (*User, error) {
    uid, err := id.NewCUID2()
    if err != nil {
        return nil, err
    }
    user := &User{
        ID:        uid,
        Name:      name,
        Email:     email,
        CreatedAt: time.Now().Unix(),
    }
    return user, nil
}
```

### 分布式追踪 ID

用于微服务架构中的请求追踪和分布式事务关联。

```go
type RequestContext struct {
    TraceID   string
    UserID    string
    RequestID string
}

func NewRequestContext(userID string) *RequestContext {
    traceID, err := id.NewCUID2()
    if err != nil {
        panic(err)
    }
    requestID, err := id.NewCUID2()
    if err != nil {
        panic(err)
    }
    return &RequestContext{
        TraceID:   traceID,
        UserID:    userID,
        RequestID: requestID,
    }
}
```

### 会话管理

生成用户会话或一次性访问令牌。

```go
type Session struct {
    SessionID   string
    UserID      string
    Data        map[string]interface{}
    ExpiresAt   time.Time
}

func CreateSession(userID string, duration time.Duration) *Session {
    sessionID, err := id.NewCUID2()
    if err != nil {
        panic(err)
    }
    return &Session{
        SessionID: sessionID,
        UserID:    userID,
        Data:      make(map[string]interface{}),
        ExpiresAt: time.Now().Add(duration),
    }
}
```

## API

### 函数

#### `func NewCUID2() (string, error)`

生成 CUID2 风格的唯一标识符。

**返回值**
- `string`: 25 个字符的小写字母数字字符串
- `error`: 生成失败时返回错误

**特性**
- 时间有序：前缀包含时间戳
- 高唯一性：碰撞概率极低
- 可读性：仅包含小写字母和数字
- 分布式安全：无需中央协调

## 性能基准

在标准硬件环境下的性能测试结果：

```
BenchmarkNewCUID2-8                 500000    2.8 µs/op    416 B/op    9 allocs/op
BenchmarkNewCUID2_Parallel-8       1000000    1.9 µs/op    416 B/op    9 allocs/op
BenchmarkNewCUID2_Batch-8            50000   32.5 µs/op   41600 B/op  900 allocs/op
```

## 实现细节

### 编码方式

使用 Base36 编码（0-9 + a-z）：
- 相比十六进制（Base16）更紧凑
- 不包含易混淆字符（0/O, 1/I/l）
- URL 安全，无需额外编码

### 随机数生成

使用 `crypto/rand` 生成 16 字节加密级随机数，提供 128 位熵值，确保在分布式环境中的唯一性。

### 时间排序

ID 前缀包含毫秒级 Unix 时间戳，保证生成的 ID 大致按时间排序，对数据库索引友好。

## 并发安全性

本实现使用 `crypto/rand.Read()` 生成随机数，该函数是并发安全的，可直接在高并发场景中使用。

```go
const goroutines = 100
const idsPerGoroutine = 100

var wg sync.WaitGroup
ids := make([]string, goroutines*idsPerGoroutine)

wg.Add(goroutines)
for i := 0; i < goroutines; i++ {
    go func(idx int) {
        defer wg.Done()
        for j := 0; j < idsPerGoroutine; j++ {
            uid, err := id.NewCUID2()
            if err != nil {
                continue
            }
            ids[idx*idsPerGoroutine+j] = uid
        }
    }(i)
}

wg.Wait()
```

## 最佳实践

### 数据库设计

- 使用 `VARCHAR(25)` 或 `CHAR(25)` 类型存储
- 固定长度类型（CHAR）性能更优
- 利用时间有序性优化查询性能

### 错误处理

```go
func ProcessRequest() error {
    requestID, err := id.NewCUID2()
    if err != nil {
        return fmt.Errorf("生成请求 ID 失败: %w", err)
    }

    log.Printf("[%s] 开始处理请求", requestID)

    if err := doSomething(); err != nil {
        return fmt.Errorf("请求 %s 失败: %w", requestID, err)
    }

    return nil
}
```

### 批量生成

如需批量生成 ID，可使用并发方式提高效率：

```go
func GenerateBatchID(count int) []string {
    ids := make([]string, count)
    var wg sync.WaitGroup

    batchSize := 100
    for i := 0; i < count; i += batchSize {
        end := i + batchSize
        if end > count {
            end = count
        }

        wg.Add(1)
        go func(start, end int) {
            defer wg.Done()
            for j := start; j < end; j++ {
                uid, err := id.NewCUID2()
                if err != nil {
                    continue
                }
                ids[j] = uid
            }
        }(i, end)
    }

    wg.Wait()
    return ids
}
```

## 常见问题

### Q: CUID2 适合哪些场景？

A: 适合数据库主键、分布式追踪 ID、会话 ID、订单号等需要唯一标识符的场景。不适合安全性敏感的令牌（如 JWT、API 密钥）。

### Q: CUID2 是全局唯一的吗？

A: 理论上存在极低碰撞概率。正常使用场景下（每秒百万级生成量），碰撞概率几乎为零。

### Q: 如何验证 CUID2 的有效性？

A: 可使用正则表达式验证格式：

```go
func IsValidCUID2(id string) bool {
    matched, _ := regexp.MatchString(`^[0-9a-z]{25}$`, id)
    return matched
}
```

### Q: CUID2 与 UUID v4 相比如何？

A: CUID2 长度更短（25 字符 vs 36 字符），时间有序适合数据库索引，可读性更高。两者碰撞概率都极低。

## 测试

```bash
# 运行所有测试
go test ./util/id

# 运行测试并显示覆盖率
go test ./util/id -cover

# 运行性能基准测试
go test ./util/id -bench=. -benchmem
```

## 参考资源

- [CUID2 官方规范](https://github.com/paralleldrive/cuid2)
