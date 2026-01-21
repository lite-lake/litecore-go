# Request 请求处理

提供 HTTP 请求绑定和验证功能的工具包。

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

// 1. 定义请求类型
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

// 2. 实现验证器（通常应用层会提供标准实现）
type GinValidator struct{}

func (v *GinValidator) Validate(ctx *gin.Context, obj interface{}) error {
    return ctx.ShouldBindJSON(obj)
}

func main() {
    // 3. 设置默认验证器（应用启动时）
    request.SetDefaultValidator(&GinValidator{})

    // 4. 在 Handler 中使用
    handler := func(ctx *gin.Context) {
        req, err := request.BindRequest[CreateUserRequest](ctx)
        if err != nil {
            ctx.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // 使用 req...
        ctx.JSON(200, gin.H{"message": "success", "name": req.Name})
    }

    r := gin.Default()
    r.POST("/users", handler)
    r.Run(":8080")
}
```

## BindRequest 绑定请求

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

    // 使用 req 进行业务处理...
}
```

**注意事项：**
- 必须先调用 `SetDefaultValidator` 设置验证器，否则会 panic
- 泛型类型 `T` 必须是指针类型或可寻址的值类型
- 验证失败时会返回错误，由调用者决定如何处理

## 自定义验证器

通过实现 `ValidatorInterface` 接口，可以自定义验证逻辑。

### ValidatorInterface 接口

```go
type ValidatorInterface interface {
    Validate(ctx *gin.Context, obj interface{}) error
}
```

### Gin 验证器实现

```go
type GinValidator struct{}

func (v *GinValidator) Validate(ctx *gin.Context, obj interface{}) error {
    return ctx.ShouldBindJSON(obj)
}
```

### XML 验证器实现

```go
type XMLValidator struct{}

func (v *XMLValidator) Validate(ctx *gin.Context, obj interface{}) error {
    return ctx.ShouldBindXML(obj)
}
```

### 混合验证器实现

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

## 管理默认验证器

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
    // 使用 mock 验证器进行测试
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
        // 判断错误类型
        if e, ok := err.(apperrors.IAppError); ok {
            ctx.Error(e)
            return
        }

        // 转换为应用错误
        ctx.Error(apperrors.BadRequest("invalid request").Wrap(err))
        return
    }

    // 处理请求...
}
```

## 最佳实践

1. **应用层设置验证器** - 在应用启动时设置默认验证器，避免在业务代码中重复设置

2. **统一错误处理** - 使用应用层的错误包装机制，统一处理验证错误

3. **使用结构体标签** - 利用 Gin 的 binding 标签进行字段级验证

```go
type UserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=50"`
    Age      int    `json:"age" binding:"gte=0,lte=150"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}
```

4. **测试时替换验证器** - 在单元测试中使用 mock 验证器，避免实际绑定操作

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

### 函数

- **BindRequest[T]** - 绑定并验证请求
- **SetDefaultValidator** - 设置默认验证器
- **GetDefaultValidator** - 获取默认验证器

### 变量

- **defaultValidator** - 默认验证器实例（包级变量）
