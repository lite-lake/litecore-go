// Package request 提供 HTTP 请求绑定和验证功能。
//
// 核心特性：
//   - 泛型请求绑定：使用 BindRequest[T] 快速绑定和验证任意类型的请求参数
//   - 可插拔验证器：通过 ValidatorInterface 接口支持自定义验证逻辑
//   - 全局默认验证器：支持设置包级默认验证器，简化使用
//
// 基本用法：
//
//	// 1. 定义请求类型
//	type CreateUserRequest struct {
//	    Name  string `json:"name" binding:"required"`
//	    Email string `json:"email" binding:"required,email"`
//	}
//
//	// 2. 设置默认验证器（通常在应用启动时）
//	validator := NewGinValidator()
//	request.SetDefaultValidator(validator)
//
//	// 3. 在 Controller 中绑定和验证请求
//	req, err := request.BindRequest[CreateUserRequest](ctx)
//	if err != nil {
//	    ctx.Error(apperrors.BadRequest("invalid request").Wrap(err))
//	    return
//	}
//
// 自定义验证器：
//
//	// 实现 ValidatorInterface 接口
//	type MyValidator struct{}
//
//	func (v *MyValidator) Validate(ctx *gin.Context, obj interface{}) error {
//	    // 自定义验证逻辑
//	    return ctx.ShouldBindJSON(obj)
//	}
package request
