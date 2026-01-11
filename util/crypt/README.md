# Crypt - 加密解密工具包

提供全面的加密解密功能，包括对称加密、非对称加密、哈希、签名、Base64 编码等。

## 特性

- **对称加密**：支持 AES-128/192/256，使用 GCM 模式保证机密性和完整性
- **非对称加密**：支持 RSA 加密解密，提供多种密钥长度选择
- **密码哈希**：提供 Bcrypt 和 PBKDF2 两种安全的密码哈希算法
- **消息签名**：支持 HMAC 和 ECDSA 数字签名，确保数据完整性和身份认证
- **编码转换**：提供 Base64 和 Hex 编码解码功能
- **安全工具**：包含随机数生成、常数时间比较等安全辅助函数

## 快速开始

```go
package main

import (
    "fmt"
    "log"

    "yourproject/util/crypt"
)

func main() {
    // Base64 编码解码
    encoded := crypt.Crypt.Base64Encode("Hello, World!")
    fmt.Println("Base64 编码:", encoded)

    decoded, err := crypt.Crypt.Base64Decode(encoded)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Base64 解码:", decoded)

    // AES 对称加密
    key, _ := crypt.Crypt.GenerateAESKey(crypt.AES256)
    encrypted, _ := crypt.Crypt.AESEncryptToBase64("敏感信息", key)
    fmt.Println("AES 加密:", encrypted)

    decrypted, _ := crypt.Crypt.AESDecryptFromBase64(encrypted, key)
    fmt.Println("AES 解密:", decrypted)

    // 密码哈希
    hash, _ := crypt.Crypt.BcryptHash("mypassword123", 10)
    fmt.Println("密码哈希:", hash)

    verified := crypt.Crypt.BcryptVerify("mypassword123", hash)
    fmt.Println("密码验证:", verified)
}
```

## 功能详解

### Base64 编码解码

提供标准的 Base64 编码解码功能，以及 URL 安全的 Base64 变体。

#### 基本使用

```go
// 字符串编码
encoded := crypt.Crypt.Base64Encode("Hello, World!")
// 输出: SGVsbG8sIFdvcmxkIQ==

// 字符串解码
decoded, err := crypt.Crypt.Base64Decode("SGVsbG8sIFdvcmxkIQ==")
if err != nil {
    log.Fatal(err)
}
// 输出: Hello, World!

// 字节数组编码
data := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}
encoded := crypt.Crypt.Base64EncodeBytes(data)
// 输出: SGVsbG8=

// 字节数组解码
decoded, err := crypt.Crypt.Base64DecodeBytes("SGVsbG8=")
if err != nil {
    log.Fatal(err)
}
```

#### URL 安全编码

```go
// URL 安全的 Base64 编码
encoded := crypt.Crypt.Base64URLEncode("Hello, World!")
// 输出: SGVsbG8sIFdvcmxkIQ==

// URL 安全的 Base64 解码
decoded, err := crypt.Crypt.Base64URLDecode(encoded)
if err != nil {
    log.Fatal(err)
}
```

#### 验证函数

```go
// 检查是否为有效的 Base64
isValid := crypt.Crypt.IsBase64("SGVsbG8sIFdvcmxkIQ==")
// 输出: true

isValid = crypt.Crypt.IsBase64("Invalid@Base64!")
// 输出: false
```

### Hex 编码解码

提供十六进制编码解码功能，常用于显示二进制数据。

```go
// 十六进制编码
encoded := crypt.Crypt.HexEncode("Hello")
// 输出: 48656c6c6f

// 十六进制解码
decoded, err := crypt.Crypt.HexDecode("48656c6c6f")
if err != nil {
    log.Fatal(err)
}
// 输出: Hello

// 字节数组编码
data := []byte{0x00, 0xFF, 0xAA, 0x55}
encoded := crypt.Crypt.HexEncodeBytes(data)
// 输出: 00ffaa55

// 字节数组解码
decoded, err := crypt.Crypt.HexDecodeBytes("00ffaa55")
if err != nil {
    log.Fatal(err)
}

// 检查是否为有效的十六进制
isValid := crypt.Crypt.IsHex("48656c6c6f")
// 输出: true
```

### AES 对称加密

支持 AES-128、AES-192、AES-256 三种密钥长度，使用 GCM 模式提供认证加密。

#### 生成密钥

```go
// 生成 AES-128 密钥（16 字节）
key128, err := crypt.Crypt.GenerateAESKey(crypt.AES128)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("密钥长度: %d 字节\n", len(key128))
// 输出: 密钥长度: 16 字节

// 生成 AES-256 密钥（32 字节）
key256, err := crypt.Crypt.GenerateAESKey(crypt.AES256)
if err != nil {
    log.Fatal(err)
}

// 生成十六进制格式的密钥（便于存储和传输）
keyHex, err := crypt.Crypt.GenerateAESKeyHex(crypt.AES256)
if err != nil {
    log.Fatal(err)
}
// 输出示例: 1a2b3c4d5e6f78901a2b3c4d5e6f78901a2b3c4d5e6f78901a2b3c4d5e6f7890
```

#### 加密解密

```go
// 生成密钥
key, err := crypt.Crypt.GenerateAESKey(crypt.AES256)
if err != nil {
    log.Fatal(err)
}

// 方法 1: 直接使用字节数组
plaintext := []byte("这是需要加密的敏感信息")
ciphertext, err := crypt.Crypt.AESEncrypt(plaintext, key)
if err != nil {
    log.Fatal(err)
}

// 解密
decrypted, err := crypt.Crypt.AESDecrypt(ciphertext, key)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("解密结果: %s\n", string(decrypted))
// 输出: 这是需要加密的敏感信息

// 方法 2: 使用 Base64 编码（推荐，便于存储和传输）
plaintext := "这是需要加密的敏感信息"
encrypted, err := crypt.Crypt.AESEncryptToBase64(plaintext, key)
if err != nil {
    log.Fatal(err)
}

// 解密
decrypted, err := crypt.Crypt.AESDecryptFromBase64(encrypted, key)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("解密结果: %s\n", decrypted)
```

#### 密钥管理

```go
// 编码密钥为可传输格式
key := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
encodedKey := crypt.Crypt.EncodeKey(key)
// 输出: AQIDBAU=

// 解码密钥
decodedKey, err := crypt.Crypt.DecodeKey(encodedKey)
if err != nil {
    log.Fatal(err)
}
```

### RSA 非对称加密

支持 RSA 非对称加密，提供多种密钥长度选择。

#### 生成密钥对

```go
// 生成 2048 位 RSA 密钥对
privateKey, publicKey, err := crypt.Crypt.GenerateRSAKeys(crypt.RSA2048)
if err != nil {
    log.Fatal(err)
}

// 其他密钥长度选项:
// crypt.RSA1024 - 1024 位（不推荐用于生产环境）
// crypt.RSA2048 - 2048 位（推荐）
// crypt.RSA3072 - 3072 位（更高安全性）
// crypt.RSA4096 - 4096 位（最高安全性）
```

#### 加密解密

```go
// 方法 1: 直接使用字节数组
plaintext := []byte("RSA 加密的秘密消息")
ciphertext, err := crypt.Crypt.RSAEncrypt(plaintext, publicKey)
if err != nil {
    log.Fatal(err)
}

// 解密
decrypted, err := crypt.Crypt.RSADecrypt(ciphertext, privateKey)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("解密结果: %s\n", string(decrypted))
// 输出: RSA 加密的秘密消息

// 方法 2: 使用 Base64 编码（推荐）
plaintext := "RSA 加密的秘密消息"
encrypted, err := crypt.Crypt.RSAEncryptToBase64(plaintext, publicKey)
if err != nil {
    log.Fatal(err)
}

// 解密
decrypted, err := crypt.Crypt.RSADecryptFromBase64(encrypted, privateKey)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("解密结果: %s\n", decrypted)
```

#### 密钥格式转换

```go
// 转换私钥为 PEM 格式字符串
privateKeyPEM := crypt.Crypt.PrivateKeyToPEM(privateKey)
fmt.Println("私钥 PEM:", privateKeyPEM)

// 转换公钥为 PEM 格式字符串
publicKeyPEM := crypt.Crypt.PublicKeyToPEM(publicKey)
fmt.Println("公钥 PEM:", publicKeyPEM)
```

### 密码哈希

提供两种安全的密码哈希算法：Bcrypt 和 PBKDF2。

#### Bcrypt 哈希

Bcrypt 是专门为密码存储设计的哈希算法，自带盐值且计算速度可调。

```go
// 生成哈希
password := "mypassword123"
cost := 10 // 计算成本因子，范围 4-31，推荐 10-12
hash, err := crypt.Crypt.BcryptHash(password, cost)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Bcrypt 哈希:", hash)
// 输出示例: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy

// 验证密码
isCorrect := crypt.Crypt.BcryptVerify("mypassword123", hash)
fmt.Println("密码正确:", isCorrect)
// 输出: 密码正确: true

isWrong := crypt.Crypt.BcryptVerify("wrongpassword", hash)
fmt.Println("错误密码:", isWrong)
// 输出: 错误密码: false
```

**Bcrypt 成本因子说明：**
- `cost = 4`：最快，安全性最低（仅用于测试）
- `cost = 10`：推荐用于大多数应用
- `cost = 12`：高安全性要求
- `cost = 14+`：极高安全性，但计算时间显著增加

#### PBKDF2 哈希

PBKDF2 是另一种基于密钥派生的密码哈希算法。

```go
// 生成随机盐值
salt, err := crypt.Crypt.GenerateSalt(16)
if err != nil {
    log.Fatal(err)
}

// 或者生成十六进制格式的盐值
saltHex, err := crypt.Crypt.GenerateSaltHex(16)
if err != nil {
    log.Fatal(err)
}

// 生成哈希
password := "mypassword123"
iterations := 10000 // 迭代次数，推荐至少 10000 次
keyLen := 32       // 输出密钥长度（字节）
hash := crypt.Crypt.PBKDF2Hash(password, salt, iterations, keyLen)
fmt.Println("PBKDF2 哈希:", hash)

// 验证密码
isCorrect := crypt.Crypt.PBKDF2Verify(password, salt, hash, iterations, keyLen)
fmt.Println("密码正确:", isCorrect)
// 输出: 密码正确: true
```

**PBKDF2 参数建议：**
- `salt`：至少 16 字节，每个密码使用不同的随机盐值
- `iterations`：最少 10000 次，2024 年推荐 100000+ 次
- `keyLen`：通常使用 32 字节（256 位）

### HMAC 签名

HMAC（基于哈希的消息认证码）用于验证消息完整性和真实性。

#### 基本使用

```go
data := []byte("需要签名的数据")
key := []byte("密钥")

// 方法 1: 生成原始签名
signature := crypt.Crypt.HMACSign(data, key, crypto.SHA256.New)
fmt.Printf("签名长度: %d 字节\n", len(signature))
// 输出: 签名长度: 32 字节

// 方法 2: 生成十六进制格式的签名
signatureHex := crypt.Crypt.HMACSignHex(data, key, crypto.SHA256.New)
fmt.Println("十六进制签名:", signatureHex)

// 方法 3: 生成 Base64 格式的签名
signatureBase64 := crypt.Crypt.HMACSignBase64(data, key, crypto.SHA256.New)
fmt.Println("Base64 签名:", signatureBase64)

// 验证签名
isValid := crypt.Crypt.HMACVerify(data, key, signature, crypto.SHA256.New)
fmt.Println("签名有效:", isValid)
// 输出: 签名有效: true
```

#### 便捷方法

```go
data := []byte("需要签名的数据")
key := []byte("密钥")

// 使用 SHA256 的便捷方法
signature := crypt.Crypt.HMACSignWithSHA256(data, key)
signatureHex := crypt.Crypt.HMACSignHexWithSHA256(data, key)

// 使用 SHA512 的便捷方法
signature = crypt.Crypt.HMACSignWithSHA512(data, key)
signatureHex = crypt.Crypt.HMACSignHexWithSHA512(data, key)
```

#### 支持的哈希算法

```go
import (
    "crypto/md5"
    "crypto/sha1"
    "crypto/sha256"
    "crypto/sha512"
)

// HMAC 支持任何实现了 hash.Hash 的算法
// MD5（不推荐，仅用于兼容性）
signature := crypt.Crypt.HMACSign(data, key, md5.New)

// SHA-1（不推荐，仅用于兼容性）
signature = crypt.Crypt.HMACSign(data, key, sha1.New)

// SHA-256（推荐）
signature = crypt.Crypt.HMACSign(data, key, sha256.New)

// SHA-512（最高安全性）
signature = crypt.Crypt.HMACSign(data, key, sha512.New)
```

### ECDSA 数字签名

ECDSA（椭圆曲线数字签名算法）提供高安全性的数字签名功能。

#### 生成密钥对

```go
// 生成 ECDSA 密钥对（使用 P-256 曲线）
privateKey, publicKey, err := crypt.Crypt.GenerateECDSAKeys()
if err != nil {
    log.Fatal(err)
}
fmt.Println("ECDSA 密钥对生成成功")
```

#### 签名和验证

```go
data := []byte("需要签名的数据")

// 方法 1: 生成原始签名
signature, err := crypt.Crypt.ECDSASign(data, privateKey)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("签名长度: %d 字节\n", len(signature))
// 输出: 签名长度: 约 70-72 字节

// 方法 2: 生成十六进制格式的签名
signatureHex, err := crypt.Crypt.ECDSASignHex(data, privateKey)
if err != nil {
    log.Fatal(err)
}
fmt.Println("十六进制签名:", signatureHex)

// 验证签名（字节数组）
isValid := crypt.Crypt.ECDSAVerify(data, signature, publicKey)
fmt.Println("签名有效:", isValid)
// 输出: 签名有效: true

// 验证签名（十六进制）
isValid, err = crypt.Crypt.ECDSAVerifyHex(data, signatureHex, publicKey)
if err != nil {
    log.Fatal(err)
}
fmt.Println("签名有效:", isValid)
```

### 安全工具函数

#### 随机数生成

```go
// 生成随机字节
randomBytes, err := crypt.Crypt.GenerateRandomBytes(16)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("随机字节: %x\n", randomBytes)

// 生成随机字符串（字母+数字）
randomString, err := crypt.Crypt.GenerateRandomString(32)
if err != nil {
    log.Fatal(err)
}
fmt.Println("随机字符串:", randomString)
// 输出示例: a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6
```

#### 安全比较

```go
// 常数时间比较（防止时序攻击）
a := []byte("sensitive-data")
b := []byte("sensitive-data")
isEqual := crypt.Crypt.ConstantTimeCompare(a, b)
fmt.Println("相等:", isEqual)
// 输出: 相等: true

// 安全字符串比较
isEqual = crypt.Crypt.SecureEqual("password123", "password123")
fmt.Println("相等:", isEqual)
// 输出: 相等: true
```

## API 参考

### 编码解码

#### Base64

| 函数 | 说明 |
|------|------|
| `Base64Encode(data string) string` | Base64 编码字符串 |
| `Base64EncodeBytes(data []byte) string` | Base64 编码字节数组 |
| `Base64Decode(data string) (string, error)` | Base64 解码为字符串 |
| `Base64DecodeBytes(data string) ([]byte, error)` | Base64 解码为字节数组 |
| `Base64URLEncode(data string) string` | URL 安全的 Base64 编码 |
| `Base64URLDecode(data string) (string, error)` | URL 安全的 Base64 解码 |
| `IsBase64(s string) bool` | 检查是否为有效 Base64 |

#### Hex

| 函数 | 说明 |
|------|------|
| `HexEncode(data string) string` | 十六进制编码字符串 |
| `HexEncodeBytes(data []byte) string` | 十六进制编码字节数组 |
| `HexDecode(data string) (string, error)` | 十六进制解码为字符串 |
| `HexDecodeBytes(data string) ([]byte, error)` | 十六进制解码为字节数组 |
| `IsHex(s string) bool` | 检查是否为有效十六进制 |

### AES 对称加密

| 函数 | 说明 |
|------|------|
| `GenerateAESKey(keySize AESKeySize) ([]byte, error)` | 生成 AES 密钥 |
| `GenerateAESKeyHex(keySize AESKeySize) (string, error)` | 生成十六进制格式 AES 密钥 |
| `AESEncrypt(plaintext, key []byte) ([]byte, error)` | AES 加密 |
| `AESEncryptToBase64(plaintext string, key []byte) (string, error)` | AES 加密并 Base64 编码 |
| `AESDecrypt(ciphertext, key []byte) ([]byte, error)` | AES 解密 |
| `AESDecryptFromBase64(ciphertext string, key []byte) (string, error)` | 从 Base64 字符串 AES 解密 |

**AES 密钥大小常量：**
- `AES128` - 128 位密钥（16 字节）
- `AES192` - 192 位密钥（24 字节）
- `AES256` - 256 位密钥（32 字节）

### RSA 非对称加密

| 函数 | 说明 |
|------|------|
| `GenerateRSAKeys(bits RSABits) (*rsa.PrivateKey, *rsa.PublicKey, error)` | 生成 RSA 密钥对 |
| `RSAEncrypt(plaintext []byte, publicKey *rsa.PublicKey) ([]byte, error)` | RSA 公钥加密 |
| `RSAEncryptToBase64(plaintext string, publicKey *rsa.PublicKey) (string, error)` | RSA 加密并 Base64 编码 |
| `RSADecrypt(ciphertext []byte, privateKey *rsa.PrivateKey) ([]byte, error)` | RSA 私钥解密 |
| `RSADecryptFromBase64(ciphertext string, privateKey *rsa.PrivateKey) (string, error)` | 从 Base64 字符串 RSA 解密 |

**RSA 密钥位数常量：**
- `RSA1024` - 1024 位密钥（不推荐用于生产环境）
- `RSA2048` - 2048 位密钥（推荐）
- `RSA3072` - 3072 位密钥（高安全性）
- `RSA4096` - 4096 位密钥（最高安全性）

### 密码哈希

#### Bcrypt

| 函数 | 说明 |
|------|------|
| `BcryptHash(password string, cost int) (string, error)` | Bcrypt 密码哈希 |
| `BcryptVerify(password, hash string) bool` | Bcrypt 密码验证 |

**Bcrypt 成本因子建议：**
- 测试环境：`cost = 4`
- 生产环境：`cost = 10-12`
- 高安全性：`cost = 14+`

#### PBKDF2

| 函数 | 说明 |
|------|------|
| `PBKDF2Hash(password, salt string, iterations, keyLen int) string` | PBKDF2 密码哈希 |
| `PBKDF2Verify(password, salt, hash string, iterations, keyLen int) bool` | PBKDF2 密码验证 |
| `GenerateSalt(length int) ([]byte, error)` | 生成随机盐值 |
| `GenerateSaltHex(length int) (string, error)` | 生成十六进制格式的随机盐值 |

**PBKDF2 参数建议：**
- `salt`：至少 16 字节
- `iterations`：至少 10000 次，推荐 100000+ 次
- `keyLen`：通常使用 32 字节

### HMAC 签名

| 函数 | 说明 |
|------|------|
| `HMACSign(data, key []byte, hashFunc func() hash.Hash) []byte` | HMAC 签名 |
| `HMACSignHex(data, key []byte, hashFunc func() hash.Hash) string` | HMAC 签名并转十六进制 |
| `HMACSignBase64(data, key []byte, hashFunc func() hash.Hash) string` | HMAC 签名并转 Base64 |
| `HMACVerify(data, key, signature []byte, hashFunc func() hash.Hash) bool` | HMAC 验证 |
| `HMACSignWithSHA256(data, key []byte) []byte` | 使用 SHA256 的 HMAC 签名 |
| `HMACSignHexWithSHA256(data, key []byte) string` | 使用 SHA256 的 HMAC 签名（十六进制） |
| `HMACSignWithSHA512(data, key []byte) []byte` | 使用 SHA512 的 HMAC 签名 |
| `HMACSignHexWithSHA512(data, key []byte) string` | 使用 SHA512 的 HMAC 签名（十六进制） |

### ECDSA 数字签名

| 函数 | 说明 |
|------|------|
| `GenerateECDSAKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error)` | 生成 ECDSA 密钥对 |
| `ECDSASign(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error)` | ECDSA 签名 |
| `ECDSASignHex(data []byte, privateKey *ecdsa.PrivateKey) (string, error)` | ECDSA 签名并转十六进制 |
| `ECDSAVerify(data, signature []byte, publicKey *ecdsa.PublicKey) bool` | ECDSA 验证 |
| `ECDSAVerifyHex(data []byte, signatureHex string, publicKey *ecdsa.PublicKey) (bool, error)` | ECDSA 验证十六进制签名 |

### 密钥格式转换

| 函数 | 说明 |
|------|------|
| `PrivateKeyToPEM(privateKey *rsa.PrivateKey) string` | RSA 私钥转 PEM 格式 |
| `PublicKeyToPEM(publicKey *rsa.PublicKey) string` | RSA 公钥转 PEM 格式 |
| `EncodeKey(key []byte) string` | 编码密钥为可传输格式 |
| `DecodeKey(encodedKey string) ([]byte, error)` | 解码密钥 |

### 工具函数

| 函数 | 说明 |
|------|------|
| `ConstantTimeCompare(a, b []byte) bool` | 常数时间比较（防止时序攻击） |
| `SecureEqual(a, b string) bool` | 安全字符串比较 |
| `GenerateRandomBytes(length int) ([]byte, error)` | 生成随机字节 |
| `GenerateRandomString(length int) (string, error)` | 生成随机字符串 |
| `IsBase64(s string) bool` | 检查是否为有效 Base64 |
| `IsHex(s string) bool` | 检查是否为有效十六进制 |

## 错误处理

所有可能失败的操作都返回错误，应该始终检查并处理错误：

```go
// 推荐的错误处理方式
encrypted, err := crypt.Crypt.AESEncryptToBase64(plaintext, key)
if err != nil {
    log.Printf("加密失败: %v", err)
    // 根据错误类型进行相应处理
    return
}

// 解密时的错误处理
decrypted, err := crypt.Crypt.AESDecryptFromBase64(encrypted, key)
if err != nil {
    log.Printf("解密失败: %v", err)
    // 可能是密钥错误或数据被篡改
    return
}
```

## 性能考虑

### Bcrypt 成本因子选择

Bcrypt 的成本因子直接影响计算时间：

```go
// 测试不同成本因子的性能
costs := []int{8, 10, 12, 14}
for _, cost := range costs {
    start := time.Now()
    _, err := crypt.Crypt.BcryptHash("password", cost)
    if err != nil {
        continue
    }
    duration := time.Since(start)
    fmt.Printf("Cost %d: %v\n", cost, duration)
}
```

**建议：**
- 在生产环境中，选择使哈希操作耗时 100-250ms 的成本因子
- 通常 `cost = 10` 或 `cost = 12` 是合适的起点

### PBKDF2 迭代次数选择

```go
// 测试不同迭代次数的性能
iterations := []int{10000, 50000, 100000, 200000}
for _, iter := range iterations {
    start := time.Now()
    hash := crypt.Crypt.PBKDF2Hash("password", "salt", iter, 32)
    _ = hash
    duration := time.Since(start)
    fmt.Printf("Iterations %d: %v\n", iter, duration)
}
```

**建议：**
- 2024 年推荐至少 100,000 次迭代
- 根据服务器性能调整，使哈希操作耗时 100-250ms

### AES 加密性能

AES-GCM 模式已经过优化，对于大多数应用来说性能足够：

```go
// 大数据分块加密
largeData := []byte(strings.Repeat("A", 1024*1024)) // 1MB
chunkSize := 4096 // 4KB 分块

for i := 0; i < len(largeData); i += chunkSize {
    end := i + chunkSize
    if end > len(largeData) {
        end = len(largeData)
    }
    chunk := largeData[i:end]
    encrypted, err := crypt.Crypt.AESEncrypt(chunk, key)
    if err != nil {
        log.Fatal(err)
    }
    // 处理加密后的数据块
    _ = encrypted
}
```

## 安全最佳实践

### 1. 密钥管理

```go
// ❌ 错误：硬编码密钥
key := []byte("my-secret-key-123")

// ✅ 正确：从环境变量或密钥管理系统获取
key := []byte(os.Getenv("AES_ENCRYPTION_KEY"))

// ✅ 正确：使用密钥派生函数
import "golang.org/x/crypto/scrypt"
key, _ := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
```

### 2. 密码存储

```go
// ❌ 错误：使用普通哈希
hash := sha256.Sum256([]byte(password))

// ❌ 错误：使用 MD5
hash := md5.Sum([]byte(password))

// ✅ 正确：使用 Bcrypt
hash, _ := crypt.Crypt.BcryptHash(password, 12)

// ✅ 正确：使用 PBKDF2
salt, _ := crypt.Crypt.GenerateSalt(16)
hash := crypt.Crypt.PBKDF2Hash(password, string(salt), 100000, 32)
```

### 3. 数据验证

```go
// 使用常数时间比较防止时序攻击
if crypt.Crypt.SecureEqual(receivedMAC, calculatedMAC) {
    // 验证通过
}

// 而不是直接比较
if receivedMAC == calculatedMAC { // ❌ 可能受到时序攻击
    // 验证通过
}
```

### 4. 错误信息

```go
// ❌ 错误：泄露过多信息
if err != nil {
    return fmt.Errorf("密码错误: %s", userPassword)
}

// ✅ 正确：使用通用错误信息
if err != nil {
    return fmt.Errorf("认证失败")
}
```

### 5. 随机数生成

```go
// ❌ 错误：使用不安全的随机数
rand.Seed(time.Now().UnixNano())
randomNumber := rand.Intn(1000000)

// ✅ 正确：使用加密安全的随机数
randomBytes, _ := crypt.Crypt.GenerateRandomBytes(4)
```

## 常见问题

### Q1: AES 和 RSA 应该如何选择？

**A:**
- **AES（对称加密）**：速度快，适合加密大量数据
- **RSA（非对称加密）**：速度慢，适合加密少量数据（如密钥）
- **最佳实践**：使用 RSA 加密 AES 密钥，使用 AES 加密实际数据（混合加密）

```go
// 混合加密示例
// 1. 生成 AES 密钥
aesKey, _ := crypt.Crypt.GenerateAESKey(crypt.AES256)

// 2. 使用 RSA 加密 AES 密钥
encryptedAESKey, _ := crypt.Crypt.RSAEncrypt(aesKey, rsaPublicKey)

// 3. 使用 AES 加密实际数据
encryptedData, _ := crypt.Crypt.AESEncryptToBase64(largeData, aesKey)

// 发送：encryptedAESKey + encryptedData
```

### Q2: Bcrypt 和 PBKDF2 应该如何选择？

**A:**
- **Bcrypt**：专为密码设计，自带盐值，使用简单，推荐用于大多数应用
- **PBKDF2**：更灵活，需要自己管理盐值，适合需要兼容性的场景
- **选择建议**：优先使用 Bcrypt，除非有特殊需求

### Q3: 如何选择 Bcrypt 的成本因子？

**A:**
- 从 `cost = 10` 开始
- 测试哈希操作耗时
- 如果耗时 < 100ms，增加成本因子
- 如果耗时 > 250ms，减少成本因子
- 目标：使哈希操作耗时 100-250ms

### Q4: HMAC 应该使用哪种哈希算法？

**A:**
- **SHA-256**：推荐，性能和安全性平衡良好
- **SHA-512**：更高安全性，但计算时间更长
- **MD5/SHA-1**：不推荐，仅用于兼容性

### Q5: 为什么需要常数时间比较？

**A:** 普通的字符串比较会因为第一个不同的字符而立即返回，攻击者可以通过测量响应时间来猜测正确的值。常数时间比较无论输入如何都花费相同时间，防止时序攻击。

## 完整示例

### 用户认证系统

```go
package main

import (
    "fmt"
    "log"

    "yourproject/util/crypt"
)

// User 用户结构
type User struct {
    Username     string
    PasswordHash string
}

// Register 用户注册
func Register(username, password string) (*User, error) {
    // 生成密码哈希
    hash, err := crypt.Crypt.BcryptHash(password, 12)
    if err != nil {
        return nil, fmt.Errorf("密码哈希失败: %w", err)
    }

    user := &User{
        Username:     username,
        PasswordHash: hash,
    }

    // 在实际应用中，这里应该保存到数据库
    return user, nil
}

// Login 用户登录
func Login(user *User, password string) error {
    // 验证密码
    if !crypt.Crypt.BcryptVerify(password, user.PasswordHash) {
        return fmt.Errorf("用户名或密码错误")
    }

    return nil
}

func main() {
    // 注册用户
    user, err := Register("alice", "SecurePassword123!")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("用户注册成功")

    // 登录 - 正确密码
    err = Login(user, "SecurePassword123!")
    if err != nil {
        fmt.Println("登录失败:", err)
    } else {
        fmt.Println("登录成功")
    }

    // 登录 - 错误密码
    err = Login(user, "WrongPassword")
    if err != nil {
        fmt.Println("登录失败:", err)
    } else {
        fmt.Println("登录成功")
    }
}
```

### 数据加密工具

```go
package main

import (
    "fmt"
    "log"
    "os"

    "yourproject/util/crypt"
)

// EncryptData 加密数据
func EncryptData(plaintext string) (string, error) {
    // 从环境变量获取密钥
    keyHex := os.Getenv("AES_ENCRYPTION_KEY")
    if keyHex == "" {
        return "", fmt.Errorf("未设置 AES_ENCRYPTION_KEY 环境变量")
    }

    key, err := crypt.Crypt.HexDecodeBytes(keyHex)
    if err != nil {
        return "", fmt.Errorf("密钥解码失败: %w", err)
    }

    // 加密
    encrypted, err := crypt.Crypt.AESEncryptToBase64(plaintext, key)
    if err != nil {
        return "", fmt.Errorf("加密失败: %w", err)
    }

    return encrypted, nil
}

// DecryptData 解密数据
func DecryptData(ciphertext string) (string, error) {
    // 从环境变量获取密钥
    keyHex := os.Getenv("AES_ENCRYPTION_KEY")
    if keyHex == "" {
        return "", fmt.Errorf("未设置 AES_ENCRYPTION_KEY 环境变量")
    }

    key, err := crypt.Crypt.HexDecodeBytes(keyHex)
    if err != nil {
        return "", fmt.Errorf("密钥解码失败: %w", err)
    }

    // 解密
    decrypted, err := crypt.Crypt.AESDecryptFromBase64(ciphertext, key)
    if err != nil {
        return "", fmt.Errorf("解密失败: %w", err)
    }

    return decrypted, nil
}

func main() {
    // 生成密钥（在实际应用中应该只生成一次并妥善保管）
    key, err := crypt.Crypt.GenerateAESKeyHex(crypt.AES256)
    if err != nil {
        log.Fatal(err)
    }

    // 设置环境变量（仅用于演示）
    os.Setenv("AES_ENCRYPTION_KEY", key)

    // 加密数据
    original := "这是需要加密的敏感信息"
    encrypted, err := EncryptData(original)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("原始数据: %s\n", original)
    fmt.Printf("加密数据: %s\n", encrypted)

    // 解密数据
    decrypted, err := DecryptData(encrypted)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("解密数据: %s\n", decrypted)
}
```

### API 签名验证

```go
package main

import (
    "crypto/sha256"
    "fmt"
    "log"

    "yourproject/util/crypt"
)

// APIClient API 客户端
type APIClient struct {
    apikey string
    secret string
}

// NewAPIClient 创建 API 客户端
func NewAPIClient(apikey, secret string) *APIClient {
    return &APIClient{
        apikey: apikey,
        secret: secret,
    }
}

// SignRequest 签名请求
func (c *APIClient) SignRequest(data []byte) string {
    signature := crypt.Crypt.HMACSignHexWithSHA256(data, []byte(c.secret))
    return signature
}

// VerifyRequest 验证请求
func (c *APIClient) VerifyRequest(data []byte, signature string) bool {
    expectedSignature := c.SignRequest(data)
    return crypt.Crypt.SecureEqual(signature, expectedSignature)
}

func main() {
    // 创建客户端
    client := NewAPIClient("my-apikey", "my-secret-key")

    // 准备请求数据
    request := []byte(`{"user":"alice","action":"transfer","amount":100}`)

    // 生成签名
    signature := client.SignRequest(request)
    fmt.Printf("签名: %s\n", signature)

    // 验证签名
    isValid := client.VerifyRequest(request, signature)
    fmt.Printf("签名有效: %v\n", isValid)

    // 篡改数据
    tamperedRequest := []byte(`{"user":"bob","action":"transfer","amount":1000}`)
    isValid = client.VerifyRequest(tamperedRequest, signature)
    fmt.Printf("篡改后的签名有效: %v\n", isValid)
}
```

## 许可证

本包是 litecore-go 项目的一部分，遵循项目许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.0.0
- 初始版本发布
- 支持 AES、RSA 加密
- 支持 Bcrypt、PBKDF2 密码哈希
- 支持 HMAC、ECDSA 签名
- 提供 Base64、Hex 编码解码
- 提供安全工具函数
