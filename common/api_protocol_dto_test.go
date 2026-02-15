package common

import (
	"encoding/json"
	"testing"
)

func TestDTO(t *testing.T) {
	t.Run("SuccessResponse", func(t *testing.T) {
		t.Run("带消息和数据", func(t *testing.T) {
			resp := SuccessResponse("操作成功", map[string]string{"id": "123"})
			if resp.Code != HTTPStatusOK {
				t.Errorf("期望 Code=%d, 实际 Code=%d", HTTPStatusOK, resp.Code)
			}
			if resp.Message != "操作成功" {
				t.Errorf("期望 Message=%s, 实际 Message=%s", "操作成功", resp.Message)
			}
			if resp.Data == nil {
				t.Error("期望 Data 不为 nil")
			}
		})

		t.Run("仅数据", func(t *testing.T) {
			resp := SuccessWithData(map[string]int{"count": 10})
			if resp.Code != HTTPStatusOK {
				t.Errorf("期望 Code=%d, 实际 Code=%d", HTTPStatusOK, resp.Code)
			}
			if resp.Message != "success" {
				t.Errorf("期望 Message=%s, 实际 Message=%s", "success", resp.Message)
			}
		})

		t.Run("仅消息", func(t *testing.T) {
			resp := SuccessWithMessage("删除成功")
			if resp.Code != HTTPStatusOK {
				t.Errorf("期望 Code=%d, 实际 Code=%d", HTTPStatusOK, resp.Code)
			}
			if resp.Message != "删除成功" {
				t.Errorf("期望 Message=%s, 实际 Message=%s", "删除成功", resp.Message)
			}
			if resp.Data != nil {
				t.Error("期望 Data 为 nil")
			}
		})

		t.Run("带自定义状态码", func(t *testing.T) {
			resp := SuccessResponseWith(HTTPStatusCreated, "创建成功", map[string]string{"id": "abc"})
			if resp.Code != HTTPStatusCreated {
				t.Errorf("期望 Code=%d, 实际 Code=%d", HTTPStatusCreated, resp.Code)
			}
		})
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		resp := ErrorResponse(HTTPStatusNotFound, "用户不存在")
		if resp.Code != HTTPStatusNotFound {
			t.Errorf("期望 Code=%d, 实际 Code=%d", HTTPStatusNotFound, resp.Code)
		}
		if resp.Message != "用户不存在" {
			t.Errorf("期望 Message=%s, 实际 Message=%s", "用户不存在", resp.Message)
		}
		if resp.Data != nil {
			t.Error("期望 Data 为 nil")
		}
	})

	t.Run("预定义错误响应", func(t *testing.T) {
		tests := []struct {
			name     string
			resp     CommonResponse
			expected int
		}{
			{"ErrBadRequest", ErrBadRequest, HTTPStatusBadRequest},
			{"ErrUnauthorized", ErrUnauthorized, HTTPStatusUnauthorized},
			{"ErrForbidden", ErrForbidden, HTTPStatusForbidden},
			{"ErrNotFound", ErrNotFound, HTTPStatusNotFound},
			{"ErrInternalServer", ErrInternalServer, HTTPStatusInternalServerError},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.resp.Code != tt.expected {
					t.Errorf("期望 Code=%d, 实际 Code=%d", tt.expected, tt.resp.Code)
				}
			})
		}
	})

	t.Run("PageResponse", func(t *testing.T) {
		t.Run("默认消息", func(t *testing.T) {
			list := []map[string]string{{"id": "1"}, {"id": "2"}}
			resp := PagedResponse(list, 100, 1, 10)
			if resp.Code != HTTPStatusOK {
				t.Errorf("期望 Code=%d, 实际 Code=%d", HTTPStatusOK, resp.Code)
			}
			if resp.Message != "success" {
				t.Errorf("期望 Message=%s, 实际 Message=%s", "success", resp.Message)
			}
			if resp.Total != 100 {
				t.Errorf("期望 Total=%d, 实际 Total=%d", 100, resp.Total)
			}
			if resp.Page != 1 {
				t.Errorf("期望 Page=%d, 实际 Page=%d", 1, resp.Page)
			}
			if resp.PageSize != 10 {
				t.Errorf("期望 PageSize=%d, 实际 PageSize=%d", 10, resp.PageSize)
			}
		})

		t.Run("自定义消息", func(t *testing.T) {
			list := []string{"a", "b", "c"}
			resp := PagedResponseWithMessage("查询成功", list, 3, 1, 10)
			if resp.Message != "查询成功" {
				t.Errorf("期望 Message=%s, 实际 Message=%s", "查询成功", resp.Message)
			}
		})
	})

	t.Run("ValidationErrorResponse", func(t *testing.T) {
		t.Run("默认状态码", func(t *testing.T) {
			errors := FieldErrors{
				{Field: "username", Message: "不能为空"},
				{Field: "email", Message: "格式不正确"},
			}
			resp := ValidationResponse("验证失败", errors)
			if resp.Code != HTTPStatusBadRequest {
				t.Errorf("期望 Code=%d, 实际 Code=%d", HTTPStatusBadRequest, resp.Code)
			}
			if len(resp.Errors) != 2 {
				t.Errorf("期望 Errors 长度=2, 实际长度=%d", len(resp.Errors))
			}
		})

		t.Run("自定义状态码", func(t *testing.T) {
			errors := FieldErrors{{Field: "password", Message: "长度不足"}}
			resp := ValidationResponseWithCode(HTTPStatusUnprocessableEntity, "数据验证失败", errors)
			if resp.Code != HTTPStatusUnprocessableEntity {
				t.Errorf("期望 Code=%d, 实际 Code=%d", HTTPStatusUnprocessableEntity, resp.Code)
			}
		})
	})

	t.Run("FieldErrors_AddFieldError", func(t *testing.T) {
		var errors FieldErrors
		errors.AddFieldError("name", "名称不能为空")
		errors.AddFieldError("age", "年龄必须大于0")

		if len(errors) != 2 {
			t.Errorf("期望长度=2, 实际长度=%d", len(errors))
		}
		if errors[0].Field != "name" {
			t.Errorf("期望 Field=%s, 实际 Field=%s", "name", errors[0].Field)
		}
		if errors[1].Message != "年龄必须大于0" {
			t.Errorf("期望 Message=%s, 实际 Message=%s", "年龄必须大于0", errors[1].Message)
		}
	})

	t.Run("JSON序列化", func(t *testing.T) {
		t.Run("CommonResponse", func(t *testing.T) {
			resp := SuccessResponse("成功", map[string]int{"count": 1})
			data, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}
			if result["code"].(float64) != float64(HTTPStatusOK) {
				t.Error("JSON 序列化后 code 不正确")
			}
		})

		t.Run("PageResponse", func(t *testing.T) {
			resp := PagedResponse([]string{"a", "b"}, 2, 1, 10)
			data, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}
			if result["total"].(float64) != 2 {
				t.Error("JSON 序列化后 total 不正确")
			}
		})

		t.Run("ValidationErrorResponse", func(t *testing.T) {
			errors := FieldErrors{{Field: "email", Message: "格式错误"}}
			resp := ValidationResponse("验证失败", errors)
			data, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}
			errList := result["errors"].([]interface{})
			if len(errList) != 1 {
				t.Error("JSON 序列化后 errors 长度不正确")
			}
		})
	})
}
