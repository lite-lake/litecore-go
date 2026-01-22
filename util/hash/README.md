# Hash 哈希工具包

提供多种哈希算法（MD5、SHA1、SHA256、SHA512）和 HMAC 计算功能的 Go 语言工具包，支持泛型编程。

## 特性

- **多种哈希算法支持** - 提供 MD5、SHA1、SHA256、SHA512 四种常用哈希算法
- **HMAC 签名计算** - 支持基于各种哈希算法的 HMAC 签名计算
- **泛型编程支持** - 使用 Go 泛型特性，提供类型安全且灵活的 API
- **多种输出格式** - 支持原始字节数组、十六进制字符串等多种输出格式
- **流式处理** - 支持从 io.Reader 直接计算哈希值，适合处理大文件
- **便捷方法** - 提供默认实例 `util.Hash`，开箱即用

## 快速开始

### 安装

```bash
go get github.com/lite-lake/litecore-go
```

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/hash"
)

func main() {

    // 返回原始字节数组
    hashBytes := hash.Hash.SHA512(data)
    fmt.Printf("字节数组长度: %d\n", len(hashBytes))

    // 返回完整 128 位十六进制字符串
    hashString := hash.Hash.SHA512String(data)
    fmt.Printf("SHA512: %s\n", hashString)
}
```

### HMAC 签名

#### HMAC-MD5

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/hash"
)

func main() {
    // 打开文件
    file, err := os.Open("large-file.dat")
    if err != nil {
        fmt.Printf("打开文件失败: %v\n", err)
        return
    }
    defer file.Close()

    // 计算文件的 SHA256 哈希值
    hashString, err := hash.HashReaderStringGeneric(file, hash.SHA256Algorithm{})
    if err != nil {
        fmt.Printf("计算哈希失败: %v\n", err)
        return
    }

    fmt.Printf("文件 SHA256: %s\n", hashString)
}
```

### 哈希输出格式

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/hash"
)

func main() {
    data := "hello world"

    // FormatBytes: 原始字节数组（转换为字符串）
    bytesHash := hash.HashHexGeneric(data, hash.MD5Algorithm{}, hash.FormatBytes)
    fmt.Printf("Bytes: %s\n", bytesHash)

    // FormatHexShort: 16位十六进制字符串
    shortHash := hash.HashHexGeneric(data, hash.MD5Algorithm{}, hash.FormatHexShort)
    fmt.Printf("Short: %s\n", shortHash)
    // 输出: 5eb63bbb

    // FormatHexMedium: 32位十六进制字符串
    mediumHash := hash.HashHexGeneric(data, hash.MD5Algorithm{}, hash.FormatHexMedium)
    fmt.Printf("Medium: %s\n", mediumHash)
    // 输出: 5eb63bbbe01eeed093cb22bb8f5acdc3

    // FormatHexFull: 完整长度十六进制字符串
    fullHash := hash.HashHexGeneric(data, hash.MD5Algorithm{}, hash.FormatHexFull)
    fmt.Printf("Full: %s\n", fullHash)
    // 输出: 5eb63bbbe01eeed093cb22bb8f5acdc3
}
```

## API 参考

### 默认实例

```go
// 默认的哈希操作实例
var Hash = &hashEngine{}
```

### 哈希算法接口

```go
// 哈希算法接口
type HashAlgorithm interface {
    Hash() hash.Hash
}

// MD5 算法实现
type MD5Algorithm struct{}

// SHA1 算法实现
type SHA1Algorithm struct{}

// SHA256 算法实现
type SHA256Algorithm struct{}

// SHA512 算法实现
type SHA512Algorithm struct{}
```

### 输出格式常量

```go
// 哈希输出格式枚举
type HashOutputFormat int

const (
    FormatBytes    HashOutputFormat = iota  // 原始字节数组格式
    FormatHexShort                          // 16位十六进制字符串
    FormatHexMedium                         // 32位十六进制字符串
    FormatHexFull                           // 完整长度十六进制字符串
)
```

### MD5 便捷方法

```go
// 计算并返回 MD5 字节数组
func (h *hashEngine) MD5(data string) []byte

// 计算并返回 MD5 完整十六进制字符串
func (h *hashEngine) MD5String(data string) string

// 计算并返回 MD5 16位十六进制字符串
func (h *hashEngine) MD5String16(data string) string

// 计算并返回 MD5 32位十六进制字符串
func (h *hashEngine) MD5String32(data string) string
```

### SHA1 便捷方法

```go
// 计算并返回 SHA1 字节数组
func (h *hashEngine) SHA1(data string) []byte

// 计算并返回 SHA1 完整十六进制字符串
func (h *hashEngine) SHA1String(data string) string
```

### SHA256 便捷方法

```go
// 计算并返回 SHA256 字节数组
func (h *hashEngine) SHA256(data string) []byte

// 计算并返回 SHA256 完整十六进制字符串
func (h *hashEngine) SHA256String(data string) string
```

### SHA512 便捷方法

```go
// 计算并返回 SHA512 字节数组
func (h *hashEngine) SHA512(data string) []byte

// 计算并返回 SHA512 完整十六进制字符串
func (h *hashEngine) SHA512String(data string) string
```

### HMAC 便捷方法

```go
// 计算并返回 HMAC-MD5 字节数组
func (h *hashEngine) HMACMD5(data string, key string) []byte

// 计算并返回 HMAC-MD5 完整十六进制字符串
func (h *hashEngine) HMACMD5String(data string, key string) string

// 计算并返回 HMAC-SHA1 字节数组
func (h *hashEngine) HMACSHA1(data string, key string) []byte

// 计算并返回 HMAC-SHA1 完整十六进制字符串
func (h *hashEngine) HMACSHA1String(data string, key string) string

// 计算并返回 HMAC-SHA256 字节数组
func (h *hashEngine) HMACSHA256(data string, key string) []byte

// 计算并返回 HMAC-SHA256 完整十六进制字符串
func (h *hashEngine) HMACSHA256String(data string, key string) string

// 计算并返回 HMAC-SHA512 字节数组
func (h *hashEngine) HMACSHA512(data string, key string) []byte

// 计算并返回 HMAC-SHA512 完整十六进制字符串
func (h *hashEngine) HMACSHA512String(data string, key string) string
```

### 泛型哈希函数

```go
// 计算字符串的哈希值，返回字节数组
func HashGeneric[T HashAlgorithm](data string, algorithm T) []byte

// 计算字节数组的哈希值
func HashBytesGeneric[T HashAlgorithm](data []byte, algorithm T) []byte

// 从 io.Reader 计算哈希值
func HashReaderGeneric[T HashAlgorithm](r io.Reader, algorithm T) ([]byte, error)

// 计算哈希值并返回指定格式的十六进制字符串
func HashHexGeneric[T HashAlgorithm](data string, algorithm T, format HashOutputFormat) string

// 计算字节数组的哈希值并返回指定格式的十六进制字符串
func HashBytesHexGeneric[T HashAlgorithm](data []byte, algorithm T, format HashOutputFormat) string

// 从 io.Reader 计算哈希值并返回指定格式的十六进制字符串
func HashReaderHexGeneric[T HashAlgorithm](r io.Reader, algorithm T, format HashOutputFormat) (string, error)

// 计算哈希并返回完整十六进制字符串
func HashStringGeneric[T HashAlgorithm](data string, algorithm T) string

// 从 io.Reader 计算哈希并返回完整十六进制字符串
func HashReaderStringGeneric[T HashAlgorithm](r io.Reader, algorithm T) (string, error)
```

### 泛型 HMAC 函数

```go
// 计算 HMAC 哈希值，返回字节数组
func HMACGeneric[T HashAlgorithm](data string, key string, algorithm T) []byte

// 计算字节数组的 HMAC 哈希值
func HMACBytesGeneric[T HashAlgorithm](data []byte, key []byte, algorithm T) []byte

// 从 io.Reader 计算 HMAC 哈希值
func HMACReaderGeneric[T HashAlgorithm](r io.Reader, key []byte, algorithm T) ([]byte, error)

// 计算 HMAC 哈希值并返回指定格式的十六进制字符串
func HMACHexGeneric[T HashAlgorithm](data string, key string, algorithm T, format HashOutputFormat) string

// 计算字节数组的 HMAC 哈希值并返回指定格式的十六进制字符串
func HMACBytesHexGeneric[T HashAlgorithm](data []byte, key []byte, algorithm T, format HashOutputFormat) string

// 从 io.Reader 计算 HMAC 哈希值并返回指定格式的十六进制字符串
func HMACReaderHexGeneric[T HashAlgorithm](r io.Reader, key []byte, algorithm T, format HashOutputFormat) (string, error)

// 计算 HMAC 并返回完整十六进制字符串
func HMACStringGeneric[T HashAlgorithm](data string, key string, algorithm T) string

// 从 io.Reader 计算 HMAC 并返回完整十六进制字符串
func HMACReaderStringGeneric[T HashAlgorithm](r io.Reader, key []byte, algorithm T) (string, error)
```

## 常见使用场景

### 密码哈希存储

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/hash"
)

func HashPassword(password string) string {
    // 使用 SHA256 哈希密码（生产环境建议使用 bcrypt 等专门的密码哈希算法）
    return hash.Hash.SHA256String(password)
}

func main() {
    password := "my-secure-password"
    hashedPassword := HashPassword(password)
    fmt.Printf("密码哈希: %s\n", hashedPassword)
}
```

### API 签名验证

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/hash"
)

// 生成 API 签名
func GenerateAPISignature(data string, secretKey string) string {
    return hash.Hash.HMACSHA256String(data, secretKey)
}

// 验证 API 签名
func VerifyAPISignature(data string, signature string, secretKey string) bool {
    expectedSignature := GenerateAPISignature(data, secretKey)
    return expectedSignature == signature
}

func main() {
    data := "request-data"
    secretKey := "api-secret-key"

    // 生成签名
    signature := GenerateAPISignature(data, secretKey)
    fmt.Printf("签名: %s\n", signature)

    // 验证签名
    isValid := VerifyAPISignature(data, signature, secretKey)
    fmt.Printf("签名验证: %v\n", isValid)
}
```

### 文件完整性校验

```go
package main

import (
    "fmt"
    "os"
    "github.com/lite-lake/litecore-go/util/hash"
)

// 计算文件的哈希值
func CalculateFileHash(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    return hash.HashReaderStringGeneric(file, hash.SHA256Algorithm{})
}

// 验证文件完整性
func VerifyFileIntegrity(filePath string, expectedHash string) (bool, error) {
    actualHash, err := CalculateFileHash(filePath)
    if err != nil {
        return false, err
    }
    return actualHash == expectedHash, nil
}

func main() {
    filePath := "example.txt"
    expectedHash := "预期的哈希值"

    isValid, err := VerifyFileIntegrity(filePath, expectedHash)
    if err != nil {
        fmt.Printf("验证失败: %v\n", err)
        return
    }

    if isValid {
        fmt.Println("文件完整性验证通过")
    } else {
        fmt.Println("文件已被篡改")
    }
}
```

### 数据去重

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecorego/util/hash"
)

// 计算数据的唯一标识
func GetDataUniqueID(data string) string {
    return hash.Hash.SHA256String(data)
}

// 检查数据是否重复
func IsDuplicate(data string, seenHashes map[string]bool) bool {
    dataHash := GetDataUniqueID(data)
    return seenHashes[dataHash]
}

// 记录已处理的数据
func RecordData(data string, seenHashes map[string]bool) {
    dataHash := GetDataUniqueID(data)
    seenHashes[dataHash] = true
}

func main() {
    seenHashes := make(map[string]bool)

    data1 := "hello world"
    data2 := "hello world"
    data3 := "different data"

    // 检查并记录数据
    fmt.Printf("Data1 重复: %v\n", IsDuplicate(data1, seenHashes))
    RecordData(data1, seenHashes)

    fmt.Printf("Data2 重复: %v\n", IsDuplicate(data2, seenHashes))

    fmt.Printf("Data3 重复: %v\n", IsDuplicate(data3, seenHashes))
}
```

### 缓存键生成

```go
package main

import (
    "fmt"
    "github.com/lite-lake/litecore-go/util/hash"
)

// 生成缓存键
func GenerateCacheKey(prefix string, params ...string) string {
    key := prefix
    for _, param := range params {
        key += ":" + param
    }
    // 使用 MD5 短格式作为缓存键
    return hash.Hash.MD5String16(key)
}

func main() {
    // 为不同的查询参数生成缓存键
    cacheKey1 := GenerateCacheKey("user", "123", "profile")
    cacheKey2 := GenerateCacheKey("user", "456", "profile")
    cacheKey3 := GenerateCacheKey("user", "123", "profile")

    fmt.Printf("缓存键 1: %s\n", cacheKey1)
    fmt.Printf("缓存键 2: %s\n", cacheKey2)
    fmt.Printf("缓存键 3: %s\n", cacheKey3)

    fmt.Printf("键1和键3相同: %v\n", cacheKey1 == cacheKey3)
}
```

## 性能考虑

### 算法选择

- **MD5**: 最快，但已不安全，仅适用于非安全场景的校验和
- **SHA1**: 速度较快，但已不推荐用于安全场景
- **SHA256**: 速度和安全性的良好平衡，推荐用于大多数场景
- **SHA512**: 最安全但速度较慢，适用于高安全要求场景

### 大文件处理

对于大文件，建议使用 `io.Reader` 接口的函数，避免将整个文件加载到内存：

```go
file, _ := os.Open("large-file.dat")
defer file.Close()

hashString, err := hash.HashReaderStringGeneric(file, hash.SHA256Algorithm{})
```

### 批量计算

如果需要计算多个哈希值，可以考虑并行处理以提高性能。

## 注意事项

1. **安全性**: MD5 和 SHA1 已被证明存在安全漏洞，不应在安全敏感场景中使用
2. **密码存储**: 对于密码存储，建议使用专门的密码哈希算法如 bcrypt、Argon2 或 scrypt
3. **错误处理**: 使用 io.Reader 相关函数时，务必处理可能的错误
4. **字符编码**: 确保输入数据的字符编码一致，避免因编码不同导致哈希结果不同
5. **密钥管理**: 使用 HMAC 时，妥善保管密钥，避免硬编码在代码中

## 运行测试

```bash
# 运行所有测试
go test ./util/hash

# 运行测试并显示覆盖率
go test -cover ./util/hash

# 运行性能基准测试
go test -bench=. ./util/hash

# 查看详细的测试输出
go test -v ./util/hash
```

## 许可证

本工具包是 litecore-go 项目的一部分，遵循项目的开源许可证。
