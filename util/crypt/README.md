# Crypt - 加密解密工具包

提供全面的加密解密功能，包括对称加密、非对称加密、密码哈希、数字签名、编码转换等。

## 特性

- **对称加密**：支持 AES-128/192/256，使用 GCM 模式提供认证加密
- **非对称加密**：支持 RSA-1024/2048/3072/4096，使用 OAEP 填充
- **密码哈希**：支持 Bcrypt（推荐）和 PBKDF2 两种安全算法
- **数字签名**：支持 HMAC（SHA256/SHA512）和 ECDSA（P-256）
- **编码转换**：Base64、Hex 编码解码，支持 URL 安全格式
- **安全工具**：随机数生成、常数时间比较（防时序攻击）

## 快速开始

```go
package main

import (
    "fmt"
    "log"

    "github.com/lite-lake/litecore-go/util/crypt"
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

## Base64 编码解码

### 基本使用

```go
// 字符串编码解码
encoded := crypt.Crypt.Base64Encode("Hello, World!")
decoded, _ := crypt.Crypt.Base64Decode("SGVsbG8sIFdvcmxkIQ==")

// 字节数组编码解码
data := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}
encodedBytes := crypt.Crypt.Base64EncodeBytes(data)
decodedBytes, _ := crypt.Crypt.Base64DecodeBytes("SGVsbG8=")
```

### URL 安全编码

```go
// URL 安全的 Base64 编码
encoded := crypt.Crypt.Base64URLEncode("Hello, World!")
decoded, _ := crypt.Crypt.Base64URLDecode(encoded)
```

## Hex 编码解码

```go
// 十六进制编码解码
encoded := crypt.Crypt.HexEncode("Hello")
decoded, _ := crypt.Crypt.HexDecode("48656c6c6f")

// 字节数组编码解码
data := []byte{0x00, 0xFF, 0xAA, 0x55}
encodedBytes := crypt.Crypt.HexEncodeBytes(data)
decodedBytes, _ := crypt.Crypt.HexDecodeBytes("00ffaa55")
```

## AES 对称加密

### 生成密钥

```go
// 生成 AES-256 密钥
key, _ := crypt.Crypt.GenerateAESKey(crypt.AES256)

// 生成十六进制格式的密钥
keyHex, _ := crypt.Crypt.GenerateAESKeyHex(crypt.AES256)
```

### 加密解密

```go
// 方法 1: 字节数组
plaintext := []byte("敏感信息")
ciphertext, _ := crypt.Crypt.AESEncrypt(plaintext, key)
decrypted, _ := crypt.Crypt.AESDecrypt(ciphertext, key)

// 方法 2: Base64 编码（推荐）
encrypted, _ := crypt.Crypt.AESEncryptToBase64("敏感信息", key)
decrypted, _ := crypt.Crypt.AESDecryptFromBase64(encrypted, key)
```

## RSA 非对称加密

### 生成密钥对

```go
// 生成 RSA-2048 密钥对
privateKey, publicKey, _ := crypt.Crypt.GenerateRSAKeys(crypt.RSA2048)
```

### 加密解密

```go
// 方法 1: 字节数组
plaintext := []byte("秘密消息")
ciphertext, _ := crypt.Crypt.RSAEncrypt(plaintext, publicKey)
decrypted, _ := crypt.Crypt.RSADecrypt(ciphertext, privateKey)

// 方法 2: Base64 编码（推荐）
encrypted, _ := crypt.Crypt.RSAEncryptToBase64("秘密消息", publicKey)
decrypted, _ := crypt.Crypt.RSADecryptFromBase64(encrypted, privateKey)
```

## 密码哈希

### Bcrypt 哈希（推荐）

```go
// 生成哈希
hash, _ := crypt.Crypt.BcryptHash("mypassword123", 10)

// 验证密码
isCorrect := crypt.Crypt.BcryptVerify("mypassword123", hash)
```

**成本因子建议：**
- `cost = 4`：测试环境
- `cost = 10`：生产环境（推荐）
- `cost = 12+`：高安全性

### PBKDF2 哈希

```go
// 生成盐值
salt, _ := crypt.Crypt.GenerateSalt(16)

// 生成哈希
hash := crypt.Crypt.PBKDF2Hash("mypassword123", string(salt), 10000, 32)

// 验证密码
isCorrect := crypt.Crypt.PBKDF2Verify("mypassword123", string(salt), hash, 10000, 32)
```

**参数建议：**
- `salt`：至少 16 字节
- `iterations`：最少 10000 次，推荐 100000+ 次
- `keyLen`：通常 32 字节（256 位）

## HMAC 签名

```go
data := []byte("需要签名的数据")
key := []byte("密钥")

// 使用 SHA256 的便捷方法
signature := crypt.Crypt.HMACSignWithSHA256(data, key)
signatureHex := crypt.Crypt.HMACSignHexWithSHA256(data, key)

// 验证签名
isValid := crypt.Crypt.HMACVerify(data, key, signature, crypto.SHA256.New)
```

## ECDSA 数字签名

```go
// 生成密钥对
privateKey, publicKey, _ := crypt.Crypt.GenerateECDSAKeys()

// 签名
data := []byte("需要签名的数据")
signature, _ := crypt.Crypt.ECDSASign(data, privateKey)
signatureHex, _ := crypt.Crypt.ECDSASignHex(data, privateKey)

// 验证签名
isValid := crypt.Crypt.ECDSAVerify(data, signature, publicKey)
isValid, _ := crypt.Crypt.ECDSAVerifyHex(data, signatureHex, publicKey)
```

## 安全工具函数

### 随机数生成

```go
// 生成随机字节
randomBytes, _ := crypt.Crypt.GenerateRandomBytes(16)

// 生成随机字符串
randomString, _ := crypt.Crypt.GenerateRandomString(32)
```

### 常数时间比较

```go
// 防止时序攻击的比较
a := []byte("sensitive-data")
b := []byte("sensitive-data")
isEqual := crypt.Crypt.ConstantTimeCompare(a, b)

// 安全字符串比较
isEqual = crypt.Crypt.SecureEqual("password123", "password123")
```

## API 参考

### 编码解码

| 函数 | 说明 |
|------|------|
| `Base64Encode(data string) string` | Base64 编码字符串 |
| `Base64EncodeBytes(data []byte) string` | Base64 编码字节数组 |
| `Base64Decode(data string) (string, error)` | Base64 解码为字符串 |
| `Base64DecodeBytes(data string) ([]byte, error)` | Base64 解码为字节数组 |
| `Base64URLEncode(data string) string` | URL 安全的 Base64 编码 |
| `Base64URLDecode(data string) (string, error)` | URL 安全的 Base64 解码 |

### Hex 编码解码

| 函数 | 说明 |
|------|------|
| `HexEncode(data string) string` | 十六进制编码字符串 |
| `HexEncodeBytes(data []byte) string` | 十六进制编码字节数组 |
| `HexDecode(data string) (string, error)` | 十六进制解码为字符串 |
| `HexDecodeBytes(data string) ([]byte, error)` | 十六进制解码为字节数组 |

### AES 对称加密

| 函数 | 说明 |
|------|------|
| `GenerateAESKey(keySize AESKeySize) ([]byte, error)` | 生成 AES 密钥 |
| `GenerateAESKeyHex(keySize AESKeySize) (string, error)` | 生成十六进制 AES 密钥 |
| `AESEncrypt(plaintext, key []byte) ([]byte, error)` | AES 加密 |
| `AESEncryptToBase64(plaintext string, key []byte) (string, error)` | AES 加密并 Base64 编码 |
| `AESDecrypt(ciphertext, key []byte) ([]byte, error)` | AES 解密 |
| `AESDecryptFromBase64(ciphertext string, key []byte) (string, error)` | 从 Base64 字符串 AES 解密 |

### RSA 非对称加密

| 函数 | 说明 |
|------|------|
| `GenerateRSAKeys(bits RSABits) (*rsa.PrivateKey, *rsa.PublicKey, error)` | 生成 RSA 密钥对 |
| `RSAEncrypt(plaintext []byte, publicKey *rsa.PublicKey) ([]byte, error)` | RSA 公钥加密 |
| `RSAEncryptToBase64(plaintext string, publicKey *rsa.PublicKey) (string, error)` | RSA 加密并 Base64 编码 |
| `RSADecrypt(ciphertext []byte, privateKey *rsa.PrivateKey) ([]byte, error)` | RSA 私钥解密 |
| `RSADecryptFromBase64(ciphertext string, privateKey *rsa.PrivateKey) (string, error)` | 从 Base64 字符串 RSA 解密 |

### 密码哈希

| 函数 | 说明 |
|------|------|
| `BcryptHash(password string, cost int) (string, error)` | Bcrypt 密码哈希 |
| `BcryptVerify(password, hash string) bool` | Bcrypt 密码验证 |
| `PBKDF2Hash(password, salt string, iterations, keyLen int) string` | PBKDF2 密码哈希 |
| `PBKDF2Verify(password, salt, hash string, iterations, keyLen int) bool` | PBKDF2 密码验证 |
| `GenerateSalt(length int) ([]byte, error)` | 生成随机盐值 |
| `GenerateSaltHex(length int) (string, error)` | 生成十六进制盐值 |

### HMAC 签名

| 函数 | 说明 |
|------|------|
| `HMACSign(data, key []byte, hashFunc func() hash.Hash) []byte` | HMAC 签名 |
| `HMACSignHex(data, key []byte, hashFunc func() hash.Hash) string` | HMAC 签名并转十六进制 |
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

### 工具函数

| 函数 | 说明 |
|------|------|
| `ConstantTimeCompare(a, b []byte) bool` | 常数时间比较（防时序攻击） |
| `SecureEqual(a, b string) bool` | 安全字符串比较 |
| `GenerateRandomBytes(length int) ([]byte, error)` | 生成随机字节 |
| `GenerateRandomString(length int) (string, error)` | 生成随机字符串 |

## 常量定义

### AES 密钥大小

```go
const (
    AES128 AESKeySize = 16  // 128 位密钥
    AES192 AESKeySize = 24  // 192 位密钥
    AES256 AESKeySize = 32  // 256 位密钥
)
```

### RSA 密钥位数

```go
const (
    RSA1024 RSABits = 1024  // 不推荐用于生产环境
    RSA2048 RSABits = 2048  // 推荐
    RSA3072 RSABits = 3072  // 高安全性
    RSA4096 RSABits = 4096  // 最高安全性
)
```

## 错误处理

所有加密解密函数均返回 `error` 类型，应始终检查并处理错误：

```go
encrypted, err := crypt.Crypt.AESEncryptToBase64(plaintext, key)
if err != nil {
    log.Printf("加密失败: %v", err)
    return
}
```

## 安全最佳实践

### 1. 密钥管理

```go
// ✅ 从环境变量获取密钥
key := []byte(os.Getenv("AES_ENCRYPTION_KEY"))

// ❌ 避免硬编码密钥
key := []byte("my-secret-key-123")
```

### 2. 密码存储

```go
// ✅ 使用 Bcrypt 哈希密码
hash, _ := crypt.Crypt.BcryptHash(password, 12)

// ❌ 避免使用普通哈希
hash := sha256.Sum256([]byte(password))
```

### 3. 安全比较

```go
// ✅ 使用常数时间比较
if crypt.Crypt.SecureEqual(receivedMAC, calculatedMAC) {
    // 验证通过
}
```

### 4. 随机数生成

```go
// ✅ 使用加密安全的随机数
randomBytes, _ := crypt.Crypt.GenerateRandomBytes(16)

// ❌ 避免使用不安全的随机数
rand.Seed(time.Now().UnixNano())
```

## 性能考虑

- **Bcrypt 成本因子**：生产环境推荐 10-12，使哈希操作耗时 100-250ms
- **PBKDF2 迭代次数**：推荐至少 100,000 次，根据服务器性能调整
- **AES vs RSA**：RSA 较慢，大量数据加密建议使用 AES+RSA 混合加密

## 使用示例

### 混合加密（RSA + AES）

```go
// 1. 生成 AES 密钥
aesKey, _ := crypt.Crypt.GenerateAESKey(crypt.AES256)

// 2. 使用 RSA 加密 AES 密钥
privateKey, publicKey, _ := crypt.Crypt.GenerateRSAKeys(crypt.RSA2048)
encryptedAESKey, _ := crypt.Crypt.RSAEncrypt(aesKey, publicKey)

// 3. 使用 AES 加密实际数据
encryptedData, _ := crypt.Crypt.AESEncryptToBase64("大量敏感数据", aesKey)

// 发送：encryptedAESKey + encryptedData
```

### API 签名验证

```go
// 生成签名
data := []byte(`{"user":"alice","action":"transfer","amount":100}`)
secret := []byte("my-secret-key")
signature := crypt.Crypt.HMACSignHexWithSHA256(data, secret)

// 验证签名
expectedSig := crypt.Crypt.HMACSignHexWithSHA256(data, secret)
isValid := crypt.Crypt.SecureEqual(signature, expectedSig)
```
