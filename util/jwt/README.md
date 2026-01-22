# JWT 工具包

JWT (JSON Web Token) 令牌生成、解析和验证工具库，支持多种签名算法。

## 特性

- **多种签名算法支持**：支持 HMAC (HS256/HS384/HS512)、RSA (RS256/RS384/RS512) 和 ECDSA (ES256/ES384/ES512) 签名算法
- **灵活的 Claims 结构**：提供标准 Claims (`StandardClaims`) 和映射 Claims (`MapClaims`) 两种数据结构
- **完善的验证机制**：支持令牌过期时间、生效时间、签发者、主题、受众等字段验证
- **便捷的辅助方法**：提供设置过期时间、签发时间、签发者等便捷方法
- **类型安全接口**：通过接口定义统一操作，支持扩展和自定义实现
- **无外部依赖**：纯 Go 标准库实现，轻量且高效

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
    "time"

    "github.com/lite-lake/litecore-go/util/jwt"
)

func main() {
    // 定义密钥（实际使用中应从安全配置中读取）
    secretKey := []byte("your-secret-key")

    // 创建 Claims
    claims := jwt.MapClaims{
        "user_id": float64(12345),
        "username": "john_doe",
        "role": "admin",
        "exp": float64(time.Now().Add(24 * time.Hour).Unix()),
        "iat": float64(time.Now().Unix()),
    }

    // 生成 JWT Token
    token, err := jwt.JWT.GenerateHS256Token(claims, secretKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("生成的 Token:", token)

    // 解析 JWT Token
    parsedClaims, err := jwt.JWT.ParseHS256Token(token, secretKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("解析的用户名:", parsedClaims["username"])

    // 验证 Claims
    err = jwt.JWT.ValidateClaims(parsedClaims)
    if err != nil {
        panic("Token 验证失败: " + err.Error())
    }

    fmt.Println("Token 验证成功")
}
```

## 功能详解

### HMAC 算法

HMAC 算法使用共享密钥进行签名和验证，适用于服务端生成的场景。

#### HS256 (HMAC SHA-256)

```go
package main

import (
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/util/jwt"
)

func main() {
    secretKey := []byte("your-256-bit-secret")

    // 创建标准 Claims
    claims := jwt.JWT.NewStandardClaims()
    jwt.JWT.SetIssuer(claims, "my-app")
    jwt.JWT.SetSubject(claims, "user123")
    jwt.JWT.SetAudience(claims, "my-app-users")
    jwt.JWT.SetExpiration(claims, 24*time.Hour)
    jwt.JWT.SetIssuedAt(claims, time.Now())

    // 生成 Token
    token, err := jwt.JWT.GenerateHS256Token(claims, secretKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("HS256 Token:", token)

    // 解析 Token
    parsedClaims, err := jwt.JWT.ParseHS256Token(token, secretKey)
    if err != nil {
        panic(err)
    }

    fmt.Printf("签发者: %s\n", parsedClaims.GetIssuer())
    fmt.Printf("主题: %s\n", parsedClaims.GetSubject())
}
```

#### HS512 (HMAC SHA-512)

```go
secretKey := []byte("your-512-bit-secret")

// 创建自定义 Claims
claims := jwt.MapClaims{
    "user_id": float64(12345),
    "email": "user@example.com",
    "exp": float64(time.Now().Add(7 * 24 * time.Hour).Unix()),
}

// 生成 Token
token, err := jwt.JWT.GenerateHS512Token(claims, secretKey)
if err != nil {
    panic(err)
}

// 解析 Token
parsedClaims, err := jwt.JWT.ParseHS512Token(token, secretKey)
if err != nil {
    panic(err)
}
```

### RSA 算法

RSA 算法使用非对称密钥，私钥用于签名，公钥用于验证，适用于分布式系统。

#### 生成 RSA 密钥对

```go
package main

import (
    "crypto/rand"
    "crypto/rsa"
    "fmt"

    "github.com/lite-lake/litecore-go/util/jwt"
)

func main() {
    // 生成 2048 位 RSA 密钥对
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        panic(err)
    }
    publicKey := &privateKey.PublicKey

    // 创建 Claims
    claims := jwt.MapClaims{
        "user_id": float64(12345),
        "role": "admin",
        "exp": float64(time.Now().Add(24 * time.Hour).Unix()),
    }

    // 使用私钥生成 Token
    token, err := jwt.JWT.GenerateRS256Token(claims, privateKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("RS256 Token:", token)

    // 使用公钥解析 Token
    parsedClaims, err := jwt.JWT.ParseRS256Token(token, publicKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("用户 ID:", parsedClaims["user_id"])
    fmt.Println("角色:", parsedClaims["role"])
}
```

#### RS256 常量签名场景

```go
import (
    "crypto/rsa"
    "os"
)

// 从 PEM 文件加载私钥
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
    // 实现略 - 读取 PEM 文件并解析
    // 可以使用 x509.ParsePKCS1PrivateKey 或 x509.ParsePKCS8PrivateKey
    return nil, nil
}

// 从 PEM 文件加载公钥
func loadPublicKey(path string) (*rsa.PublicKey, error) {
    // 实现略 - 读取 PEM 文件并解析
    return nil, nil
}
```

### ECDSA 算法

ECDSA 算法使用椭圆曲线加密，签名更短，性能更好。

```go
package main

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/util/jwt"
)

func main() {
    // 生成 ECDSA P-256 密钥对
    privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        panic(err)
    }
    publicKey := &privateKey.PublicKey

    // 创建 Claims
    claims := jwt.MapClaims{
        "user_id": float64(12345),
        "email": "user@example.com",
        "exp": float64(time.Now().Add(24 * time.Hour).Unix()),
    }

    // 生成 Token
    token, err := jwt.JWT.GenerateES256Token(claims, privateKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("ES256 Token:", token)

    // 解析 Token
    parsedClaims, err := jwt.JWT.ParseES256Token(token, publicKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("邮箱:", parsedClaims["email"])
}
```

### Claims 验证

提供灵活的 Claims 验证机制，支持多种验证选项。

```go
package main

import (
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/util/jwt"
)

func main() {
    secretKey := []byte("your-secret-key")

    // 创建包含过期时间的 Claims
    claims := jwt.MapClaims{
        "iss": "my-app",
        "sub": "user123",
        "aud": "my-app-users",
        "exp": float64(time.Now().Add(24 * time.Hour).Unix()),
        "nbf": float64(time.Now().Unix()),
    }

    // 生成 Token
    token, _ := jwt.JWT.GenerateHS256Token(claims, secretKey)

    // 解析 Token
    parsedClaims, _ := jwt.JWT.ParseHS256Token(token, secretKey)

    // 基础验证（只验证过期时间和生效时间）
    err := jwt.JWT.ValidateClaims(parsedClaims)
    if err != nil {
        panic("基础验证失败: " + err.Error())
    }

    // 完整验证（验证签发者、主题、受众等）
    err = jwt.JWT.ValidateClaims(parsedClaims,
        jwt.WithIssuer("my-app"),
        jwt.WithSubject("user123"),
        jwt.WithAudience("my-app-users"),
    )
    if err != nil {
        panic("完整验证失败: " + err.Error())
    }

    fmt.Println("所有验证通过")

    // 测试过期 Token
    expiredClaims := jwt.MapClaims{
        "exp": float64(time.Now().Add(-1 * time.Hour).Unix()),
    }
    err = jwt.JWT.ValidateClaims(expiredClaims)
    if err != nil {
        fmt.Println("过期验证:", err.Error()) // "token is expired"
    }

    // 测试未生效 Token
    futureClaims := jwt.MapClaims{
        "nbf": float64(time.Now().Add(1 * time.Hour).Unix()),
    }
    err = jwt.JWT.ValidateClaims(futureClaims)
    if err != nil {
        fmt.Println("未生效验证:", err.Error()) // "token is not valid yet"
    }

    // 测试自定义当前时间（用于单元测试）
    past := time.Now().Add(-2 * time.Hour)
    futureTime := time.Now().Add(1 * time.Hour)
    testClaims := jwt.MapClaims{
        "exp": float64(futureTime.Unix()),
    }
    err = jwt.JWT.ValidateClaims(testClaims, jwt.WithCurrentTime(past))
    if err != nil {
        panic("不应在指定时间点失败")
    }
}
```

### 便捷方法

提供一系列便捷方法用于设置 Claims 字段。

```go
package main

import (
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/util/jwt"
)

func main() {
    // 使用 StandardClaims
    claims1 := jwt.JWT.NewStandardClaims()
    jwt.JWT.SetExpiration(claims1, 24*time.Hour)
    jwt.JWT.SetIssuedAt(claims1, time.Now())
    jwt.JWT.SetNotBefore(claims1, time.Now())
    jwt.JWT.SetIssuer(claims1, "my-app")
    jwt.JWT.SetSubject(claims1, "user123")
    jwt.JWT.SetAudience(claims1, "audience1", "audience2")

    fmt.Printf("StandardClaims: %+v\n", claims1)

    // 使用 MapClaims
    claims2 := jwt.JWT.NewMapClaims()
    jwt.JWT.SetExpiration(claims2, 24*time.Hour)
    jwt.JWT.SetIssuer(claims2, "my-app")
    jwt.JWT.SetSubject(claims2, "user123")

    // 添加自定义字段
    jwt.JWT.AddCustomClaim(claims2, "user_id", 12345)
    jwt.JWT.AddCustomClaim(claims2, "role", "admin")
    jwt.JWT.AddCustomClaim(claims2, "permissions", []string{"read", "write", "delete"})

    fmt.Printf("MapClaims: %+v\n", claims2)

    // 获取自定义字段
    customClaims := claims2.GetCustomClaims()
    fmt.Println("自定义字段:", customClaims)

    // 获取标准字段
    fmt.Println("签发者:", claims2.GetIssuer())
    fmt.Println("主题:", claims2.GetSubject())
    fmt.Println("受众:", claims2.GetAudience())
    fmt.Println("过期时间:", claims2.GetExpiresAt())
    fmt.Println("签发时间:", claims2.GetIssuedAt())
    fmt.Println("生效时间:", claims2.GetNotBefore())
}
```

### 自定义 Claims 结构

实现 `ILiteUtilJWTClaims` 接口以支持自定义 Claims 结构。

```go
package main

import (
    "time"

    "github.com/lite-lake/litecore-go/util/jwt"
)

// CustomClaims 自定义 Claims 结构
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

func main() {
    secretKey := []byte("your-secret-key")

    claims := &CustomClaims{
        UserID:    12345,
        Email:     "user@example.com",
        Roles:     []string{"admin", "user"},
        ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
        IssuedAt:  time.Now().Unix(),
    }

    token, err := jwt.JWT.GenerateHS256Token(claims, secretKey)
    if err != nil {
        panic(err)
    }

    parsedClaims, err := jwt.JWT.ParseHS256Token(token, secretKey)
    if err != nil {
        panic(err)
    }

    // 从解析的 MapClaims 中读取自定义字段
    userID := int(parsedClaims["user_id"].(float64))
    email := parsedClaims["email"].(string)
    roles := parsedClaims["roles"].([]interface{})

    println("用户 ID:", userID)
    println("邮箱:", email)
    println("角色:", roles[0].(string))
}
```

## API 参考

### 类型定义

#### JWTAlgorithm

JWT 签名算法类型。

```go
type JWTAlgorithm string

const (
    HS256 JWTAlgorithm = "HS256"  // HMAC 使用 SHA-256
    HS384 JWTAlgorithm = "HS384"  // HMAC 使用 SHA-384
    HS512 JWTAlgorithm = "HS512"  // HMAC 使用 SHA-512
    RS256 JWTAlgorithm = "RS256"  // RSASSA-PKCS1-v1_5 使用 SHA-256
    RS384 JWTAlgorithm = "RS384"  // RSASSA-PKCS1-v1_5 使用 SHA-384
    RS512 JWTAlgorithm = "RS512"  // RSASSA-PKCS1-v1_5 使用 SHA-512
    ES256 JWTAlgorithm = "ES256"  // ECDSA 使用 P-256 和 SHA-256
    ES384 JWTAlgorithm = "ES384"  // ECDSA 使用 P-384 和 SHA-384
    ES512 JWTAlgorithm = "ES512"  // ECDSA 使用 P-521 和 SHA-512
)
```

#### StandardClaims

标准 JWT Claims 结构体。

```go
type StandardClaims struct {
    Audience  []string `json:"aud,omitempty"`  // 受众
    ExpiresAt int64    `json:"exp,omitempty"`  // 过期时间（Unix 时间戳）
    ID        string   `json:"jti,omitempty"`  // JWT ID
    IssuedAt  int64    `json:"iat,omitempty"`  // 签发时间（Unix 时间戳）
    Issuer    string   `json:"iss,omitempty"`  // 签发者
    NotBefore int64    `json:"nbf,omitempty"`  // 生效时间（Unix 时间戳）
    Subject   string   `json:"sub,omitempty"`  // 主题
}
```

#### MapClaims

映射形式的 JWT Claims，支持自定义字段。

```go
type MapClaims map[string]interface{}
```

#### ILiteUtilJWTClaims

JWT Claims 接口，所有 Claims 类型必须实现此接口。

```go
type ILiteUtilJWTClaims interface {
    GetExpiresAt() *time.Time
    GetIssuedAt() *time.Time
    GetNotBefore() *time.Time
    GetIssuer() string
    GetSubject() string
    GetAudience() []string
    GetCustomClaims() map[string]interface{}
    SetCustomClaims(claims map[string]interface{})
}
```

### Token 生成

#### GenerateHS256Token

使用 HMAC SHA-256 算法生成 JWT。

```go
func (j *jwtEngine) GenerateHS256Token(claims ILiteUtilJWTClaims, secretKey []byte) (string, error)
```

**参数：**
- `claims`: JWT Claims 数据
- `secretKey`: HMAC 密钥（推荐至少 32 字节）

**返回：**
- `string`: 生成的 JWT Token
- `error`: 错误信息

**示例：**

```go
secretKey := []byte("your-256-bit-secret")
claims := jwt.MapClaims{
    "user_id": float64(12345),
    "exp": float64(time.Now().Add(24 * time.Hour).Unix()),
}

token, err := jwt.JWT.GenerateHS256Token(claims, secretKey)
```

#### GenerateHS512Token

使用 HMAC SHA-512 算法生成 JWT。

```go
func (j *jwtEngine) GenerateHS512Token(claims ILiteUtilJWTClaims, secretKey []byte) (string, error)
```

#### GenerateRS256Token

使用 RSA SHA-256 算法生成 JWT。

```go
func (j *jwtEngine) GenerateRS256Token(claims ILiteUtilJWTClaims, privateKey *rsa.PrivateKey) (string, error)
```

**参数：**
- `claims`: JWT Claims 数据
- `privateKey`: RSA 私钥（推荐至少 2048 位）

**示例：**

```go
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
    panic(err)
}

claims := jwt.MapClaims{
    "user_id": float64(12345),
    "exp": float64(time.Now().Add(24 * time.Hour).Unix()),
}

token, err := jwt.JWT.GenerateRS256Token(claims, privateKey)
```

#### GenerateES256Token

使用 ECDSA P-256 算法生成 JWT。

```go
func (j *jwtEngine) GenerateES256Token(claims ILiteUtilJWTClaims, privateKey *ecdsa.PrivateKey) (string, error)
```

**参数：**
- `claims`: JWT Claims 数据
- `privateKey`: ECDSA 私钥

**示例：**

```go
privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
if err != nil {
    panic(err)
}

claims := jwt.MapClaims{
    "user_id": float64(12345),
    "exp": float64(time.Now().Add(24 * time.Hour).Unix()),
}

token, err := jwt.JWT.GenerateES256Token(claims, privateKey)
```

#### GenerateToken

通用 Token 生成方法，支持所有算法。

```go
func (j *jwtEngine) GenerateToken(
    claims ILiteUtilJWTClaims,
    algorithm JWTAlgorithm,
    secretKey []byte,
    rsaPrivateKey *rsa.PrivateKey,
    ecdsaPrivateKey *ecdsa.PrivateKey,
) (string, error)
```

### Token 解析

#### ParseHS256Token

解析使用 HMAC SHA-256 算法签名的 JWT。

```go
func (j *jwtEngine) ParseHS256Token(token string, secretKey []byte) (MapClaims, error)
```

**参数：**
- `token`: JWT Token 字符串
- `secretKey`: HMAC 密钥

**返回：**
- `MapClaims`: 解析后的 Claims
- `error`: 错误信息（签名验证失败或格式错误时返回错误）

**示例：**

```go
secretKey := []byte("your-256-bit-secret")
parsedClaims, err := jwt.JWT.ParseHS256Token(token, secretKey)
if err != nil {
    panic(err)
}

userId := int(parsedClaims["user_id"].(float64))
```

#### ParseHS512Token

解析使用 HMAC SHA-512 算法签名的 JWT。

```go
func (j *jwtEngine) ParseHS512Token(token string, secretKey []byte) (MapClaims, error)
```

#### ParseRS256Token

解析使用 RSA SHA-256 算法签名的 JWT。

```go
func (j *jwtEngine) ParseRS256Token(token string, publicKey *rsa.PublicKey) (MapClaims, error)
```

**参数：**
- `token`: JWT Token 字符串
- `publicKey`: RSA 公钥

**示例：**

```go
publicKey := &privateKey.PublicKey
parsedClaims, err := jwt.JWT.ParseRS256Token(token, publicKey)
if err != nil {
    panic(err)
}
```

#### ParseES256Token

解析使用 ECDSA P-256 算法签名的 JWT。

```go
func (j *jwtEngine) ParseES256Token(token string, publicKey *ecdsa.PublicKey) (MapClaims, error)
```

**参数：**
- `token`: JWT Token 字符串
- `publicKey`: ECDSA 公钥

**示例：**

```go
publicKey := &privateKey.PublicKey
parsedClaims, err := jwt.JWT.ParseES256Token(token, publicKey)
if err != nil {
    panic(err)
}
```

#### ParseToken

通用 Token 解析方法，支持所有算法。

```go
func (j *jwtEngine) ParseToken(
    token string,
    algorithm JWTAlgorithm,
    secretKey []byte,
    rsaPublicKey *rsa.PublicKey,
    ecdsaPublicKey *ecdsa.PublicKey,
) (MapClaims, error)
```

### Claims 验证

#### ValidateClaims

验证 Claims 的有效性。

```go
func (j *jwtEngine) ValidateClaims(claims ILiteUtilJWTClaims, options ...ValidateOption) error
```

**参数：**
- `claims`: 要验证的 Claims
- `options`: 验证选项（可选）

**返回：**
- `error`: 验证失败时返回错误

**验证项：**
- 过期时间 (`exp`)：当前时间超过过期时间时返回错误
- 生效时间 (`nbf`)：当前时间早于生效时间时返回错误
- 签发者 (`iss`)：使用 `WithIssuer` 选项时验证
- 主题 (`sub`)：使用 `WithSubject` 选项时验证
- 受众 (`aud`)：使用 `WithAudience` 选项时验证

**验证选项：**

```go
// 设置签发者验证
func WithIssuer(issuer string) ValidateOption

// 设置主题验证
func WithSubject(subject string) ValidateOption

// 设置受众验证
func WithAudience(audience ...string) ValidateOption

// 设置当前时间（用于测试）
func WithCurrentTime(t time.Time) ValidateOption
```

**示例：**

```go
err := jwt.JWT.ValidateClaims(parsedClaims,
    jwt.WithIssuer("my-app"),
    jwt.WithSubject("user123"),
    jwt.WithAudience("my-app-users"),
)
if err != nil {
    panic("验证失败: " + err.Error())
}
```

### 便捷方法

#### NewStandardClaims

创建标准 Claims。

```go
func (j *jwtEngine) NewStandardClaims() *StandardClaims
```

#### NewMapClaims

创建映射 Claims。

```go
func (j *jwtEngine) NewMapClaims() MapClaims
```

#### SetExpiration

设置 Claims 过期时间。

```go
func (j *jwtEngine) SetExpiration(claims ILiteUtilJWTClaims, duration time.Duration)
```

**参数：**
- `claims`: Claims 对象
- `duration`: 过期时长（相对于当前时间）

**示例：**

```go
claims := jwt.JWT.NewStandardClaims()
jwt.JWT.SetExpiration(claims, 24*time.Hour) // 24 小时后过期
```

#### SetIssuedAt

设置 Claims 签发时间。

```go
func (j *jwtEngine) SetIssuedAt(claims ILiteUtilJWTClaims, t time.Time)
```

#### SetNotBefore

设置 Claims 生效时间。

```go
func (j *jwtEngine) SetNotBefore(claims ILiteUtilJWTClaims, t time.Time)
```

#### SetIssuer

设置 Claims 签发者。

```go
func (j *jwtEngine) SetIssuer(claims ILiteUtilJWTClaims, issuer string)
```

#### SetSubject

设置 Claims 主题。

```go
func (j *jwtEngine) SetSubject(claims ILiteUtilJWTClaims, subject string)
```

#### SetAudience

设置 Claims 受众。

```go
func (j *jwtEngine) SetAudience(claims ILiteUtilJWTClaims, audience ...string)
```

#### AddCustomClaim

添加自定义声明。

```go
func (j *jwtEngine) AddCustomClaim(claims ILiteUtilJWTClaims, key string, value interface{})
```

**示例：**

```go
claims := jwt.JWT.NewMapClaims()
jwt.JWT.AddCustomClaim(claims, "user_id", 12345)
jwt.JWT.AddCustomClaim(claims, "role", "admin")
jwt.JWT.AddCustomClaim(claims, "permissions", []string{"read", "write"})
```

## 支持的算法

| 算法 | 类型 | 描述 | 生成方法 | 解析方法 |
|------|------|------|----------|----------|
| HS256 | HMAC | HMAC 使用 SHA-256 | `GenerateHS256Token` | `ParseHS256Token` |
| HS384 | HMAC | HMAC 使用 SHA-384 | `GenerateToken` (使用 `HS384`) | `ParseToken` (使用 `HS384`) |
| HS512 | HMAC | HMAC 使用 SHA-512 | `GenerateHS512Token` | `ParseHS512Token` |
| RS256 | RSA | RSASSA-PKCS1-v1_5 使用 SHA-256 | `GenerateRS256Token` | `ParseRS256Token` |
| RS384 | RSA | RSASSA-PKCS1-v1_5 使用 SHA-384 | `GenerateToken` (使用 `RS384`) | `ParseToken` (使用 `RS384`) |
| RS512 | RSA | RSASSA-PKCS1-v1_5 使用 SHA-512 | `GenerateToken` (使用 `RS512`) | `ParseToken` (使用 `RS512`) |
| ES256 | ECDSA | ECDSA 使用 P-256 和 SHA-256 | `GenerateES256Token` | `ParseES256Token` |
| ES384 | ECDSA | ECDSA 使用 P-384 和 SHA-384 | `GenerateToken` (使用 `ES384`) | `ParseToken` (使用 `ES384`) |
| ES512 | ECDSA | ECDSA 使用 P-521 和 SHA-512 | `GenerateToken` (使用 `ES512`) | `ParseToken` (使用 `ES512`) |

## Claims 字段说明

### 标准字段

| 字段 | 全称 | 类型 | 必需 | 描述 |
|------|------|------|------|------|
| `iss` | Issuer | string | 否 | 签发者 |
| `sub` | Subject | string | 否 | 主题（通常是用户 ID） |
| `aud` | Audience | string 或 []string | 否 | 受众 |
| `exp` | Expiration | int64 | 否 | 过期时间（Unix 时间戳） |
| `nbf` | Not Before | int64 | 否 | 生效时间（Unix 时间戳） |
| `iat` | Issued At | int64 | 否 | 签发时间（Unix 时间戳） |
| `jti` | JWT ID | string | 否 | JWT 唯一标识符 |

### 自定义字段

除了标准字段外，你可以添加任意自定义字段到 `MapClaims` 中。

```go
claims := jwt.MapClaims{
    // 标准字段
    "iss": "my-app",
    "sub": "user123",
    "exp": float64(time.Now().Add(24 * time.Hour).Unix()),

    // 自定义字段
    "user_id": float64(12345),
    "email": "user@example.com",
    "role": "admin",
    "permissions": []string{"read", "write", "delete"},
    "metadata": map[string]interface{}{
        "department": "engineering",
        "location": "beijing",
    },
}
```

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

// 或从环境变量读取
secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
if len(secretKey) < 32 {
    panic("JWT secret key too short")
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

### 3. 错误处理

```go
func parseAndValidateToken(tokenString string, secretKey []byte) (jwt.MapClaims, error) {
    // 解析 Token
    claims, err := jwt.JWT.ParseHS256Token(tokenString, secretKey)
    if err != nil {
        return nil, fmt.Errorf("token 解析失败: %w", err)
    }

    // 验证 Token
    err = jwt.JWT.ValidateClaims(claims)
    if err != nil {
        return nil, fmt.Errorf("token 验证失败: %w", err)
    }

    return claims, nil
}
```

### 4. HTTP 中间件示例

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

        // 将 claims 存储到请求上下文
        // ctx := context.WithValue(r.Context(), "claims", claims)
        // next.ServeHTTP(w, r.WithContext(ctx))
    }
}
```

## 常见问题

### Q: 如何选择合适的算法？

- **HS256**：适用于单体应用或服务端生成的场景，性能好，实现简单
- **RS256**：适用于分布式系统，私钥签名、公钥验证，安全性高
- **ES256**：适用于对 Token 大小敏感的场景，签名更短

### Q: Token 过期时间应该设置多久？

根据应用场景和安全要求：
- 短期 Token：15-30 分钟（适用于敏感操作）
- 中期 Token：1-24 小时（适用于一般用户会话）
- 长期 Token：7-30 天（适用于"记住我"功能）

建议使用刷新 Token 机制，短 Token 有效期 + 长 Token 有效期。

### Q: 如何处理 Token 过期？

实现刷新 Token 机制：
1. 访问 Token (Access Token)：短期有效（15-30 分钟）
2. 刷新 Token (Refresh Token)：长期有效（7-30 天）
3. Access Token 过期时使用 Refresh Token 获取新的 Access Token

### Q: MapClaims 中的数字为什么是 float64？

由于 JSON 标准不区分整数和浮点数，`encoding/json` 在解析时会将所有数字解析为 `float64`。使用时需要类型转换：

```go
userId := int(claims["user_id"].(float64))
price := claims["price"].(float64)
```

### Q: 如何在生产环境中保护密钥？

- 使用环境变量或密钥管理服务（如 AWS Secrets Manager、HashiCorp Vault）
- 密钥应足够长（HMAC 至少 32 字节，RSA 至少 2048 位）
- 定期轮换密钥
- 不要将密钥硬编码在代码中
- 不要将密钥提交到版本控制系统

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.0.0
- 初始版本
- 支持 HMAC、RSA、ECDSA 签名算法
- 提供 StandardClaims 和 MapClaims 两种数据结构
- 完整的验证机制和便捷方法
