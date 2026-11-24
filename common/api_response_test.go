package common

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSuccessWithData(t *testing.T) {
	data := map[string]string{"name": "test"}
	resp := SuccessWithData(data)

	if resp.Code != 200 {
		t.Errorf("期望 code=200, 实际 code=%d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("期望 message=success, 实际 message=%s", resp.Message)
	}
	if resp.Data == nil {
		t.Error("期望 data 不为 nil")
	}
}

func TestSuccessWithMessage(t *testing.T) {
	resp := SuccessWithMessage("创建成功", map[string]int{"id": 1})

	if resp.Code != 200 {
		t.Errorf("期望 code=200, 实际 code=%d", resp.Code)
	}
	if resp.Message != "创建成功" {
		t.Errorf("期望 message=创建成功, 实际 message=%s", resp.Message)
	}
}

func TestSuccessOnly(t *testing.T) {
	resp := SuccessOnly()

	if resp.Code != 200 {
		t.Errorf("期望 code=200, 实际 code=%d", resp.Code)
	}
	if resp.Data != nil {
		t.Errorf("期望 data=nil, 实际 data=%v", resp.Data)
	}
}

func TestSuccessWithPage(t *testing.T) {
	items := []string{"a", "b", "c"}
	resp := SuccessWithPage(items, 100, 1, 20)

	if resp.Code != 200 {
		t.Errorf("期望 code=200, 实际 code=%d", resp.Code)
	}

	paginated, ok := resp.Data.(PaginatedData)
	if !ok {
		t.Fatal("期望 Data 类型为 PaginatedData")
	}
	if paginated.Total != 100 {
		t.Errorf("期望 total=100, 实际 total=%d", paginated.Total)
	}
	if paginated.Page != 1 {
		t.Errorf("期望 page=1, 实际 page=%d", paginated.Page)
	}
	if paginated.PageSize != 20 {
		t.Errorf("期望 pageSize=20, 实际 pageSize=%d", paginated.PageSize)
	}
}

func TestErrorWith(t *testing.T) {
	resp := ErrorWith(CodeBadRequest, "参数错误")

	if resp.Code != 400 {
		t.Errorf("期望 code=400, 实际 code=%d", resp.Code)
	}
	if resp.Message != "参数错误" {
		t.Errorf("期望 message=参数错误, 实际 message=%s", resp.Message)
	}
}

func TestErrorWithCode(t *testing.T) {
	resp := ErrorWithCode(10001, "自定义错误")

	if resp.Code != 10001 {
		t.Errorf("期望 code=10001, 实际 code=%d", resp.Code)
	}
}

func TestConvenienceErrorFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) *APIErrorResponse
		wantCode int
	}{
		{"BadRequestError", BadRequestError, 400},
		{"UnauthorizedError", UnauthorizedError, 401},
		{"ForbiddenError", ForbiddenError, 403},
		{"NotFoundError", NotFoundError, 404},
		{"InternalError", InternalError, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.fn("test message")
			if resp.Code != tt.wantCode {
				t.Errorf("期望 code=%d, 实际 code=%d", tt.wantCode, resp.Code)
			}
			if resp.Message != "test message" {
				t.Errorf("期望 message=test message, 实际 message=%s", resp.Message)
			}
		})
	}
}

func TestAPIResponseJSON(t *testing.T) {
	resp := SuccessWithData(map[string]string{"key": "value"})
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	jsonStr := string(jsonBytes)
	if jsonStr != `{"code":200,"message":"success","data":{"key":"value"}}` {
		t.Errorf("JSON 输出不符合预期: %s", jsonStr)
	}
}

func TestAPIErrorResponseJSON(t *testing.T) {
	resp := BadRequestError("参数无效")
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	jsonStr := string(jsonBytes)
	if jsonStr != `{"code":400,"message":"参数无效"}` {
		t.Errorf("JSON 输出不符合预期: %s", jsonStr)
	}
}

func TestSuccessResponseOmitempty(t *testing.T) {
	resp := SuccessOnly()
	jsonBytes, _ := json.Marshal(resp)
	jsonStr := string(jsonBytes)

	// data 字段为 nil 时应该被省略
	if jsonStr != `{"code":200,"message":"success"}` {
		t.Errorf("data 为 nil 时应被省略: %s", jsonStr)
	}
}

// TestErrorWithCodeData 错误体可选扩展数据：Data 未设置时 JSON 与旧结构完全一致
// （存量调用方无感），设置时随 body 下发（下载通道 2009/2013/2014 携带 expiresAt 等）
func TestErrorWithCodeData(t *testing.T) {
	// 不带 data：输出与旧结构一致
	resp := ErrorWithCodeData(2014, "额度用完", nil)
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}
	if string(jsonBytes) != `{"code":2014,"message":"额度用完"}` {
		t.Errorf("data 为 nil 时期望与旧结构一致, 实际=%s", jsonBytes)
	}

	// 带 data：随 body 下发
	resp2 := ErrorWithCodeData(2014, "额度用完", map[string]any{"taskId": "t1"})
	jsonBytes2, _ := json.Marshal(resp2)
	if !strings.Contains(string(jsonBytes2), `"taskId":"t1"`) {
		t.Errorf("data 应随 body 下发, 实际=%s", jsonBytes2)
	}
}

// ---- W10 稳定错误码契约（内部PRD §7.11）----

// TestAPIErrorResponseWithoutErrorCodeKeepsLegacyJSON error_code 可选向后兼容：
// 未设置时 JSON 输出不含 error_code 键，与旧结构完全一致（旧客户端无感）
func TestAPIErrorResponseWithoutErrorCodeKeepsLegacyJSON(t *testing.T) {
	resp := ErrorWithCode(409, "订单状态非待支付，无法关闭")
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}
	if string(jsonBytes) != `{"code":409,"message":"订单状态非待支付，无法关闭"}` {
		t.Errorf("未设置 error_code 时期望与旧结构一致, 实际=%s", jsonBytes)
	}
}

// TestErrorWithErrorCode 设置 error_code 后错误体携带该字段，code/message 不变
func TestErrorWithErrorCode(t *testing.T) {
	resp := ErrorWithErrorCode(409, "ORDER_NOT_CLOSABLE", "订单状态非待支付，无法关闭")
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}
	if string(jsonBytes) != `{"code":409,"message":"订单状态非待支付，无法关闭","error_code":"ORDER_NOT_CLOSABLE"}` {
		t.Errorf("error_code 契约不符, 实际=%s", jsonBytes)
	}
	if resp.ErrorCode != "ORDER_NOT_CLOSABLE" || resp.Code != 409 {
		t.Errorf("结构体字段不符: code=%d error_code=%s", resp.Code, resp.ErrorCode)
	}
}

// TestErrorWithErrorCodeData error_code 与 data 并发（软删前置检查计数明细）
func TestErrorWithErrorCodeData(t *testing.T) {
	resp := ErrorWithErrorCodeData(409, "USER_HAS_PENDING_RECORDS", "存在未完结数据", map[string]int{"pending_orders": 2})
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}
	if !strings.Contains(string(jsonBytes), `"error_code":"USER_HAS_PENDING_RECORDS"`) ||
		!strings.Contains(string(jsonBytes), `"pending_orders":2`) {
		t.Errorf("error_code+data 契约不符, 实际=%s", jsonBytes)
	}
}

// TestSuccessAPIResponseNeverCarriesErrorCode 成功响应无错误语义，永不输出 error_code
func TestSuccessAPIResponseNeverCarriesErrorCode(t *testing.T) {
	jsonBytes, _ := json.Marshal(SuccessWithData(map[string]string{"action": "granted"}))
	if strings.Contains(string(jsonBytes), "error_code") {
		t.Errorf("成功响应不应携带 error_code: %s", jsonBytes)
	}
}
