// Package validator 基于 gin 框架的数据验证器，使用 go-playground/validator 提供结构体验证
package validator

/*
核心特性：
  - 结构体验证：基于结构体标签进行声明式验证，支持嵌套结构和数组
  - 丰富验证规则：内置 required、email、min、max、gte、lte 等常用验证标签
  - 自定义验证器：支持注册自定义验证函数扩展验证逻辑
  - 友好错误提示：自动格式化验证错误，使用 JSON 标签名称作为字段名
  - 泛型辅助函数：提供 BindAndValidate 泛型函数简化请求绑定和验证流程
  - Gin 集成：与 Gin 框架无缝集成，自动处理 JSON 绑定和验证

基本用法：

	import (
	    "net/http"
	    "github.com/gin-gonic/gin"
	    "yourproject/util/validator"
	)

	// 定义请求结构体
	type CreateUserRequest struct {
	    Name     string `json:"name" validate:"required,min=2,max=50"`
	    Email    string `json:"email" validate:"required,email"`
	    Password string `json:"password" validate:"required,min=8"`
	    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
	}

	// 方式一：使用 Validate 方法
	func CreateUser(c *gin.Context) {
	    var req CreateUserRequest

	    // 创建验证器实例
	    v := validator.NewDefaultValidator()

	    // 验证请求
	    if err := v.Validate(c, &req); err != nil {
	        c.JSON(http.StatusBadRequest, gin.H{
	            "error": err.Error(),
	        })
	        return
	    }

	    // 处理业务逻辑
	    c.JSON(http.StatusOK, gin.H{
	        "message": "User created successfully",
	        "data":    req,
	    })
	}

	// 方式二：使用泛型辅助函数
	func CreateUserWithGeneric(c *gin.Context) {
	    v := validator.NewDefaultValidator()

	    req, err := validator.BindAndValidate[CreateUserRequest](c, v)
	    if err != nil {
	        c.JSON(http.StatusBadRequest, gin.H{
	            "error": err.Error(),
	        })
	        return
	    }

	    c.JSON(http.StatusOK, gin.H{
	        "message": "User created successfully",
	        "data":    req,
	    })
	}

自定义验证器：

	// 注册自定义验证函数
	v := validator.NewDefaultValidator()

	err := v.RegisterValidation("strongPassword", func(fl validator.FieldLevel) bool {
	    password := fl.Field().String()
	    if len(password) < 12 {
	        return false
	    }

	    hasUpper := false
	    hasLower := false
	    hasNumber := false
	    hasSpecial := false

	    for _, c := range password {
	        switch {
	        case c >= 'A' && c <= 'Z':
	            hasUpper = true
	        case c >= 'a' && c <= 'z':
	            hasLower = true
	        case c >= '0' && c <= '9':
	            hasNumber = true
	        case c == '!' || c == '@' || c == '#' || c == '$' || c == '%' || c == '^' || c == '&':
	            hasSpecial = true
	        }
	    }

	    return hasUpper && hasLower && hasNumber && hasSpecial
	})

	if err != nil {
	    panic(err)
	}

	// 使用自定义验证器
	type UserRequest struct {
	    Password string `json:"password" validate:"required,strongPassword"`
	}

常用验证标签：

	required    : 字段必填
	email       : 有效的邮箱格式
	url         : 有效的 URL 格式
	min=n       : 字符串长度或数值最小为 n
	max=n       : 字符串长度或数值最大为 n
	gte=n       : 大于或等于 n
	gt=n        : 大于 n
	lte=n       : 小于或等于 n
	lt=n        : 小于 n
	omitempty   : 字段为空时跳过验证
	alpha       : 只包含字母字符
	alphanum    : 只包含字母和数字字符
	numeric     : 有效的数值字符串

嵌套结构体验证：

	type Address struct {
	    Street string `json:"street" validate:"required"`
	    City   string `json:"city" validate:"required"`
	    Zip    string `json:"zip" validate:"required,len=6"`
	}

	type UserRequest struct {
	    Name    string  `json:"name" validate:"required"`
	    Address Address `json:"address" validate:"required"`
	}

错误处理：

	验证失败时，Validate 方法返回 ValidationError 类型错误：
	- Message: 格式化后的错误消息，包含所有验证错误的描述
	- Errors: 原始的 validator.ValidationErrors，可用于获取详细错误信息

	错误消息示例：
	"name is required; email must be a valid email; password must be at least 8 characters"

性能建议：

	- 验证器实例可以全局复用，避免重复创建
	- 建议在应用启动时初始化 DefaultValidator 并注册所有自定义验证器
	- 复杂验证逻辑应放在自定义验证函数中，保持代码清晰
*/
