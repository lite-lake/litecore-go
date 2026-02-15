package common

// CommonResponse 统一响应结构
type CommonResponse struct {
	Code    int         `json:"code"`           // 响应状态码
	Message string      `json:"message"`        // 响应消息
	Data    interface{} `json:"data,omitempty"` // 响应数据（可选）
}

// SuccessResponse 创建成功响应（带消息和数据）
func SuccessResponse(message string, data interface{}) CommonResponse {
	return CommonResponse{
		Code:    HTTPStatusOK,
		Message: message,
		Data:    data,
	}
}

// SuccessResponseWith 创建成功响应（带自定义状态码、消息和数据）
func SuccessResponseWith(code int, message string, data interface{}) CommonResponse {
	return CommonResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// SuccessWithData 创建成功响应（仅数据）
func SuccessWithData(data interface{}) CommonResponse {
	return CommonResponse{
		Code:    HTTPStatusOK,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage 创建成功响应（仅消息）
func SuccessWithMessage(message string) CommonResponse {
	return CommonResponse{
		Code:    HTTPStatusOK,
		Message: message,
	}
}

// ErrorResponse 创建错误响应
func ErrorResponse(code int, message string) CommonResponse {
	return CommonResponse{
		Code:    code,
		Message: message,
	}
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code     int         `json:"code"`      // 响应状态码
	Message  string      `json:"message"`   // 响应消息
	List     interface{} `json:"list"`      // 数据列表
	Total    int64       `json:"total"`     // 总记录数
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页条数
}

// PagedResponse 创建分页响应
func PagedResponse(list interface{}, total int64, page, pageSize int) PageResponse {
	return PageResponse{
		Code:     HTTPStatusOK,
		Message:  "success",
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// PagedResponseWithMessage 创建分页响应（带自定义消息）
func PagedResponseWithMessage(message string, list interface{}, total int64, page, pageSize int) PageResponse {
	return PageResponse{
		Code:     HTTPStatusOK,
		Message:  message,
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// FieldError 字段级验证错误
type FieldError struct {
	Field   string `json:"field"`   // 字段名
	Message string `json:"message"` // 错误消息
}

// ValidationErrorResponse 验证错误响应结构
type ValidationErrorResponse struct {
	Code    int         `json:"code"`             // 响应状态码
	Message string      `json:"message"`          // 响应消息
	Errors  FieldErrors `json:"errors,omitempty"` // 字段级错误列表
}

// FieldErrors 字段错误列表
type FieldErrors []FieldError

// ValidationResponse 创建验证错误响应
func ValidationResponse(message string, errors FieldErrors) ValidationErrorResponse {
	return ValidationErrorResponse{
		Code:    HTTPStatusBadRequest,
		Message: message,
		Errors:  errors,
	}
}

// ValidationResponseWithCode 创建验证错误响应（带自定义状态码）
func ValidationResponseWithCode(code int, message string, errors FieldErrors) ValidationErrorResponse {
	return ValidationErrorResponse{
		Code:    code,
		Message: message,
		Errors:  errors,
	}
}

// AddFieldError 添加字段错误
func (fe *FieldErrors) AddFieldError(field, message string) {
	*fe = append(*fe, FieldError{Field: field, Message: message})
}

// 常用错误响应
var (
	ErrBadRequest     = ErrorResponse(HTTPStatusBadRequest, "请求参数错误")
	ErrUnauthorized   = ErrorResponse(HTTPStatusUnauthorized, "未授权")
	ErrForbidden      = ErrorResponse(HTTPStatusForbidden, "禁止访问")
	ErrNotFound       = ErrorResponse(HTTPStatusNotFound, "资源不存在")
	ErrInternalServer = ErrorResponse(HTTPStatusInternalServerError, "服务器内部错误")
)
