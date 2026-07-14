package common

// APIResponse 统一 API 响应结构体
// 所有内部业务接口必须使用此结构体，文件传输和外部行业标准接口除外
type APIResponse struct {
	Code    int         `json:"code"`           // 业务状态码，成功固定为 200
	Message string      `json:"message"`        // 响应消息
	Data    interface{} `json:"data,omitempty"` // 响应数据
}

// APIErrorResponse 统一错误响应结构体（无 Data 字段）
type APIErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// PaginatedData 分页响应数据
type PaginatedData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// BusinessErrorCode 业务错误码类型
// 范围划分：
//
//	200      : 成功
//	400-499  : 客户端错误（参数、认证、权限、资源）
//	500-599  : 服务端错误（内部、上游、超时）
type BusinessErrorCode int

const (
	// CodeSuccess 成功
	CodeSuccess BusinessErrorCode = 200

	// 客户端错误 4xx
	CodeBadRequest      BusinessErrorCode = 400 // 请求参数错误
	CodeUnauthorized    BusinessErrorCode = 401 // 未认证
	CodeForbidden       BusinessErrorCode = 403 // 无权限
	CodeNotFound        BusinessErrorCode = 404 // 资源不存在
	CodeConflict        BusinessErrorCode = 409 // 资源冲突
	CodeUnprocessable   BusinessErrorCode = 422 // 无法处理
	CodeTooManyRequests BusinessErrorCode = 429 // 请求过于频繁

	// 服务端错误 5xx
	CodeInternalError  BusinessErrorCode = 500 // 内部服务器错误
	CodeBadGateway     BusinessErrorCode = 502 // 上游服务错误
	CodeServiceUnavail BusinessErrorCode = 503 // 服务不可用
	CodeGatewayTimeout BusinessErrorCode = 504 // 上游超时
)

// SuccessWithData 返回带数据的成功响应
func SuccessWithData(data interface{}) *APIResponse {
	return &APIResponse{
		Code:    int(CodeSuccess),
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage 返回带自定义消息的成功响应
func SuccessWithMessage(message string, data interface{}) *APIResponse {
	return &APIResponse{
		Code:    int(CodeSuccess),
		Message: message,
		Data:    data,
	}
}

// SuccessOnly 返回无数据的成功响应
func SuccessOnly() *APIResponse {
	return &APIResponse{
		Code:    int(CodeSuccess),
		Message: "success",
	}
}

// SuccessWithPage 返回分页成功响应
func SuccessWithPage(list interface{}, total int64, page, pageSize int) *APIResponse {
	return &APIResponse{
		Code:    int(CodeSuccess),
		Message: "success",
		Data: PaginatedData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	}
}

// ErrorWith 返回业务错误响应
func ErrorWith(code BusinessErrorCode, message string) *APIErrorResponse {
	return &APIErrorResponse{
		Code:    int(code),
		Message: message,
	}
}

// ErrorWithCode 返回自定义数字码错误响应
func ErrorWithCode(code int, message string) *APIErrorResponse {
	return &APIErrorResponse{
		Code:    code,
		Message: message,
	}
}

// BadRequestError 返回参数错误响应
func BadRequestError(message string) *APIErrorResponse {
	return ErrorWith(CodeBadRequest, message)
}

// UnauthorizedError 返回未认证响应
func UnauthorizedError(message string) *APIErrorResponse {
	return ErrorWith(CodeUnauthorized, message)
}

// ForbiddenError 返回无权限响应
func ForbiddenError(message string) *APIErrorResponse {
	return ErrorWith(CodeForbidden, message)
}

// NotFoundError 返回资源不存在响应
func NotFoundError(message string) *APIErrorResponse {
	return ErrorWith(CodeNotFound, message)
}

// InternalError 返回内部错误响应
func InternalError(message string) *APIErrorResponse {
	return ErrorWith(CodeInternalError, message)
}
