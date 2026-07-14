package common

import (
	"encoding/json"
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
