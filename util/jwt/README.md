# JWT 工具包

JWT（JSON Web Token）令牌生成、解析和验证工具库，支持多种签名算法。

## 特性

- **多种签名算法支持**：支持 HMAC（HS256/HS384/HS512）、RSA（RS256/RS384/RS512）和 ECDSA（ES256/ES384/ES512）签名算法
- **灵活的 Claims 结构**：提供标准 Claims（`StandardClaims`）和映射 Claims（`MapClaims`）两种数据结构
- **完善的验证机制**：支持令牌过期时间、生效时间、签发者、主题、受众等字段验证
- **便捷的辅助方法**：提供设置过期时间、签发时间、签发者等便捷方法
- **无外部依赖**：纯 Go 标准库实现，轻量且高效
- **性能优化**：使用 sync.Pool 复用对象，减少内存分配

## 快速开始

```go
package main

import (
    "time"

    "github.com/lite-lake/litecore-go/util/jwt"
)

func main() {
    // 定义密钥（实际使用中应从安全配置中读取）
    secretKey := []byte("your-secret-key")

    // 创建 Claims
    claims := jwt.JWT.NewStandardClaims()
    jwt.JWT.SetIssuer(claims, "my-app")
    jwt.JWT.SetSubject(claims, "user123")
    jwt.JWT.SetExpiration(claims, 24*time.Hour)
    jwt.JWT.SetIssuedAt(claims, time.Now())

    // 生成 JWT Token
    token, err := jwt.JWT.GenerateHS256Token(claims, secretKey)
    if err != nil {
        panic(err)
    }

    // 解析 JWT Token
    parsedClaims, err := jwt.JWT.ParseHS256Token(token, secretKey)
    if err != nil {
        panic(err)
    }

    // 验证 Claims
    err = jwt.JWT.ValidateClaims(parsedClaims)
    if err != nil {
        panic("Token 验证失败: " + err.Error())
    }
}
```

## HMAC 算法

HMAC 算法使用共享密钥进行签名和验证，适用于服务端生成的场景。

### HS256（推荐）

```go
secretKey := []byte("your-256-bit-secret")

// 生成 Token
claims := jwt.JWT.NewStandardClaims()
jwt.JWT.SetIssuer(claims, "my-app")
jwt.JWT.SetSubject(claims, "user123")
jwt.JWT.SetExpiration(claims, 24*time.Hour)

token, err := jwt.JWT.GenerateHS256Token(claims, secretKey)

// 解析 Token
parsedClaims, err := jwt.JWT.ParseHS256Token(token, secretKey)
```

### HS512

```go
secretKey := []byte("your-512-bit-secret")

// 生成 Token
claims := jwt.JWT.NewMapClaims()
jwt.JWT.AddCustomClaim(claims, "user_id", 12345)
jwt.JWT.SetExpiration(claims, 24*time.Hour)

token, err := jwt.JWT.GenerateHS512Token(claims, secretKey)

// 解析 Token
parsedClaims, err := jwt.JWT.ParseHS512Token(token, secretKey)
```

## RSA 算法

RSA 算法使用非对称密钥，私钥用于签名，公钥用于验证，适用于分布式系统。

```go
import (
    "crypto/rand"
    "crypto/rsa"
)

// 生成 RSA 密钥对
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
    panic(err)
}
publicKey := &privateKey.PublicKey

// 生成 Token
claims := jwt.JWT.NewMapClaims()
jwt.JWT.AddCustomClaim(claims, "user_id", 12345)
jwt.JWT.SetExpiration(claims, 24*time.Hour)

token, err := jwt.JWT.GenerateRS256Token(claims, privateKey)

// 解析 Token（使用公钥）
parsedClaims, err := jwt.JWT.ParseRS256Token(token, publicKey)
```

## ECDSA 算法

ECDSA 算法使用椭圆曲线加密，签名更短，性能更好。

```go
import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
)

// 生成 ECDSA P-256 密钥对
privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
if err != nil {
    panic(err)
}
publicKey := &privateKey.PublicKey

// 生成 Token
claims := jwt.JWT.NewMapClaims()
jwt.JWT.AddCustomClaim(claims, "user_id", 12345)
jwt.JWT.SetExpiration(claims, 24*time.Hour)

token, err := jwt.JWT.GenerateES256Token(claims, privateKey)

// 解析 Token
parsedClaims, err := jwt.JWT.ParseES256Token(token, publicKey)
```

## Claims 验证

```go
// 基础验证（只验证过期时间和生效时间）
err := jwt.JWT.ValidateClaims(parsedClaims)

// 完整验证（验证签发者、主题、受众等）
err = jwt.JWT.ValidateClaims(parsedClaims,
    jwt.WithIssuer("my-app"),
    jwt.WithSubject("user123"),
    jwt.WithAudience("my-app-users"),
)

// 测试自定义当前时间（用于单元测试）
past := time.Now().Add(-2 * time.Hour)
err = jwt.JWT.ValidateClaims(parsedClaims, jwt.WithCurrentTime(past))
```

## 自定义 Claims

### 使用 MapClaims

```go
claims := jwt.JWT.NewMapClaims()

// 设置标准字段
jwt.JWT.SetIssuer(claims, "my-app")
jwt.JWT.SetSubject(claims, "user123")
jwt.JWT.SetExpiration(claims, 24*time.Hour)

// 添加自定义字段
jwt.JWT.AddCustomClaim(claims, "user_id", 12345)
jwt.JWT.AddCustomClaim(claims, "role", "admin")
jwt.JWT.AddCustomClaim(claims, "permissions", []string{"read", "write"})

// 生成和解析 Token
token, _ := jwt.JWT.GenerateHS256Token(claims, secretKey)
parsedClaims, _ := jwt.JWT.ParseHS256Token(token, secretKey)

// 读取自定义字段
customClaims := parsedClaims.GetCustomClaims()
```

### 实现自定义 Claims 结构

```go
type CustomClaims struct {
    UserID    int64    `json:"user_id"`
    Email     string   `json:"email"`
    Roles     []string `json:"roles"`
    ExpiresAt int64    `json:"exp"`
    IssuedAt  int64    `json:"iat"`
}

// 实现 ILiteUtilJWTClaims 接口
func (c *CustomClaims) GetExpiresAt() *time.Time {
    if c.ExpiresAt == 0 {
        return nil
    }
    t := time.Unix(c.ExpiresAt, 0)
    return &t
}

func (c *CustomClaims) GetIssuedAt() *time.Time {
    if c.IssuedAt == 0 {
        return nil
    }
    t := time.Unix(c.IssuedAt, 0)
    return &t
}

func (c *CustomClaims) GetNotBefore() *time.Time {
    return nil
}

func (c *CustomClaims) GetIssuer() string {
    return ""
}

func (c *CustomClaims) GetSubject() string {
    return ""
}

func (c *CustomClaims) GetAudience() []string {
    return nil
}

func (c *CustomClaims) GetCustomClaims() map[string]interface{} {
    return map[string]interface{}{
        "user_id": c.UserID,
        "email":   c.Email,
        "roles":   c.Roles,
    }
}

func (c *CustomClaims) SetCustomClaims(claims map[string]interface{}) {
    // 实现略
}
```

## API 参考

### Token 生成

| 方法 | 算法 | 参数 | 返回 |
|------|------|------|------|
| `GenerateHS256Token` | HMAC SHA-256 | `claims ILiteUtilJWTClaims`, `secretKey []byte` | `(string, error)` |
| `GenerateHS512Token` | HMAC SHA-512 | `claims ILiteUtilJWTClaims`, `secretKey []byte` | `(string, error)` |
| `GenerateRS256Token` | RSA SHA-256 | `claims ILiteUtilJWTClaims`, `privateKey *rsa.PrivateKey` | `(string, error)` |
| `GenerateES256Token` | ECDSA P-256 | `claims ILiteUtilJWTClaims`, `privateKey *ecdsa.PrivateKey` | `(string, error)` |
| `GenerateToken` | 通用 | `claims`, `algorithm`, `secretKey`, `rsaPrivateKey`, `ecdsaPrivateKey` | `(string, error)` |

### Token 解析

| 方法 | 算法 | 参数 | 返回 |
|------|------|------|------|
| `ParseHS256Token` | HMAC SHA-256 | `token string`, `secretKey []byte` | `(MapClaims, error)` |
| `ParseHS512Token` | HMAC SHA-512 | `token string`, `secretKey []byte` | `(MapClaims, error)` |
| `ParseRS256Token` | RSA SHA-256 | `token string`, `publicKey *rsa.PublicKey` | `(MapClaims, error)` |
| `ParseES256Token` | ECDSA P-256 | `token string`, `publicKey *ecdsa.PublicKey` | `(MapClaims, error)` |
| `ParseToken` | 通用 | `token`, `algorithm`, `secretKey`, `rsaPublicKey`, `ecdsaPublicKey` | `(MapClaims, error)` |

### Claims 验证

| 方法 | 说明 |
|------|------|
| `ValidateClaims(claims, options...)` | 验证 Claims 的有效性，支持以下验证选项：`WithIssuer`、`WithSubject`、`WithAudience`、`WithCurrentTime` |

### 便捷方法

| 方法 | 说明 |
|------|------|
| `NewStandardClaims()` | 创建标准 Claims |
| `NewMapClaims()` | 创建映射 Claims |
| `SetExpiration(claims, duration)` | 设置过期时间（相对于当前时间） |
| `SetIssuedAt(claims, time)` | 设置签发时间 |
| `SetNotBefore(claims, time)` | 设置生效时间 |
| `SetIssuer(claims, issuer)` | 设置签发者 |
| `SetSubject(claims, subject)` | 设置主题 |
| `SetAudience(claims, audience...)` | 设置受众 |
| `AddCustomClaim(claims, key, value)` | 添加自定义声明 |

## 支持的算法

| 算法 | 类型 | 说明 |
|------|------|------|
| HS256 | HMAC | HMAC 使用 SHA-256 |
| HS384 | HMAC | HMAC 使用 SHA-384 |
| HS512 | HMAC | HMAC 使用 SHA-512 |
| RS256 | RSA | RSASSA-PKCS1-v1_5 使用 SHA-256 |
| RS384 | RSA | RSASSA-PKCS1-v1_5 使用 SHA-384 |
| RS512 | RSA | RSASSA-PKCS1-v1_5 使用 SHA-512 |
| ES256 | ECDSA | ECDSA 使用 P-256 和 SHA-256 |
| ES384 | ECDSA | ECDSA 使用 P-384 和 SHA-384 |
| ES512 | ECDSA | ECDSA 使用 P-521 和 SHA-512 |

## Claims 字段说明

| 字段 | 全称 | 类型 | 必需 | 描述 |
|------|------|------|------|------|
| `iss` | Issuer | string | 否 | 签发者 |
| `sub` | Subject | string | 否 | 主题（通常是用户 ID） |
| `aud` | Audience | string 或 []string | 否 | 受众 |
| `exp` | Expiration | int64 | 否 | 过期时间（Unix 时间戳） |
| `nbf` | Not Before | int64 | 否 | 生效时间（Unix 时间戳） |
| `iat` | Issued At | int64 | 否 | 签发时间（Unix 时间戳） |
| `jti` | JWT ID | string | 否 | JWT 唯一标识符 |

## 最佳实践

### 1. 密钥管理

```go
import (
    "crypto/rand"
    "encoding/base64"
)

// 生成安全的随机密钥
func generateSecretKey() ([]byte, error) {
    key := make([]byte, 32) // 256 位
    _, err := rand.Read(key)
    if err != nil {
        return nil, err
    }
    return key, nil
}
```

### 2. Token 刷新机制

```go
func refreshToken(oldToken string, secretKey []byte) (string, error) {
    // 解析旧 Token
    claims, err := jwt.JWT.ParseHS256Token(oldToken, secretKey)
    if err != nil {
        return "", err
    }

    // 检查是否在刷新窗口期内（例如过期前 1 小时）
    exp := claims.GetExpiresAt()
    if exp == nil || time.Until(*exp) > time.Hour {
        return "", errors.New("not in refresh window")
    }

    // 更新签发时间和过期时间
    jwt.JWT.SetIssuedAt(claims, time.Now())
    jwt.JWT.SetExpiration(claims, 24*time.Hour)

    // 生成新 Token
    return jwt.JWT.GenerateHS256Token(claims, secretKey)
}
```

### 3. HTTP 中间件示例

```go
import (
    "net/http"
    "strings"
)

func JWTAuthMiddleware(secretKey []byte) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 从 Authorization header 获取 token
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        // 解析 Bearer token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
            return
        }

        token := parts[1]

        // 解析和验证 token
        claims, err := jwt.JWT.ParseHS256Token(token, secretKey)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        err = jwt.JWT.ValidateClaims(claims)
        if err != nil {
            http.Error(w, "Token validation failed", http.StatusUnauthorized)
            return
        }
    }
}
```

## 常见问题

### Q: 如何选择合适的算法？

- **HS256**：适用于单体应用或服务端生成的场景，性能好，实现简单
- **RS256**：适用于分布式系统，私钥签名、公钥验证，安全性高
- **ES256**：适用于对 Token 大小敏感的场景，签名更短

### Q: MapClaims 中的数字为什么是 float64？

由于 JSON 标准不区分整数和浮点数，`encoding/json` 在解析时会将所有数字解析为 `float64`。使用时需要类型转换：

```go
userId := int(claims["user_id"].(float64))
```

### Q: 如何在生产环境中保护密钥？

- 使用环境变量或密钥管理服务（如 AWS Secrets Manager、HashiCorp Vault）
- 密钥应足够长（HMAC 至少 32 字节，RSA 至少 2048 位）
- 定期轮换密钥
- 不要将密钥硬编码在代码中
- 不要将密钥提交到版本控制系统
