# Request 请求处理

提供 HTTP 请求绑定和验证功能的工具包，支持 Gin 框架。

## 特性

- **泛型请求绑定** - 使用 BindRequest[T] 快速绑定和验证任意类型的请求参数
- **可插拔验证器** - 通过 ValidatorInterface 接口支持自定义验证逻辑
- **全局默认验证器** - 支持设置包级默认验证器，简化使用
- **类型安全** - 基于泛型的类型安全绑定，避免运行时类型错误

## 快速开始

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/util/request"
)

type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

func main() {
    r := gin.Default()

    r.POST("/users", func(ctx *gin.Context) {
        req, err := request.BindRequest[CreateUserRequest](ctx)
        if err != nil {
            ctx.JSON(400, gin.H{"error": err.Error()})
            return
        }

        ctx.JSON(200, gin.H{"message": "success", "name": req.Name})
    })

    r.Run(":8080")
}
```

## 请求绑定

### BindRequest 泛型绑定

`BindRequest[T]` 是一个泛型辅助函数，用于在 Gin Handler 中快速绑定和验证请求参数。

```go
func BindRequest[T any](ctx *gin.Context) (*T, error)
```

**使用示例：**

```go
type UpdateArticleRequest struct {
    Title   string `json:"title" binding:"required,min=3,max=100"`
    Content string `json:"content" binding:"required"`
    Tags    []string `json:"tags"`
}

func UpdateArticleHandler(ctx *gin.Context) {
    req, err := request.BindRequest[UpdateArticleRequest](ctx)
    if err != nil {
        ctx.Error(apperrors.BadRequest("invalid request").Wrap(err))
        return
    }

    ctx.JSON(200, gin.H{"title": req.Title})
}
```

**注意事项：**
- 必须先调用 `SetDefaultValidator` 设置验证器，否则会返回错误
- 泛型类型 `T` 可以是任意结构体类型
- 验证失败时会返回错误，由调用者决定如何处理

## 验证器

### ValidatorInterface 接口

通过实现 `ValidatorInterface` 接口，可以自定义验证逻辑。

```go
type ValidatorInterface interface {
    Validate(ctx *gin.Context, obj interface{}) error
}
```

### Gin JSON 验证器

```go
type GinValidator struct{}

func (v *GinValidator) Validate(ctx *gin.Context, obj interface{}) error {
    return ctx.ShouldBindJSON(obj)
}
```

### Gin XML 验证器

```go
type XMLValidator struct{}

func (v *XMLValidator) Validate(ctx *gin.Context, obj interface{}) error {
    return ctx.ShouldBindXML(obj)
}
```

### 混合验证器

```go
type HybridValidator struct{}

func (v *HybridValidator) Validate(ctx *gin.Context, obj interface{}) error {
    contentType := ctx.GetHeader("Content-Type")

    switch {
    case strings.Contains(contentType, "application/json"):
        return ctx.ShouldBindJSON(obj)
    case strings.Contains(contentType, "application/xml"):
        return ctx.ShouldBindXML(obj)
    case strings.Contains(contentType, "application/x-www-form-urlencoded"):
        return ctx.ShouldBind(obj)
    default:
        return errors.New("unsupported content type")
    }
}
```

### Query 参数验证器

```go
type QueryValidator struct{}

func (v *QueryValidator) Validate(ctx *gin.Context, obj interface{}) error {
    return ctx.ShouldBindQuery(obj)
}
```

## 验证器管理

### SetDefaultValidator 设置默认验证器

```go
func SetDefaultValidator(validator ValidatorInterface)
```

通常在应用启动时调用一次，设置全局验证器实例。

```go
func init() {
    request.SetDefaultValidator(&GinValidator{})
}
```

### GetDefaultValidator 获取默认验证器

```go
func GetDefaultValidator() ValidatorInterface
```

获取当前设置的默认验证器，可用于测试或特殊场景。

```go
func TestMyHandler(t *testing.T) {
    originalValidator := request.GetDefaultValidator()

    mockValidator := &MockValidator{}
    request.SetDefaultValidator(mockValidator)
    defer request.SetDefaultValidator(originalValidator)

    // 运行测试...
}
```

## 错误处理

`BindRequest` 返回的错误通常来自验证器，应用层可以根据错误类型返回不同的 HTTP 响应。

```go
func CreateArticleHandler(ctx *gin.Context) {
    req, err := request.BindRequest[CreateArticleRequest](ctx)
    if err != nil {
        if e, ok := err.(apperrors.IAppError); ok {
            ctx.Error(e)
            return
        }

        ctx.Error(apperrors.BadRequest("invalid request").Wrap(err))
        return
    }

    ctx.JSON(200, gin.H{"success": true})
}
```

## 最佳实践

### 1. 应用层统一设置验证器

在应用启动时设置默认验证器，避免在业务代码中重复设置。

```go
func main() {
    request.SetDefaultValidator(&GinValidator{})
    
    r := gin.Default()
    r.Run(":8080")
}
```

### 2. 使用结构体标签进行字段级验证

利用 Gin 的 binding 标签进行字段级验证。

```go
type UserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=50"`
    Age      int    `json:"age" binding:"gte=0,lte=150"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}
```

### 3. 统一错误处理

使用应用层的错误包装机制，统一处理验证错误。

### 4. 测试时使用 Mock 验证器

在单元测试中使用 mock 验证器，避免实际绑定操作。

```go
type MockValidator struct {
    ValidateFunc func(ctx *gin.Context, obj interface{}) error
}

func (m *MockValidator) Validate(ctx *gin.Context, obj interface{}) error {
    if m.ValidateFunc != nil {
        return m.ValidateFunc(ctx, obj)
    }
    return nil
}
```

## API

### 接口

- **ValidatorInterface** - 验证器接口
  - `Validate(ctx *gin.Context, obj interface{}) error` - 验证请求

### 函数

- **BindRequest[T any](ctx *gin.Context) (*T, error)** - 绑定并验证请求
- **SetDefaultValidator(validator ValidatorInterface)** - 设置默认验证器
- **GetDefaultValidator() ValidatorInterface** - 获取默认验证器
