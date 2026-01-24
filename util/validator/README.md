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
	Age      int    `json:"age" validate:"required,gte=18,lte=120"`
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
		c.JSON(http.StatusOK, gin.H{"message": "success", "data": req})
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
	c.JSON(http.StatusOK, gin.H{"data": req})
})
```

## 数据验证

### 常用验证标签

```go
type ProductRequest struct {
	Name  string `json:"name" validate:"required"`

	SKU   string `json:"sku" validate:"min=3,max=20"`

	Email string `json:"email" validate:"email"`

	Price float64 `json:"price" validate:"gt=0"`
	Qty   int     `json:"qty" validate:"gte=1,lte=100"`

	Phone string `json:"phone,omitempty" validate:"omitempty,len=11"`

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

type TagsRequest struct {
	Tags []string `json:"tags" validate:"required,min=1,dive,min=2,max=20"`
}
```

### Map 验证

```go
type MetadataRequest struct {
	Metadata map[string]string `json:"metadata" validate:"required,dive,required,min=1"`
}
```

## 自定义验证器

### 注册自定义验证函数

```go
v := validator.NewDefaultValidator()

err := v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
})
if err != nil {
	panic(err)
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,username,min=4,max=20"`
}
```

### 注册多个自定义验证器

```go
v := validator.NewDefaultValidator()

v.RegisterValidation("even", func(fl validator.FieldLevel) bool {
	return fl.Field().Int()%2 == 0
})

v.RegisterValidation("positive", func(fl validator.FieldLevel) bool {
	return fl.Field().Int() > 0
})

type NumberRequest struct {
	Number int `json:"number" validate:"required,even,positive"`
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
fmt.Println(requirements)
```

输出：
```
Password must contain: at least 12 characters, uppercase letter, lowercase letter, number and special character
```

### 服务层密码验证

```go
func CreateUserService(email, password string) error {
	if err := validator.ValidatePassword(password, validator.DefaultPasswordConfig()); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	return nil
}
```

## API

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

func (ve *ValidationError) Error() string
```

#### PasswordConfig

```go
type PasswordConfig struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}
```

### 核心函数

#### NewDefaultValidator

```go
func NewDefaultValidator() *DefaultValidator
```

创建默认验证器实例，自动注册 JSON 标签作为字段名。

#### Validate

```go
func (v *DefaultValidator) Validate(ctx *gin.Context, obj interface{}) error
```

验证 Gin 请求，自动绑定 JSON 并验证结构体。

#### RegisterValidation

```go
func (v *DefaultValidator) RegisterValidation(tag string, fn validator.Func) error
```

注册自定义验证函数。

#### BindAndValidate

```go
func BindAndValidate[T any](ctx *gin.Context, validator Validator) (*T, error)
```

泛型辅助函数，绑定并验证请求。

### 密码验证相关

#### DefaultPasswordConfig

```go
func DefaultPasswordConfig() *PasswordConfig
```

返回默认密码配置（最小长度 12，要求大小写字母、数字和特殊字符）。

#### ValidatePassword

```go
func ValidatePassword(password string, config *PasswordConfig) error
```

使用指定配置验证密码复杂度。

#### ValidateComplexPassword

```go
func ValidateComplexPassword(fl validator.FieldLevel) bool
```

使用默认配置验证密码，用于结构体验证。

#### RegisterPasswordValidation

```go
func RegisterPasswordValidation(v *DefaultValidator) error
```

注册密码复杂度验证器，使用默认配置。

#### RegisterPasswordValidationWithConfig

```go
func RegisterPasswordValidationWithConfig(v *DefaultValidator, config *PasswordConfig) error
```

使用自定义配置注册密码验证器。

#### GetPasswordRequirements

```go
func GetPasswordRequirements() string
```

获取默认密码要求说明。

#### GetPasswordRequirementsWithConfig

```go
func GetPasswordRequirementsWithConfig(config *PasswordConfig) string
```

使用自定义配置获取密码要求说明。

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

### 条件验证

| 标签 | 说明 | 示例 |
|------|------|------|
| `oneof=a b c` | 必须是指定值之一 | `validate:"oneof=active inactive"` |
| `required_unless=Field=value` | 除非指定字段等于值，否则必填 | `validate:"required_unless=Type=guest"` |
| `required_with=Field` | 当指定字段存在时必填 | `validate:"required_with=Phone"` |
| `required_without=Field` | 当指定字段不存在时必填 | `validate:"required_without=Email"` |
| `omitempty` | 为空时跳过验证 | `validate:"omitempty,email"` |
| `dive` | 对数组/切片/map 的元素进行验证 | `validate:"dive,required"` |

### 字符串格式验证

| 标签 | 说明 | 示例 |
|------|------|------|
| `ascii` | 只包含 ASCII 字符 | `validate:"ascii"` |
| `lowercase` | 只包含小写字母 | `validate:"lowercase"` |
| `uppercase` | 只包含大写字母 | `validate:"uppercase"` |
| `e164` | 有效的 E.164 电话号码 | `validate:"e164"` |
| `uuid` | 有效的 UUID | `validate:"uuid"` |
| `uuid3` | 有效的 UUID v3 | `validate:"uuid3"` |
| `uuid4` | 有效的 UUID v4 | `validate:"uuid4"` |
| `uuid5` | 有效的 UUID v5 | `validate:"uuid5"` |

### 时间验证

| 标签 | 说明 | 示例 |
|------|------|------|
| `datetime` | 有效的日期时间格式 | `validate:"datetime=2006-01-02"` |
| `min=now()` | 不早于当前时间 | `validate:"min=now()"` |
| `max=now()` | 不晚于当前时间 | `validate:"max=now()"` |

## 错误处理

验证失败时返回格式化的错误信息：

```go
type ValidationError struct {
	Message string
	Errors  validator.ValidationErrors
}
```

### 错误示例

```go
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

错误消息示例：
```json
{"error": "email is required"}
{"error": "name must be at least 2 characters; email must be a valid email"}
{"error": "password must contain: at least 12 characters, uppercase letter, lowercase letter, number and special character"}
```

### 错误信息格式说明

- `field is required` - 必填字段为空
- `field must be at least n characters` - 长度小于最小值
- `field must be at most n characters` - 长度大于最大值
- `field must be a valid email` - 邮箱格式无效
- `field must contain: at least 12 characters, uppercase, lowercase, number and special character` - 密码复杂度不满足要求
- `field validation failed on tag` - 其他验证失败

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
type RegisterRequest struct {
	Username string `json:"username" validate:"required,alphanum,min=4,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,complexPassword"`
}

type UpdateUserRequest struct {
	Name     string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,len=11"`
	Password string `json:"password,omitempty" validate:"omitempty,min=8"`
}
```

### 使用 omitempty 处理可选字段

```go
type UpdateRequest struct {
	Name  string `json:"name,omitempty" validate:"omitempty,min=2"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}
```

### 嵌套验证的完整示例

```go
type Address struct {
	Street   string `json:"street" validate:"required"`
	City     string `json:"city" validate:"required"`
	Province string `json:"province" validate:"required,len=2"`
	ZipCode  string `json:"zip_code" validate:"required,len=6"`
}

type Contact struct {
	Phone  string `json:"phone" validate:"omitempty,len=11"`
	Email  string `json:"email" validate:"omitempty,email"`
	WeChat string `json:"wechat" validate:"omitempty,min=6,max=30"`
}

type CreateUserRequest struct {
	Name    string  `json:"name" validate:"required,min=2,max=50"`
	Email   string  `json:"email" validate:"required,email"`
	Address Address `json:"address" validate:"required,dive"`
	Contact Contact `json:"contact,omitempty" validate:"omitempty,dive"`
	Tags    []string `json:"tags" validate:"omitempty,min=1,dive,min=2,max=20"`
}
```

### 错误响应格式化

```go
func HandleValidationError(c *gin.Context, err error) {
	if ve, ok := err.(*validator.ValidationError); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "请求参数验证失败",
			"errors":  strings.Split(ve.Message, "; "),
		})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    "BAD_REQUEST",
		"message": err.Error(),
	})
}
```

### 复杂业务场景的验证

```go
type CreateOrderRequest struct {
	UserID      uint64  `json:"user_id" validate:"required,gt=0"`
	Products    []Product `json:"products" validate:"required,min=1,dive"`
	TotalAmount float64 `json:"total_amount" validate:"required,gt=0"`
	DeliveryAddress Address `json:"delivery_address" validate:"required"`
	Remark      string  `json:"remark,omitempty" validate:"omitempty,max=500"`
}

type Product struct {
	ProductID uint64 `json:"product_id" validate:"required,gt=0"`
	Quantity  int    `json:"quantity" validate:"required,gte=1,lte=9999"`
	Price     float64 `json:"price" validate:"required,gt=0"`
}
```

## 依赖

- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) - Web 框架
- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator) - 验证库
