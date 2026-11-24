# Hash 哈希工具包

提供多种哈希算法（MD5、SHA1、SHA256、SHA512）、HMAC 签名计算和 Bcrypt 密码哈希功能的 Go 语言工具包。

## 特性

- **多种哈希算法** - 支持 MD5、SHA1、SHA256、SHA512 四种常用哈希算法
- **HMAC 签名计算** - 支持基于各种哈希算法的 HMAC 签名计算
- **Bcrypt 密码哈希** - 提供安全的密码哈希和验证功能，适用于密码存储场景
- **泛型编程支持** - 使用 Go 泛型特性，提供类型安全且灵活的 API
- **多种输出格式** - 支持原始字节数组、十六进制字符串等多种输出格式
- **流式处理** - 支持从 io.Reader 直接计算哈希值，适合处理大文件
- **便捷方法** - 提供默认实例 `hash.Hash`，开箱即用

## 快速开始

### 基本哈希计算

```go
import "github.com/lite-lake/litecore-go/util/hash"

// 计算字符串哈希值
md5Hash := hash.Hash.MD5String("hello world")
sha256Hash := hash.Hash.SHA256String("hello world")

// 返回原始字节数组
hashBytes := hash.Hash.SHA256("hello world")
```

### HMAC 签名

```go
// HMAC-SHA256 签名
signature := hash.Hash.HMACSHA256String("data", "secret-key")

// 验证签名
expectedSignature := hash.Hash.HMACSHA256String("data", "secret-key")
isValid := signature == expectedSignature
```

### 密码哈希（Bcrypt）

```go
// 生成密码哈希值
hashedPassword, err := hash.Hash.BcryptHash("mypassword")
if err != nil {
    log.Fatal(err)
}

// 验证密码
isValid := hash.Hash.BcryptVerify("mypassword", hashedPassword)
```

### 大文件处理

```go
file, err := os.Open("large-file.dat")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// 计算文件哈希值
hashString, err := hash.HashReaderStringGeneric(file, hash.SHA256Algorithm{})
if err != nil {
    log.Fatal(err)
}
```

## 支持的哈希算法

### MD5

**输出长度**：32 字符（十六进制）

**使用场景**：
- 非安全场景的校验和（如文件完整性校验、缓存键生成）
- 已不适用于安全相关场景（已被证明存在安全漏洞）

```go
// 完整长度（32字符）
hash.Hash.MD5String("data")

// 16位短格式
hash.Hash.MD5String16("data")

// 32位格式
hash.Hash.MD5String32("data")

// 返回字节数组
hash.Hash.MD5("data")
```

### SHA1

**输出长度**：40 字符（十六进制）

**使用场景**：
- 遗留系统兼容
- 不推荐在新项目中使用（存在安全漏洞）

```go
// 完整长度（40字符）
hash.Hash.SHA1String("data")

// 返回字节数组
hash.Hash.SHA1("data")
```

### SHA256

**输出长度**：64 字符（十六进制）

**使用场景**：
- 推荐用于大多数安全场景
- 数字签名、API 签名验证
- 数据完整性校验
- 密码派生（需配合盐值）

```go
// 完整长度（64字符）
hash.Hash.SHA256String("data")

// 返回字节数组
hash.Hash.SHA256("data")
```

### SHA512

**输出长度**：128 字符（十六进制）

**使用场景**：
- 高安全要求场景
- 需要更长哈希值的应用
- 大数据量处理

```go
// 完整长度（128字符）
hash.Hash.SHA512String("data")

// 返回字节数组
hash.Hash.SHA512("data")
```

### Bcrypt

**输出长度**：60 字符（包含算法版本、成本因子、盐值和哈希值）

**使用场景**：
- 密码存储（专为密码设计）
- 每次生成不同的哈希值（包含随机盐值）
- 支持可调节的计算成本因子

```go
// 使用默认成本因子生成密码哈希
hashedPassword, err := hash.Hash.BcryptHash("mypassword")

// 指定成本因子（4-31，默认为 10）
hashedPassword, err := hash.Hash.BcryptHashWithCost("mypassword", 12)

// 验证密码
isValid := hash.Hash.BcryptVerify("mypassword", hashedPassword)
```

**注意事项**：
- Bcrypt 是专门为密码哈希设计的算法，每次生成的哈希值不同（包含随机盐值）
- 成本因子越高，计算时间越长，安全性越高
- 默认成本因子为 `hash.BcryptDefaultCost`（值为 10）

## HMAC 签名

HMAC（基于哈希的消息认证码）用于验证数据完整性和真实性。

```go
// HMAC-MD5
hash.Hash.HMACMD5String("data", "key")

// HMAC-SHA1
hash.Hash.HMACSHA1String("data", "key")

// HMAC-SHA256（推荐用于 API 签名）
hash.Hash.HMACSHA256String("data", "key")

// HMAC-SHA512
hash.Hash.HMACSHA512String("data", "key")

// 返回字节数组
hash.Hash.HMACSHA256("data", "key")
```

**使用场景**：
- API 签名验证
- 消息认证
- 令牌验证
- 防篡改校验

## 输出格式

```go
// FormatBytes: 原始字节数组
bytesHash := hash.HashHexGeneric("data", hash.MD5Algorithm{}, hash.FormatBytes)

// FormatHexShort: 16位十六进制字符串
shortHash := hash.HashHexGeneric("data", hash.MD5Algorithm{}, hash.FormatHexShort)

// FormatHexMedium: 32位十六进制字符串
mediumHash := hash.HashHexGeneric("data", hash.MD5Algorithm{}, hash.FormatHexMedium)

// FormatHexFull: 完整长度十六进制字符串（默认）
fullHash := hash.HashHexGeneric("data", hash.MD5Algorithm{}, hash.FormatHexFull)
```

## 泛型函数

### 哈希计算

```go
// 计算任意哈希算法的值
hashBytes := hash.HashGeneric("data", hash.SHA256Algorithm{})

// 指定输出格式
hashString := hash.HashHexGeneric("data", hash.SHA256Algorithm{}, hash.FormatHexFull)

// 处理字节数组
hashBytes = hash.HashBytesGeneric([]byte("data"), hash.SHA256Algorithm{})

// 处理 io.Reader（大文件）
file, _ := os.Open("file.txt")
hashBytes, err := hash.HashReaderGeneric(file, hash.SHA256Algorithm{})
```

### HMAC 计算

```go
// HMAC 泛型函数
hmacBytes := hash.HMACGeneric("data", "key", hash.SHA256Algorithm{})
hmacString := hash.HMACStringGeneric("data", "key", hash.SHA256Algorithm{})

// 处理字节数组
hmacBytes = hash.HMACBytesGeneric([]byte("data"), []byte("key"), hash.SHA256Algorithm{})

// 处理 io.Reader
file, _ := os.Open("file.txt")
hmacBytes, err := hash.HMACReaderGeneric(file, []byte("key"), hash.SHA256Algorithm{})
```

### 自定义算法

```go
// 实现自定义哈希算法
type CustomAlgorithm struct{}

func (CustomAlgorithm) Hash() hash.Hash {
    return sha3.New256()
}

// 使用自定义算法
result := hash.HashGeneric("data", CustomAlgorithm{})
```

## 使用场景

### 密码存储和验证

```go
// 用户注册时生成密码哈希
hashedPassword, err := hash.Hash.BcryptHash("user-password")
if err != nil {
    log.Fatal(err)
}

// 存储到数据库
db.SaveUser(username, hashedPassword)

// 用户登录时验证密码
storedHash := db.GetUserPassword(username)
isValid := hash.Hash.BcryptVerify("user-input-password", storedHash)
if isValid {
    // 登录成功
}
```

### API 签名验证

```go
// 生成 API 签名
func GenerateAPISignature(data string, secretKey string) string {
    return hash.Hash.HMACSHA256String(data, secretKey)
}

// 验证 API 签名
func VerifyAPISignature(data string, signature string, secretKey string) bool {
    expectedSignature := GenerateAPISignature(data, secretKey)
    return expectedSignature == signature
}
```

### 文件完整性校验

```go
// 计算文件哈希值
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
```

### 缓存键生成

```go
func GenerateCacheKey(prefix string, params ...string) string {
    key := prefix
    for _, param := range params {
        key += ":" + param
    }
    // 使用 MD5 短格式作为缓存键
    return hash.Hash.MD5String16(key)
}

cacheKey := GenerateCacheKey("user", "123", "profile")
```

### 数据去重

```go
func GetDataUniqueID(data string) string {
    return hash.Hash.SHA256String(data)
}

// 使用 map 记录已处理的数据哈希
seenHashes := make(map[string]bool)

func ProcessData(data string) {
    dataHash := GetDataUniqueID(data)
    if seenHashes[dataHash] {
        return // 数据已处理
    }
    // 处理数据
    seenHashes[dataHash] = true
}
```

## 性能考虑

### 算法选择

| 算法 | 速度 | 安全性 | 推荐场景 |
|------|------|--------|----------|
| MD5 | 最快 | 已不安全 | 仅适用于非安全场景的校验和 |
| SHA1 | 较快 | 已不推荐 | 避免在新项目中使用 |
| SHA256 | 中等 | 安全 | 推荐用于大多数场景 |
| SHA512 | 较慢 | 最安全 | 高安全要求场景 |
| Bcrypt | 慢 | 密码安全 | 密码存储（专为密码设计） |

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

1. **安全性**：MD5 和 SHA1 已被证明存在安全漏洞，不应在安全敏感场景中使用
2. **密码存储**：对于密码存储，必须使用 Bcrypt 等专门的密码哈希算法
3. **错误处理**：使用 io.Reader 相关函数时，务必处理可能的错误
4. **字符编码**：确保输入数据的字符编码一致，避免因编码不同导致哈希结果不同
5. **密钥管理**：使用 HMAC 时，妥善保管密钥，避免硬编码在代码中
6. **Bcrypt 成本因子**：根据实际需求选择合适的成本因子，默认为 10

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

# 运行特定测试
go test ./util/hash -run TestBcryptHash
```
