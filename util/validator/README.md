# Validator

基于 Gin 框架的数据验证器，使用 go-playground/validator 提供高效的结构体验证功能。

## 特性

- **结构体验证** - 基于结构体标签的声明式验证，简洁直观
- **自定义验证器** - 支持注册自定义验证函数，灵活扩展验证规则
- **自动错误格式化** - 友好的错误提示信息，使用 JSON 标签作为字段名
- **密码复杂度验证** - 内置密码强度验证器，可自定义复杂度要求
- **泛型支持** - 提供泛型辅助函数，简化绑定和验证流程
- **Gin 集成** - 无缝集成 Gin 框架，自动处理 JSON 绑定

## 快速开始

### 基本使用

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/util/validator"
)

// 定义请求结构体
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

func main() {
    r := gin.Default()

    // 创建验证器实例
    v := validator.NewDefaultValidator()

    r.POST("/login", func(c *gin.Context) {
        var req LoginRequest

        // 验证请求
        if err := v.Validate(c, &req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": err.Error(),
            })
            return
        }

        // 处理登录逻辑
        c.JSON(http.StatusOK, gin.H{
            "message": "Login successful",
            "email":   req.Email,
        })
    })

    r.Run(":8080")
}
```

### 使用泛型辅助函数

```go
r.POST("/register", func(c *gin.Context) {
    // 使用泛型函数一步完成绑定和验证
    req, err := validator.BindAndValidate[RegisterRequest](c, v)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // req 已经是 *RegisterRequest 类型
    c.JSON(http.StatusOK, gin.H{
        "message": "Registration successful",
        "email":   req.Email,
    })
})
```

## 功能详解

### 数据验证

使用 `Validate` 方法验证请求数据：

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=100"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
}

func CreateUser(c *gin.Context) {
    v := validator.NewDefaultValidator()
    var req CreateUserRequest

    if err := v.Validate(c, &req); err != nil {
        // 错误信息会自动格式化，例如：
        // "name is required; email must be a valid email; password must be at least 8 characters"
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 处理业务逻辑
}
```

### 常用验证标签

```go
type ProductRequest struct {
    // 必填字段
    Name  string `json:"name" validate:"required"`

    // 字符串长度验证
    SKU   string `json:"sku" validate:"min=3,max=20"`

    // 邮箱验证
    Email string `json:"email" validate:"email"`

    // 数值范围验证
    Price float64 `json:"price" validate:"gt=0"`
    Qty   int     `json:"qty" validate:"gte=1,lte=100"`

    // 可选字段（为空时跳过验证）
    Phone string `json:"phone,omitempty" validate:"omitempty,len=11"`

    // 枚举验证
    Status string `json:"status" validate:"oneof=active inactive pending"`
}
```

### 嵌套结构体验证

```go
type Address struct {
    Street  string `json:"street" validate:"required"`
    City    string `json:"city" validate:"required"`
    ZipCode string `json:"zip_code" validate:"required,len=6"`
}

type CreateUserRequest struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address" validate:"required"` // 嵌套验证
}

func CreateUser(c *gin.Context) {
    v := validator.NewDefaultValidator()
    var req CreateUserRequest

    if err := v.Validate(c, &req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
}
```

### 切片和数组验证

```go
type BatchDeleteRequest struct {
    IDs []int `json:"ids" validate:"required,min=1,dive,gte=1"`
}

// dive 标签表示对切片中的每个元素进行验证
func BatchDelete(c *gin.Context) {
    v := validator.NewDefaultValidator()
    var req BatchDeleteRequest

    if err := v.Validate(c, &req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
}
```

### 自定义验证器

注册自定义验证函数：

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    "github.com/lite-lake/litecore-go/util/validator"
)

// 自定义验证函数 - 验证用户名只包含小写字母和数字
func validateUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    for _, c := range username {
        if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
            return false
        }
    }
    return true
}

func main() {
    r := gin.Default()

    // 创建验证器
    v := validator.NewDefaultValidator()

    // 注册自定义验证器
    v.RegisterValidation("username", validateUsername)

    type RegisterRequest struct {
        Username string `json:"username" validate:"required,username,min=4,max=20"`
        Email    string `json:"email" validate:"required,email"`
    }

    r.POST("/register", func(c *gin.Context) {
        var req RegisterRequest
        if err := v.Validate(c, &req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"message": "success"})
    })

    r.Run(":8080")
}
```

### 密码复杂度验证

使用内置的密码验证器：

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/util/validator"
)

func main() {
    r := gin.Default()

    // 创建验证器
    v := validator.NewDefaultValidator()

    // 注册默认的密码复杂度验证器
    // 要求：至少12位，包含大小写字母、数字和特殊字符
    validator.RegisterPasswordValidation(v)

    type RegisterRequest struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,complexPassword"`
    }

    r.POST("/register", func(c *gin.Context) {
        var req RegisterRequest
        if err := v.Validate(c, &req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": err.Error(),
                // 如果密码不符合要求，错误信息示例：
                // "password must contain: at least 12 characters, uppercase letter,
                //  lowercase letter, number, special character"
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
    })

    r.Run(":8080")
}
```

#### 自定义密码复杂度要求

```go
// 创建自定义密码配置
customConfig := &validator.PasswordConfig{
    MinLength:      8,   // 最小8位
    MaxLength:      64,  // 最大64位
    RequireUpper:   true, // 要求大写
    RequireLower:   true, // 要求小写
    RequireNumber:  true, // 要求数字
    RequireSpecial: false, // 不要求特殊字符
}

// 使用自定义配置注册验证器
validator.RegisterPasswordValidationWithConfig(v, customConfig)

type RegisterRequest struct {
    Password string `json:"password" validate:"required,complexPassword"`
}
```

#### 获取密码要求说明

```go
// 获取默认密码要求
requirements := validator.GetPasswordRequirements()
// 返回: "Password must contain: at least 12 characters, uppercase letter,
//        lowercase letter, number, special character"

// 或使用自定义配置
requirements := validator.GetPasswordRequirementsWithConfig(customConfig)

// 在 API 响应中返回密码要求
c.JSON(200, gin.H{
    "password_requirements": requirements,
})
```

#### 服务层密码验证

```go
import "github.com/lite-lake/litecore-go/util/validator"

func CreateUserService(email, password string) error {
    // 在服务层验证密码复杂度
    if err := validator.ValidatePassword(password, validator.DefaultPasswordConfig()); err != nil {
        return fmt.Errorf("invalid password: %w", err)
    }

    // 创建用户...
    return nil
}
```

### 错误处理

验证器会返回友好的错误信息：

```go
type ValidationError struct {
    Message string
    Errors  validator.ValidationErrors
}

func handleRequest(c *gin.Context) {
    v := validator.NewDefaultValidator()
    var req CreateUserRequest

    if err := v.Validate(c, &req); err != nil {
        // 类型断言获取详细错误信息
        if ve, ok := err.(*validator.ValidationError); ok {
            c.JSON(400, gin.H{
                "error":   ve.Message,
                "details": ve.Errors, // 原始验证错误
            })
            return
        }

        // 其他错误（如 JSON 解析错误）
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 处理成功...
}
```

错误信息示例：

```json
// 单个字段错误
{
  "error": "email is required"
}

// 多个字段错误
{
  "error": "name must be at least 2 characters; email must be a valid email; password must be at least 8 characters"
}

// 密码复杂度错误
{
  "error": "password must contain: at least 12 characters, uppercase letter, lowercase letter, number and special character"
}
```

## API 参考

### 接口和类型

#### Validator 接口

```go
type Validator interface {
    Validate(ctx *gin.Context, obj interface{}) error
}
```

#### DefaultValidator

默认验证器实现。

```go
type DefaultValidator struct {
    engine *validator.Validate
}
```

### 核心函数

#### NewDefaultValidator

创建默认验证器实例。

```go
func NewDefaultValidator() *DefaultValidator
```

**示例：**

```go
v := validator.NewDefaultValidator()
```

#### Validate

验证并绑定请求数据到结构体。

```go
func (v *DefaultValidator) Validate(ctx *gin.Context, obj interface{}) error
```

**参数：**
- `ctx` - Gin 上下文
- `obj` - 指向结构体的指针

**返回：**
- `error` - 验证错误，成功时返回 nil

**示例：**

```go
var req LoginRequest
if err := v.Validate(c, &req); err != nil {
    return err
}
```

#### RegisterValidation

注册自定义验证函数。

```go
func (v *DefaultValidator) RegisterValidation(tag string, fn validator.Func) error
```

**参数：**
- `tag` - 验证标签名（如 "complexPassword"）
- `fn` - 验证函数，签名为 `func(fl validator.FieldLevel) bool`

**返回：**
- `error` - 注册错误

**示例：**

```go
err := v.RegisterValidation("even", func(fl validator.FieldLevel) bool {
    value := fl.Field().Int()
    return value%2 == 0
})
```

#### BindAndValidate

泛型辅助函数，一步完成绑定和验证。

```go
func BindAndValidate[T any](ctx *gin.Context, validator Validator) (*T, error)
```

**类型参数：**
- `T` - 请求结构体类型

**参数：**
- `ctx` - Gin 上下文
- `validator` - 验证器实例

**返回：**
- `*T` - 验证后的请求结构体指针
- `error` - 验证错误

**示例：**

```go
req, err := validator.BindAndValidate[CreateUserRequest](c, v)
if err != nil {
    return err
}
// req 是 *CreateUserRequest 类型
```

### 密码验证相关

#### PasswordConfig

密码验证配置结构体。

```go
type PasswordConfig struct {
    MinLength      int  // 密码最小长度（默认 12）
    MaxLength      int  // 密码最大长度（默认 128）
    RequireUpper   bool // 是否要求大写字母（默认 true）
    RequireLower   bool // 是否要求小写字母（默认 true）
    RequireNumber  bool // 是否要求数字（默认 true）
    RequireSpecial bool // 是否要求特殊字符（默认 true）
}
```

#### DefaultPasswordConfig

返回默认密码配置。

```go
func DefaultPasswordConfig() *PasswordConfig
```

#### ValidatePassword

使用指定配置验证密码复杂度。

```go
func ValidatePassword(password string, config *PasswordConfig) error
```

**参数：**
- `password`: 待验证的密码
- `config`: 密码配置，为 nil 时使用默认配置

**示例：**

```go
err := validator.ValidatePassword("MyPassword123!", validator.DefaultPasswordConfig())
```

#### RegisterPasswordValidation

使用默认配置注册密码复杂度验证器。

```go
func RegisterPasswordValidation(v *DefaultValidator) error
```

#### RegisterPasswordValidationWithConfig

使用自定义配置注册密码验证器。

```go
func RegisterPasswordValidationWithConfig(v *DefaultValidator, config *PasswordConfig) error
```

#### GetPasswordRequirements

获取默认密码要求说明。

```go
func GetPasswordRequirements() string
```

**返回值示例：**

```
"Password must contain: at least 12 characters, uppercase letter, lowercase letter, number, special character"
```

#### GetPasswordRequirementsWithConfig

使用自定义配置获取密码要求说明。

```go
func GetPasswordRequirementsWithConfig(config *PasswordConfig) string
```

## 常用验证标签

### 字符串验证

| 标签 | 说明 | 示例 |
|------|------|------|
| `required` | 必填字段 | `validate:"required"` |
| `min=n` | 最小长度为 n | `validate:"min=8"` |
| `max=n` | 最大长度为 n | `validate:"max=100"` |
| `len=n` | 长度必须为 n | `validate:"len=11"` |
| `email` | 有效的邮箱地址 | `validate:"email"` |
| `url` | 有效的 URL | `validate:"url"` |
| `startswith=abc` | 以指定前缀开头 | `validate:"startswith=user_"` |
| `endswith=xyz` | 以指定后缀结尾 | `validate:"endswith=@example.com"` |

### 数值验证

| 标签 | 说明 | 示例 |
|------|------|------|
| `gt=n` | 大于 n | `validate:"gt=0"` |
| `gte=n` | 大于等于 n | `validate:"gte=18"` |
| `lt=n` | 小于 n | `validate:"lt=100"` |
| `lte=n` | 小于等于 n | `validate:"lte=120"` |
| `eq=n` | 等于 n | `validate:"eq=10"` |
| `ne=n` | 不等于 n | `validate:"ne=0"` |

### 其他验证

| 标签 | 说明 | 示例 |
|------|------|------|
| `oneof=a b c` | 必须是指定值之一 | `validate:"oneof=active inactive"` |
| `required_unless=Field=value` | 除非指定字段等于值，否则必填 | `validate:"required_unless=Type=guest"` |
| `required_with=Field` | 当指定字段存在时必填 | `validate:"required_with=Phone"` |
| `required_without=Field` | 当指定字段不存在时必填 | `validate:"required_without=Email"` |
| `omitempty` | 为空时跳过验证 | `validate:"omitempty,email"` |
| `dive` | 对数组/切片/map 的元素进行验证 | `validate:"dive,required"` |

## 完整示例

### 用户注册 API

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/util/validator"
)

type RegisterRequest struct {
    Username string `json:"username" validate:"required,alphanum,min=4,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,complexPassword"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
}

type Address struct {
    Province string `json:"province" validate:"required"`
    City     string `json:"city" validate:"required"`
    Detail   string `json:"detail" validate:"required,max=200"`
}

type CompleteProfileRequest struct {
    UserID  int64   `json:"user_id" validate:"required"`
    Address Address `json:"address" validate:"required"`
}

func main() {
    r := gin.Default()

    // 初始化验证器
    v := validator.NewDefaultValidator()

    // 注册密码复杂度验证器
    if err := validator.RegisterPasswordValidation(v); err != nil {
        panic(err)
    }

    // 注册接口
    r.POST("/api/v1/register", func(c *gin.Context) {
        // 使用泛型函数
        req, err := validator.BindAndValidate[RegisterRequest](c, v)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "code":    400,
                "message": "Validation failed",
                "error":   err.Error(),
            })
            return
        }

        // 处理注册逻辑...
        c.JSON(http.StatusOK, gin.H{
            "code":    200,
            "message": "Registration successful",
            "data": gin.H{
                "username": req.Username,
                "email":    req.Email,
            },
        })
    })

    // 完善资料接口
    r.POST("/api/v1/profile/complete", func(c *gin.Context) {
        var req CompleteProfileRequest
        if err := v.Validate(c, &req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "code":    400,
                "message": "Invalid request",
                "error":   err.Error(),
            })
            return
        }

        // 处理业务逻辑...
        c.JSON(http.StatusOK, gin.H{
            "code":    200,
            "message": "Profile completed successfully",
        })
    })

    // 获取密码要求
    r.GET("/api/v1/password-requirements", func(c *gin.Context) {
        requirements := validator.GetPasswordRequirements()
        c.JSON(http.StatusOK, gin.H{
            "code":    200,
            "message": "Success",
            "data": gin.H{
                "requirements": requirements,
            },
        })
    })

    r.Run(":8080")
}
```

### 测试 API

```bash
# 成功注册
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john123",
    "email": "john@example.com",
    "password": "SecureP@ssw0rd123",
    "age": 25
  }'

# 密码不符合要求
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john123",
    "email": "john@example.com",
    "password": "weak",
    "age": 25
  }'

# 响应示例：
# {
#   "code": 400,
#   "message": "Validation failed",
#   "error": "password must contain: at least 12 characters, uppercase letter, lowercase letter, number and special character"
# }
```

## 最佳实践

### 1. 全局初始化验证器

```go
package main

import "github.com/lite-lake/litecore-go/util/validator"

var GlobalValidator *validator.DefaultValidator

func init() {
    GlobalValidator = validator.NewDefaultValidator()

    // 注册所有自定义验证器
    GlobalValidator.RegisterValidation("username", validateUsername)
    validator.RegisterPasswordValidation(GlobalValidator)
    // ...
}
```

### 2. 分离请求结构体

```go
// requests/user.go
package requests

type RegisterRequest struct {
    Username string `json:"username" validate:"required,alphanum,min=4,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,complexPassword"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}
```

### 3. 统一错误处理

```go
func HandleValidationError(c *gin.Context, err error) {
    if ve, ok := err.(*validator.ValidationError); ok {
        c.JSON(400, gin.H{
            "code":    400001,
            "message": "Validation failed",
            "details": ve.Message,
        })
        return
    }

    c.JSON(400, gin.H{
        "code":    400000,
        "message": "Invalid request",
        "error":   err.Error(),
    })
}
```

### 4. 提供友好的错误提示

```go
// 在 API 文档中返回验证规则
c.JSON(200, gin.H{
    "validation_rules": gin.H{
        "username": "必填，4-20位字母数字",
        "email": "必填，有效的邮箱地址",
        "password": validator.GetPasswordRequirements(),
    },
})
```

## 常见问题

### Q: 如何忽略某些字段的验证？

使用 `omitempty` 标签：

```go
type UpdateUserRequest struct {
    Name  string `json:"name,omitempty" validate:"omitempty,max=50"`
    Email string `json:"email,omitempty" validate:"omitempty,email"`
}
```

### Q: 如何验证多个条件？

使用逗号分隔多个标签：

```go
Password string `json:"password" validate:"required,min=12,max=100,complexPassword"`
```

### Q: 如何自定义错误消息？

可以在应用层进行转换：

```go
func formatError(err error) map[string]string {
    messages := map[string]string{
        "required": "%s 不能为空",
        "email":    "%s 必须是有效的邮箱地址",
        "min":      "%s 长度不能少于 %s 个字符",
    }

    // 自定义格式化逻辑...
    return formattedMessages
}
```

### Q: 验证器是线程安全的吗？

是的，`DefaultValidator` 创建后可以在多个 goroutine 中安全使用。

## 依赖

- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) - Web 框架
- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator) - 验证库

## License

MIT License
