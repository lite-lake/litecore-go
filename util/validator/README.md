# Validator (验证器)

基于 Gin 框架的数据验证器，使用 go-playground/validator 提供高效的结构体验证功能。

## 特性

- **结构体验证** - 基于结构体标签的声明式验证，简洁直观
- **友好错误提示** - 自动格式化验证错误，使用 JSON 标签作为字段名
- **自定义验证器** - 支持注册自定义验证函数，灵活扩展验证规则
- **密码复杂度验证** - 内置密码强度验证器，可自定义复杂度要求
- **泛型辅助函数** - 提供 BindAndValidate 泛型函数，简化绑定和验证流程
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

type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

func main() {
    r := gin.Default()
    v := validator.NewDefaultValidator()

    r.POST("/users", func(c *gin.Context) {
        var req CreateUserRequest
        if err := v.Validate(c, &req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "success"})
    })

    r.Run(":8080")
}
```

### 使用泛型辅助函数

```go
r.POST("/users", func(c *gin.Context) {
    req, err := validator.BindAndValidate[CreateUserRequest](c, v)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // req 是 *CreateUserRequest 类型
})
```

## 数据验证

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
    Address Address `json:"address" validate:"required"`
}
```

### 切片和数组验证

```go
type BatchDeleteRequest struct {
    IDs []int `json:"ids" validate:"required,min=1,dive,gte=1"`
}
// dive 标签表示对切片中的每个元素进行验证
```

## 自定义验证器

### 注册自定义验证函数

```go
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

v := validator.NewDefaultValidator()
v.RegisterValidation("username", validateUsername)

type RegisterRequest struct {
    Username string `json:"username" validate:"required,username,min=4,max=20"`
}
```

## 密码复杂度验证

### 使用内置密码验证器

```go
v := validator.NewDefaultValidator()
validator.RegisterPasswordValidation(v)

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,complexPassword"`
}
```

### 自定义密码配置

```go
config := &validator.PasswordConfig{
    MinLength:      8,
    MaxLength:      64,
    RequireUpper:   true,
    RequireLower:   true,
    RequireNumber:  true,
    RequireSpecial: false,
}
validator.RegisterPasswordValidationWithConfig(v, config)
```

### 获取密码要求说明

```go
requirements := validator.GetPasswordRequirements()
// 返回: "Password must contain: at least 12 characters, uppercase letter,
//        lowercase letter, number, special character"
```

### 服务层密码验证

```go
func CreateUserService(email, password string) error {
    if err := validator.ValidatePassword(password, validator.DefaultPasswordConfig()); err != nil {
        return fmt.Errorf("invalid password: %w", err)
    }
    // 创建用户...
    return nil
}
```

## API 参考

### 核心接口和类型

#### Validator 接口

```go
type Validator interface {
    Validate(ctx *gin.Context, obj interface{}) error
}
```

#### DefaultValidator

```go
type DefaultValidator struct {
    engine *validator.Validate
}
```

#### ValidationError

```go
type ValidationError struct {
    Message string
    Errors  validator.ValidationErrors
}
```

### 核心函数

#### NewDefaultValidator

```go
func NewDefaultValidator() *DefaultValidator
```

#### Validate

```go
func (v *DefaultValidator) Validate(ctx *gin.Context, obj interface{}) error
```

#### RegisterValidation

```go
func (v *DefaultValidator) RegisterValidation(tag string, fn validator.Func) error
```

#### BindAndValidate

```go
func BindAndValidate[T any](ctx *gin.Context, validator Validator) (*T, error)
```

### 密码验证相关

#### PasswordConfig

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

```go
func DefaultPasswordConfig() *PasswordConfig
```

#### ValidatePassword

```go
func ValidatePassword(password string, config *PasswordConfig) error
```

#### RegisterPasswordValidation

```go
func RegisterPasswordValidation(v *DefaultValidator) error
```

#### RegisterPasswordValidationWithConfig

```go
func RegisterPasswordValidationWithConfig(v *DefaultValidator, config *PasswordConfig) error
```

#### GetPasswordRequirements

```go
func GetPasswordRequirements() string
```

#### GetPasswordRequirementsWithConfig

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
| `alpha` | 只包含字母 | `validate:"alpha"` |
| `alphanum` | 只包含字母和数字 | `validate:"alphanum"` |
| `numeric` | 有效的数值字符串 | `validate:"numeric"` |

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

## 错误处理

验证失败时返回格式化的错误信息：

```go
type ValidationError struct {
    Message string
    Errors  validator.ValidationErrors
}

if err := v.Validate(c, &req); err != nil {
    if ve, ok := err.(*validator.ValidationError); ok {
        c.JSON(400, gin.H{
            "error":   ve.Message,
            "details": ve.Errors,
        })
        return
    }
}
```

错误信息示例：
```json
{"error": "email is required"}
{"error": "name must be at least 2 characters; email must be a valid email"}
{"error": "password must contain: at least 12 characters, uppercase letter, lowercase letter, number and special character"}
```

## 最佳实践

### 全局初始化验证器

```go
var GlobalValidator *validator.DefaultValidator

func init() {
    GlobalValidator = validator.NewDefaultValidator()
    GlobalValidator.RegisterValidation("username", validateUsername)
    validator.RegisterPasswordValidation(GlobalValidator)
}
```

### 分离请求结构体

```go
// requests/user.go
type RegisterRequest struct {
    Username string `json:"username" validate:"required,alphanum,min=4,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,complexPassword"`
}
```

## 依赖

- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) - Web 框架
- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator) - 验证库
