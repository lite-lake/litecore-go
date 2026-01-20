package json

import (
	"testing"
)

// 测试用例结构体
type TestPerson struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Email   string   `json:"email"`
	Tags    []string `json:"tags"`
	Address struct {
		City    string `json:"city"`
		Street  string `json:"street"`
		Country string `json:"country"`
	} `json:"address"`
}

// =========================================
// 基础验证和格式化测试
// =========================================

func TestJSONEngine_IsValid(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		expected bool
	}{
		{
			name:     "有效的对象",
			jsonStr:  `{"name":"test","value":123}`,
			expected: true,
		},
		{
			name:     "有效的数组",
			jsonStr:  `[1,2,3,"test"]`,
			expected: true,
		},
		{
			name:     "有效的字符串",
			jsonStr:  `"hello"`,
			expected: true,
		},
		{
			name:     "有效的数字",
			jsonStr:  `123.45`,
			expected: true,
		},
		{
			name:     "有效的布尔值",
			jsonStr:  `true`,
			expected: true,
		},
		{
			name:     "有效的null",
			jsonStr:  `null`,
			expected: true,
		},
		{
			name:     "无效的JSON-缺少引号",
			jsonStr:  `{name:"test"}`,
			expected: false,
		},
		{
			name:     "无效的JSON-缺少闭合大括号",
			jsonStr:  `{"name":"test"`,
			expected: false,
		},
		{
			name:     "空字符串",
			jsonStr:  ``,
			expected: false,
		},
		{
			name:     "包含特殊字符的字符串",
			jsonStr:  `{"message":"Hello\nWorld\t!"}`,
			expected: true,
		},
		{
			name:     "嵌套对象",
			jsonStr:  `{"user":{"name":"test","age":25}}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := j.IsValid(tt.jsonStr)
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJSONEngine_PrettyPrint(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		indent   string
		wantErr  bool
		contains string
	}{
		{
			name:     "简单对象-2空格缩进",
			jsonStr:  `{"name":"test","value":123}`,
			indent:   "  ",
			wantErr:  false,
			contains: "  ",
		},
		{
			name:     "简单对象-4空格缩进",
			jsonStr:  `{"name":"test","value":123}`,
			indent:   "    ",
			wantErr:  false,
			contains: "    ",
		},
		{
			name:     "简单对象-tab缩进",
			jsonStr:  `{"name":"test","value":123}`,
			indent:   "\t",
			wantErr:  false,
			contains: "\t",
		},
		{
			name:    "嵌套对象",
			jsonStr: `{"user":{"name":"test","age":25}}`,
			indent:  "  ",
			wantErr: false,
		},
		{
			name:    "数组",
			jsonStr: `[1,2,3,4,5]`,
			indent:  "  ",
			wantErr: false,
		},
		{
			name:    "无效的JSON",
			jsonStr: `{invalid}`,
			indent:  "  ",
			wantErr: true,
		},
		{
			name:    "空字符串",
			jsonStr: ``,
			indent:  "  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.PrettyPrint(tt.jsonStr, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyPrint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.contains != "" && !containsString(result, tt.contains) {
				t.Errorf("PrettyPrint() result should contain %q", tt.contains)
			}
		})
	}
}

func TestJSONEngine_PrettyPrintWithIndent(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "简单对象",
			jsonStr: `{"name":"test","value":123}`,
			wantErr: false,
		},
		{
			name:    "复杂嵌套对象",
			jsonStr: `{"user":{"profile":{"name":"test","age":25,"address":{"city":"Beijing"}}}}`,
			wantErr: false,
		},
		{
			name:    "无效JSON",
			jsonStr: `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.PrettyPrintWithIndent(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyPrintWithIndent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("PrettyPrintWithIndent() returned empty string")
			}
		})
	}
}

func TestJSONEngine_Compact(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name         string
		jsonStr      string
		wantErr      bool
		noWhitespace bool
	}{
		{
			name:         "带空格的对象",
			jsonStr:      `{ "name" : "test" , "value" : 123 }`,
			wantErr:      false,
			noWhitespace: true,
		},
		{
			name:         "带换行的对象",
			jsonStr:      "{\n  \"name\": \"test\",\n  \"value\": 123\n}",
			wantErr:      false,
			noWhitespace: true,
		},
		{
			name:         "已经压缩的对象",
			jsonStr:      `{"name":"test","value":123}`,
			wantErr:      false,
			noWhitespace: true,
		},
		{
			name:    "无效JSON",
			jsonStr: `{invalid}`,
			wantErr: true,
		},
		{
			name:    "空字符串",
			jsonStr: ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.Compact(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.noWhitespace {
				if containsWhitespace(result) {
					t.Error("Compact() result contains whitespace")
				}
			}
		})
	}
}

func TestJSONEngine_Escape(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "普通字符串",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "包含换行符",
			input:    "hello\nworld",
			expected: "hello\\nworld",
		},
		{
			name:     "包含制表符",
			input:    "hello\tworld",
			expected: "hello\\tworld",
		},
		{
			name:     "包含引号",
			input:    `say "hello"`,
			expected: `say \"hello\"`,
		},
		{
			name:     "包含反斜杠",
			input:    "path\\to\\file",
			expected: "path\\\\to\\\\file",
		},
		{
			name:     "包含特殊字符",
			input:    "\r\n\b\f",
			expected: "\\r\\n\\b\\f",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "中文字符",
			input:    "你好世界",
			expected: "你好世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.Escape(tt.input)
			if err != nil {
				t.Errorf("Escape() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Escape() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJSONEngine_Unescape(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "普通字符串",
			input:    "hello",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "转义的换行符",
			input:    "hello\\nworld",
			expected: "hello\nworld",
			wantErr:  false,
		},
		{
			name:     "转义的制表符",
			input:    "hello\\tworld",
			expected: "hello\tworld",
			wantErr:  false,
		},
		{
			name:     "转义的引号",
			input:    `say \"hello\"`,
			expected: `say "hello"`,
			wantErr:  false,
		},
		{
			name:     "转义的反斜杠",
			input:    "path\\\\to\\\\file",
			expected: "path\\to\\file",
			wantErr:  false,
		},
		{
			name:     "Unicode转义",
			input:    "\\u4e2d\\u6587",
			expected: "中文",
			wantErr:  false,
		},
		{
			name:     "无效的转义序列",
			input:    "\\x",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.Unescape(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unescape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Unescape() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// =========================================
// 数据转换测试
// =========================================

func TestJSONEngine_ToMap(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
		check   func(map[string]interface{}) bool
	}{
		{
			name:    "简单对象",
			jsonStr: `{"name":"test","value":123}`,
			wantErr: false,
			check: func(m map[string]interface{}) bool {
				return m["name"] == "test" && m["value"] == float64(123)
			},
		},
		{
			name:    "嵌套对象",
			jsonStr: `{"user":{"name":"test","age":25}}`,
			wantErr: false,
			check: func(m map[string]interface{}) bool {
				if user, ok := m["user"].(map[string]interface{}); ok {
					return user["name"] == "test" && user["age"] == float64(25)
				}
				return false
			},
		},
		{
			name:    "包含数组",
			jsonStr: `{"tags":["a","b","c"]}`,
			wantErr: false,
			check: func(m map[string]interface{}) bool {
				if tags, ok := m["tags"].([]interface{}); ok {
					return len(tags) == 3
				}
				return false
			},
		},
		{
			name:    "包含null值",
			jsonStr: `{"value":null}`,
			wantErr: false,
			check: func(m map[string]interface{}) bool {
				return m["value"] == nil
			},
		},
		{
			name:    "包含布尔值",
			jsonStr: `{"active":true,"deleted":false}`,
			wantErr: false,
			check: func(m map[string]interface{}) bool {
				return m["active"] == true && m["deleted"] == false
			},
		},
		{
			name:    "数组而非对象",
			jsonStr: `[1,2,3]`,
			wantErr: true, // ToMap 不允许数组，只能解析对象
			check:   nil,
		},
		{
			name:    "无效JSON",
			jsonStr: `{invalid}`,
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.ToMap(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("ToMap() validation check failed")
			}
		})
	}
}

func TestJSONEngine_ToMapStrict(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "有效对象",
			jsonStr: `{"name":"test"}`,
			wantErr: false,
		},
		{
			name:    "数组",
			jsonStr: `[1,2,3]`,
			wantErr: true,
		},
		{
			name:    "字符串",
			jsonStr: `"hello"`,
			wantErr: true,
		},
		{
			name:    "数字",
			jsonStr: `123`,
			wantErr: true,
		},
		{
			name:    "null",
			jsonStr: `null`,
			wantErr: true,
		},
		{
			name:    "无效JSON",
			jsonStr: `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.ToMapStrict(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMapStrict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("ToMapStrict() returned nil for valid object")
			}
		})
	}
}

func TestJSONEngine_ToStruct(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
		check   func(*TestPerson) bool
	}{
		{
			name: "有效的结构体",
			jsonStr: `{"name":"Alice","age":30,"email":"alice@example.com","tags":["dev","go"],` +
				`"address":{"city":"Beijing","street":"Main St","country":"China"}}`,
			wantErr: false,
			check: func(p *TestPerson) bool {
				return p.Name == "Alice" && p.Age == 30 && p.Email == "alice@example.com"
			},
		},
		{
			name:    "部分字段",
			jsonStr: `{"name":"Bob","age":25}`,
			wantErr: false,
			check: func(p *TestPerson) bool {
				return p.Name == "Bob" && p.Age == 25
			},
		},
		{
			name:    "嵌套结构",
			jsonStr: `{"address":{"city":"Shanghai","street":"Nanjing Rd","country":"China"}}`,
			wantErr: false,
			check: func(p *TestPerson) bool {
				return p.Address.City == "Shanghai"
			},
		},
		{
			name:    "无效JSON",
			jsonStr: `{invalid}`,
			wantErr: true,
			check:   nil,
		},
		{
			name:    "类型不匹配",
			jsonStr: `{"name":"Charlie","age":"not a number"}`,
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var person TestPerson
			err := j.ToStruct(tt.jsonStr, &person)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(&person) {
				t.Error("ToStruct() validation check failed")
			}
		})
	}
}

func TestJSONEngine_FromMap(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "简单对象",
			data: map[string]interface{}{
				"name":  "test",
				"value": 123,
			},
			wantErr: false,
			check: func(s string) bool {
				return containsString(s, `"name"`) && containsString(s, `"value"`)
			},
		},
		{
			name: "嵌套对象",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
					"age":  30,
				},
			},
			wantErr: false,
			check: func(s string) bool {
				return containsString(s, `"user"`) && containsString(s, `"name"`)
			},
		},
		{
			name: "包含数组",
			data: map[string]interface{}{
				"tags": []interface{}{"a", "b", "c"},
			},
			wantErr: false,
			check: func(s string) bool {
				return containsString(s, `"tags"`) && containsString(s, `[`)
			},
		},
		{
			name: "包含特殊类型",
			data: map[string]interface{}{
				"bool":  true,
				"null":  nil,
				"float": 123.45,
			},
			wantErr: false,
			check: func(s string) bool {
				return containsString(s, `true`) && containsString(s, `null`)
			},
		},
		{
			name:    "空对象",
			data:    map[string]interface{}{},
			wantErr: false,
			check: func(s string) bool {
				return s == `{}`
			},
		},
		{
			name:    "nil",
			data:    nil,
			wantErr: false,
			check: func(s string) bool {
				return s == `null`
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.FromMap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("FromMap() result = %v, validation check failed", result)
			}
		})
	}
}

func TestJSONEngine_FromStruct(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "结构体",
			data: TestPerson{
				Name:  "Alice",
				Age:   30,
				Email: "alice@example.com",
			},
			wantErr: false,
			check: func(s string) bool {
				return containsString(s, `"name"`) && containsString(s, `"Alice"`)
			},
		},
		{
			name: "map",
			data: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
			check: func(s string) bool {
				return containsString(s, `"key1"`) && containsString(s, `"value1"`)
			},
		},
		{
			name:    "slice",
			data:    []int{1, 2, 3, 4, 5},
			wantErr: false,
			check: func(s string) bool {
				return containsString(s, `[`) && containsString(s, `]`)
			},
		},
		{
			name:    "基本类型-字符串",
			data:    "hello",
			wantErr: false,
			check: func(s string) bool {
				return s == `"hello"`
			},
		},
		{
			name:    "基本类型-数字",
			data:    123,
			wantErr: false,
			check: func(s string) bool {
				return s == `123`
			},
		},
		{
			name:    "基本类型-布尔",
			data:    true,
			wantErr: false,
			check: func(s string) bool {
				return s == `true`
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.FromStruct(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("FromStruct() result = %v, validation check failed", result)
			}
		})
	}
}

func TestJSONEngine_FromStructWithIndent(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name    string
		data    interface{}
		indent  string
		wantErr bool
	}{
		{
			name:    "结构体-2空格",
			data:    TestPerson{Name: "Alice", Age: 30},
			indent:  "  ",
			wantErr: false,
		},
		{
			name:    "结构体-4空格",
			data:    TestPerson{Name: "Bob", Age: 25},
			indent:  "    ",
			wantErr: false,
		},
		{
			name:    "结构体-tab",
			data:    TestPerson{Name: "Charlie", Age: 35},
			indent:  "\t",
			wantErr: false,
		},
		{
			name:    "嵌套结构",
			data:    map[string]interface{}{"user": map[string]string{"name": "Alice"}},
			indent:  "  ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.FromStructWithIndent(tt.data, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromStructWithIndent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("FromStructWithIndent() returned empty string")
			}
		})
	}
}

// =========================================
// 路径操作测试
// =========================================

func TestJSONEngine_GetValue(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		path     string
		wantErr  bool
		expected interface{}
	}{
		{
			name:     "根路径",
			jsonStr:  `{"name":"test"}`,
			path:     "",
			wantErr:  false,
			expected: map[string]interface{}{"name": "test"},
		},
		{
			name:     "根路径-点",
			jsonStr:  `{"name":"test"}`,
			path:     ".",
			wantErr:  false,
			expected: map[string]interface{}{"name": "test"},
		},
		{
			name:     "简单键",
			jsonStr:  `{"name":"Alice"}`,
			path:     "name",
			wantErr:  false,
			expected: "Alice",
		},
		{
			name:     "嵌套对象",
			jsonStr:  `{"user":{"name":"Alice","age":30}}`,
			path:     "user.name",
			wantErr:  false,
			expected: "Alice",
		},
		{
			name:     "深层嵌套",
			jsonStr:  `{"a":{"b":{"c":{"d":"value"}}}}`,
			path:     "a.b.c.d",
			wantErr:  false,
			expected: "value",
		},
		{
			name:     "数组索引",
			jsonStr:  `{"items":[1,2,3,4,5]}`,
			path:     "items.0",
			wantErr:  false,
			expected: float64(1),
		},
		{
			name:     "数组中的对象",
			jsonStr:  `{"users":[{"name":"Alice"},{"name":"Bob"}]}`,
			path:     "users.1.name",
			wantErr:  false,
			expected: "Bob",
		},
		{
			name:     "不存在的键",
			jsonStr:  `{"name":"Alice"}`,
			path:     "age",
			wantErr:  true,
			expected: nil,
		},
		{
			name:     "无效的数组索引",
			jsonStr:  `{"items":[1,2,3]}`,
			path:     "items.abc",
			wantErr:  true,
			expected: nil,
		},
		{
			name:     "数组索引越界",
			jsonStr:  `{"items":[1,2,3]}`,
			path:     "items.10",
			wantErr:  true,
			expected: nil,
		},
		{
			name:     "负数索引",
			jsonStr:  `{"items":[1,2,3]}`,
			path:     "items.-1",
			wantErr:  true,
			expected: nil,
		},
		{
			name:     "在标量值上访问",
			jsonStr:  `{"name":"Alice"}`,
			path:     "name.invalid",
			wantErr:  true,
			expected: nil,
		},
		{
			name:     "无效JSON",
			jsonStr:  `{invalid}`,
			path:     "name",
			wantErr:  true,
			expected: nil,
		},
		{
			name:     "获取数组",
			jsonStr:  `{"tags":["a","b","c"]}`,
			path:     "tags",
			wantErr:  false,
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "获取对象",
			jsonStr:  `{"user":{"name":"Alice"}}`,
			path:     "user",
			wantErr:  false,
			expected: map[string]interface{}{"name": "Alice"},
		},
		{
			name:     "获取null值",
			jsonStr:  `{"value":null}`,
			path:     "value",
			wantErr:  false,
			expected: nil,
		},
		{
			name:     "获取数字",
			jsonStr:  `{"value":123.45}`,
			path:     "value",
			wantErr:  false,
			expected: float64(123.45),
		},
		{
			name:     "获取布尔值",
			jsonStr:  `{"active":true}`,
			path:     "active",
			wantErr:  false,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.GetValue(tt.jsonStr, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(result, tt.expected) {
				t.Errorf("GetValue() = %v (type %T), want %v (type %T)", result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestJSONEngine_GetString(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		path     string
		wantErr  bool
		expected string
	}{
		{
			name:     "获取字符串",
			jsonStr:  `{"name":"Alice"}`,
			path:     "name",
			wantErr:  false,
			expected: "Alice",
		},
		{
			name:     "嵌套字符串",
			jsonStr:  `{"user":{"name":"Bob"}}`,
			path:     "user.name",
			wantErr:  false,
			expected: "Bob",
		},
		{
			name:     "数字转字符串",
			jsonStr:  `{"value":123}`,
			path:     "value",
			wantErr:  false,
			expected: "123",
		},
		{
			name:     "布尔值转字符串",
			jsonStr:  `{"active":true}`,
			path:     "active",
			wantErr:  false,
			expected: "true",
		},
		{
			name:     "空字符串",
			jsonStr:  `{"name":""}`,
			path:     "name",
			wantErr:  false,
			expected: "",
		},
		{
			name:     "特殊字符",
			jsonStr:  `{"message":"Hello\nWorld"}`,
			path:     "message",
			wantErr:  false,
			expected: "Hello\nWorld",
		},
		{
			name:     "null值-错误",
			jsonStr:  `{"value":null}`,
			path:     "value",
			wantErr:  true,
			expected: "",
		},
		{
			name:     "对象转字符串",
			jsonStr:  `{"data":{"key":"value"}}`,
			path:     "data",
			wantErr:  false,
			expected: "map[key:value]",
		},
		{
			name:     "数组转字符串",
			jsonStr:  `{"items":[1,2,3]}`,
			path:     "items",
			wantErr:  false,
			expected: "[1 2 3]",
		},
		{
			name:     "不存在的路径",
			jsonStr:  `{"name":"Alice"}`,
			path:     "age",
			wantErr:  true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.GetString(tt.jsonStr, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("GetString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJSONEngine_GetFloat64(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		path     string
		wantErr  bool
		expected float64
	}{
		{
			name:     "获取整数",
			jsonStr:  `{"value":123}`,
			path:     "value",
			wantErr:  false,
			expected: 123,
		},
		{
			name:     "获取浮点数",
			jsonStr:  `{"value":123.45}`,
			path:     "value",
			wantErr:  false,
			expected: 123.45,
		},
		{
			name:     "负数",
			jsonStr:  `{"value":-99.99}`,
			path:     "value",
			wantErr:  false,
			expected: -99.99,
		},
		{
			name:     "零",
			jsonStr:  `{"value":0}`,
			path:     "value",
			wantErr:  false,
			expected: 0,
		},
		{
			name:     "嵌套数字",
			jsonStr:  `{"data":{"value":42.5}}`,
			path:     "data.value",
			wantErr:  false,
			expected: 42.5,
		},
		{
			name:     "数字字符串",
			jsonStr:  `{"value":"123.45"}`,
			path:     "value",
			wantErr:  false,
			expected: 123.45,
		},
		{
			name:     "整数字符串",
			jsonStr:  `{"value":"100"}`,
			path:     "value",
			wantErr:  false,
			expected: 100,
		},
		{
			name:     "科学计数法字符串",
			jsonStr:  `{"value":"1.23e2"}`,
			path:     "value",
			wantErr:  false,
			expected: 123,
		},
		{
			name:     "字符串值-错误",
			jsonStr:  `{"value":"not a number"}`,
			path:     "value",
			wantErr:  true,
			expected: 0,
		},
		{
			name:     "布尔值-错误",
			jsonStr:  `{"value":true}`,
			path:     "value",
			wantErr:  true,
			expected: 0,
		},
		{
			name:     "对象-错误",
			jsonStr:  `{"value":{}}`,
			path:     "value",
			wantErr:  true,
			expected: 0,
		},
		{
			name:     "null-错误",
			jsonStr:  `{"value":null}`,
			path:     "value",
			wantErr:  true,
			expected: 0,
		},
		{
			name:     "不存在的路径",
			jsonStr:  `{"name":"Alice"}`,
			path:     "age",
			wantErr:  true,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.GetFloat64(tt.jsonStr, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFloat64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("GetFloat64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJSONEngine_GetBool(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		path     string
		wantErr  bool
		expected bool
	}{
		{
			name:     "true值",
			jsonStr:  `{"active":true}`,
			path:     "active",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "false值",
			jsonStr:  `{"active":false}`,
			path:     "active",
			wantErr:  false,
			expected: false,
		},
		{
			name:     "嵌套布尔值",
			jsonStr:  `{"data":{"active":true}}`,
			path:     "data.active",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "字符串true",
			jsonStr:  `{"active":"true"}`,
			path:     "active",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "字符串false",
			jsonStr:  `{"active":"false"}`,
			path:     "active",
			wantErr:  false,
			expected: false,
		},
		{
			name:     "字符串1",
			jsonStr:  `{"active":"1"}`,
			path:     "active",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "字符串0",
			jsonStr:  `{"active":"0"}`,
			path:     "active",
			wantErr:  false,
			expected: false,
		},
		{
			name:     "数字-错误",
			jsonStr:  `{"value":123}`,
			path:     "value",
			wantErr:  true,
			expected: false,
		},
		{
			name:     "对象-错误",
			jsonStr:  `{"value":{}}`,
			path:     "value",
			wantErr:  true,
			expected: false,
		},
		{
			name:     "数组-错误",
			jsonStr:  `{"value":[]}`,
			path:     "value",
			wantErr:  true,
			expected: false,
		},
		{
			name:     "null-错误",
			jsonStr:  `{"value":null}`,
			path:     "value",
			wantErr:  true,
			expected: false,
		},
		{
			name:     "无效字符串-错误",
			jsonStr:  `{"active":"yes"}`,
			path:     "active",
			wantErr:  true,
			expected: false,
		},
		{
			name:     "不存在的路径",
			jsonStr:  `{"name":"Alice"}`,
			path:     "active",
			wantErr:  true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.GetBool(tt.jsonStr, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("GetBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// =========================================
// 高级操作测试
// =========================================

func TestJSONEngine_Merge(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr1 string
		jsonStr2 string
		wantErr  bool
		check    func(string) bool
	}{
		{
			name:     "简单对象合并",
			jsonStr1: `{"name":"Alice"}`,
			jsonStr2: `{"age":30}`,
			wantErr:  false,
			check: func(s string) bool {
				return containsString(s, `"name"`) && containsString(s, `"age"`)
			},
		},
		{
			name:     "覆盖键",
			jsonStr1: `{"name":"Alice","age":25}`,
			jsonStr2: `{"age":30}`,
			wantErr:  false,
			check: func(s string) bool {
				return containsString(s, `"age":30`)
			},
		},
		{
			name:     "嵌套对象合并",
			jsonStr1: `{"user":{"name":"Alice"}}`,
			jsonStr2: `{"user":{"age":30}}`,
			wantErr:  false,
			check: func(s string) bool {
				return containsString(s, `"name"`) && containsString(s, `"age"`)
			},
		},
		{
			name:     "深层嵌套合并",
			jsonStr1: `{"a":{"b":{"c":1}}}`,
			jsonStr2: `{"a":{"b":{"d":2}}}`,
			wantErr:  false,
			check: func(s string) bool {
				return containsString(s, `"c":1`) && containsString(s, `"d":2`)
			},
		},
		{
			name:     "第一个为空",
			jsonStr1: `{}`,
			jsonStr2: `{"name":"Alice"}`,
			wantErr:  false,
			check: func(s string) bool {
				return containsString(s, `"name"`)
			},
		},
		{
			name:     "第二个为空",
			jsonStr1: `{"name":"Alice"}`,
			jsonStr2: `{}`,
			wantErr:  false,
			check: func(s string) bool {
				return containsString(s, `"name"`)
			},
		},
		{
			name:     "混合类型覆盖",
			jsonStr1: `{"data":{"name":"Alice"}}`,
			jsonStr2: `{"data":"simple string"}`,
			wantErr:  false,
			check: func(s string) bool {
				return containsString(s, `"data":"simple string"`)
			},
		},
		{
			name:     "包含数组",
			jsonStr1: `{"tags":["a","b"]}`,
			jsonStr2: `{"name":"Alice"}`,
			wantErr:  false,
			check: func(s string) bool {
				return containsString(s, `"tags"`) && containsString(s, `"name"`)
			},
		},
		{
			name:     "第一个无效JSON",
			jsonStr1: `{invalid}`,
			jsonStr2: `{"name":"Alice"}`,
			wantErr:  true,
			check:    nil,
		},
		{
			name:     "第二个无效JSON",
			jsonStr1: `{"name":"Alice"}`,
			jsonStr2: `{invalid}`,
			wantErr:  true,
			check:    nil,
		},
		{
			name:     "第一个非对象",
			jsonStr1: `[1,2,3]`,
			jsonStr2: `{"name":"Alice"}`,
			wantErr:  true,
			check:    nil,
		},
		{
			name:     "第二个非对象",
			jsonStr1: `{"name":"Alice"}`,
			jsonStr2: `[1,2,3]`,
			wantErr:  true,
			check:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.Merge(tt.jsonStr1, tt.jsonStr2)
			if (err != nil) != tt.wantErr {
				t.Errorf("Merge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("Merge() result = %v, validation check failed", result)
			}
		})
	}
}

func TestJSONEngine_Diff(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name       string
		jsonStr1   string
		jsonStr2   string
		wantErr    bool
		expectDiff bool
	}{
		{
			name:       "相同对象",
			jsonStr1:   `{"name":"Alice","age":30}`,
			jsonStr2:   `{"name":"Alice","age":30}`,
			wantErr:    false,
			expectDiff: false,
		},
		{
			name:       "不同值",
			jsonStr1:   `{"name":"Alice"}`,
			jsonStr2:   `{"name":"Bob"}`,
			wantErr:    false,
			expectDiff: true,
		},
		{
			name:       "不同键",
			jsonStr1:   `{"name":"Alice"}`,
			jsonStr2:   `{"age":30}`,
			wantErr:    false,
			expectDiff: true,
		},
		{
			name:       "不同结构-对象vs数组",
			jsonStr1:   `{"name":"Alice"}`,
			jsonStr2:   `["Alice"]`,
			wantErr:    false,
			expectDiff: true,
		},
		{
			name:       "嵌套对象-相同",
			jsonStr1:   `{"user":{"name":"Alice","age":30}}`,
			jsonStr2:   `{"user":{"name":"Alice","age":30}}`,
			wantErr:    false,
			expectDiff: false,
		},
		{
			name:       "嵌套对象-不同",
			jsonStr1:   `{"user":{"name":"Alice","age":30}}`,
			jsonStr2:   `{"user":{"name":"Bob","age":30}}`,
			wantErr:    false,
			expectDiff: true,
		},
		{
			name:       "数组-相同顺序",
			jsonStr1:   `[1,2,3]`,
			jsonStr2:   `[1,2,3]`,
			wantErr:    false,
			expectDiff: false,
		},
		{
			name:       "数组-不同顺序",
			jsonStr1:   `[1,2,3]`,
			jsonStr2:   `[3,2,1]`,
			wantErr:    false,
			expectDiff: true,
		},
		{
			name:       "布尔值-相同",
			jsonStr1:   `true`,
			jsonStr2:   `true`,
			wantErr:    false,
			expectDiff: false,
		},
		{
			name:       "布尔值-不同",
			jsonStr1:   `true`,
			jsonStr2:   `false`,
			wantErr:    false,
			expectDiff: true,
		},
		{
			name:       "null值-相同",
			jsonStr1:   `null`,
			jsonStr2:   `null`,
			wantErr:    false,
			expectDiff: false,
		},
		{
			name:       "字符串-相同",
			jsonStr1:   `"hello"`,
			jsonStr2:   `"hello"`,
			wantErr:    false,
			expectDiff: false,
		},
		{
			name:       "字符串-不同",
			jsonStr1:   `"hello"`,
			jsonStr2:   `"world"`,
			wantErr:    false,
			expectDiff: true,
		},
		{
			name:       "数字-相同",
			jsonStr1:   `123`,
			jsonStr2:   `123`,
			wantErr:    false,
			expectDiff: false,
		},
		{
			name:       "数字-不同",
			jsonStr1:   `123`,
			jsonStr2:   `456`,
			wantErr:    false,
			expectDiff: true,
		},
		{
			name:       "第一个无效JSON",
			jsonStr1:   `{invalid}`,
			jsonStr2:   `{"name":"Alice"}`,
			wantErr:    true,
			expectDiff: false,
		},
		{
			name:       "第二个无效JSON",
			jsonStr1:   `{"name":"Alice"}`,
			jsonStr2:   `{invalid}`,
			wantErr:    true,
			expectDiff: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.Diff(tt.jsonStr1, tt.jsonStr2)
			if (err != nil) != tt.wantErr {
				t.Errorf("Diff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expectDiff {
				t.Errorf("Diff() = %v, want %v", result, tt.expectDiff)
			}
		})
	}
}

// =========================================
// 实用工具测试
// =========================================

func TestJSONEngine_IsObject(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		expected bool
	}{
		{
			name:     "简单对象",
			jsonStr:  `{"name":"Alice"}`,
			expected: true,
		},
		{
			name:     "嵌套对象",
			jsonStr:  `{"user":{"name":"Alice"}}`,
			expected: true,
		},
		{
			name:     "空对象",
			jsonStr:  `{}`,
			expected: true,
		},
		{
			name:     "带空格的对象",
			jsonStr:  ` { "name": "Alice" } `,
			expected: true,
		},
		{
			name:     "带换行的对象",
			jsonStr:  "\n{\n  \"name\": \"Alice\"\n}\n",
			expected: true,
		},
		{
			name:     "数组",
			jsonStr:  `[1,2,3]`,
			expected: false,
		},
		{
			name:     "字符串",
			jsonStr:  `"hello"`,
			expected: false,
		},
		{
			name:     "数字",
			jsonStr:  `123`,
			expected: false,
		},
		{
			name:     "布尔值",
			jsonStr:  `true`,
			expected: false,
		},
		{
			name:     "null",
			jsonStr:  `null`,
			expected: false,
		},
		{
			name:     "空字符串",
			jsonStr:  ``,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := j.IsObject(tt.jsonStr)
			if result != tt.expected {
				t.Errorf("IsObject() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJSONEngine_IsArray(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		expected bool
	}{
		{
			name:     "简单数组",
			jsonStr:  `[1,2,3]`,
			expected: true,
		},
		{
			name:     "对象数组",
			jsonStr:  `[{"name":"Alice"},{"name":"Bob"}]`,
			expected: true,
		},
		{
			name:     "空数组",
			jsonStr:  `[]`,
			expected: true,
		},
		{
			name:     "带空格的数组",
			jsonStr:  ` [ 1, 2, 3 ] `,
			expected: true,
		},
		{
			name:     "带换行的数组",
			jsonStr:  "\n[\n  1,\n  2\n]\n",
			expected: true,
		},
		{
			name:     "对象",
			jsonStr:  `{"name":"Alice"}`,
			expected: false,
		},
		{
			name:     "字符串",
			jsonStr:  `"hello"`,
			expected: false,
		},
		{
			name:     "数字",
			jsonStr:  `123`,
			expected: false,
		},
		{
			name:     "布尔值",
			jsonStr:  `true`,
			expected: false,
		},
		{
			name:     "null",
			jsonStr:  `null`,
			expected: false,
		},
		{
			name:     "空字符串",
			jsonStr:  ``,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := j.IsArray(tt.jsonStr)
			if result != tt.expected {
				t.Errorf("IsArray() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJSONEngine_GetType(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		path     string
		wantErr  bool
		expected string
	}{
		{
			name:     "对象类型",
			jsonStr:  `{"user":{"name":"Alice"}}`,
			path:     "user",
			wantErr:  false,
			expected: "object",
		},
		{
			name:     "数组类型",
			jsonStr:  `{"items":[1,2,3]}`,
			path:     "items",
			wantErr:  false,
			expected: "array",
		},
		{
			name:     "字符串类型",
			jsonStr:  `{"name":"Alice"}`,
			path:     "name",
			wantErr:  false,
			expected: "string",
		},
		{
			name:     "整数类型",
			jsonStr:  `{"value":123}`,
			path:     "value",
			wantErr:  false,
			expected: "number",
		},
		{
			name:     "浮点类型",
			jsonStr:  `{"value":123.45}`,
			path:     "value",
			wantErr:  false,
			expected: "number",
		},
		{
			name:     "布尔类型-true",
			jsonStr:  `{"active":true}`,
			path:     "active",
			wantErr:  false,
			expected: "boolean",
		},
		{
			name:     "布尔类型-false",
			jsonStr:  `{"active":false}`,
			path:     "active",
			wantErr:  false,
			expected: "boolean",
		},
		{
			name:     "null类型",
			jsonStr:  `{"value":null}`,
			path:     "value",
			wantErr:  false,
			expected: "null",
		},
		{
			name:     "根对象",
			jsonStr:  `{"name":"Alice"}`,
			path:     "",
			wantErr:  false,
			expected: "object",
		},
		{
			name:     "根数组",
			jsonStr:  `[1,2,3]`,
			path:     "",
			wantErr:  false,
			expected: "array",
		},
		{
			name:     "不存在的路径",
			jsonStr:  `{"name":"Alice"}`,
			path:     "age",
			wantErr:  true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.GetType(tt.jsonStr, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("GetType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJSONEngine_GetKeys(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name        string
		jsonStr     string
		path        string
		wantErr     bool
		expectedLen int
		contains    []string
	}{
		{
			name:        "简单对象",
			jsonStr:     `{"name":"Alice","age":30}`,
			path:        "",
			wantErr:     false,
			expectedLen: 2,
			contains:    []string{"name", "age"},
		},
		{
			name:        "嵌套对象",
			jsonStr:     `{"user":{"name":"Alice","age":30}}`,
			path:        "user",
			wantErr:     false,
			expectedLen: 2,
			contains:    []string{"name", "age"},
		},
		{
			name:        "空对象",
			jsonStr:     `{"empty":{}}`,
			path:        "empty",
			wantErr:     false,
			expectedLen: 0,
			contains:    []string{},
		},
		{
			name:        "对象包含多种类型",
			jsonStr:     `{"data":{"str":"text","num":123,"bool":true,"null":null,"arr":[1,2],"obj":{}}}`,
			path:        "data",
			wantErr:     false,
			expectedLen: 6,
			contains:    []string{"str", "num", "bool", "null", "arr", "obj"},
		},
		{
			name:        "深层嵌套",
			jsonStr:     `{"a":{"b":{"c":{"d":"value"}}}}`,
			path:        "a.b.c",
			wantErr:     false,
			expectedLen: 1,
			contains:    []string{"d"},
		},
		{
			name:        "数组-错误",
			jsonStr:     `{"items":[1,2,3]}`,
			path:        "items",
			wantErr:     true,
			expectedLen: 0,
			contains:    nil,
		},
		{
			name:        "字符串-错误",
			jsonStr:     `{"name":"Alice"}`,
			path:        "name",
			wantErr:     true,
			expectedLen: 0,
			contains:    nil,
		},
		{
			name:        "数字-错误",
			jsonStr:     `{"value":123}`,
			path:        "value",
			wantErr:     true,
			expectedLen: 0,
			contains:    nil,
		},
		{
			name:        "不存在的路径",
			jsonStr:     `{"name":"Alice"}`,
			path:        "age",
			wantErr:     true,
			expectedLen: 0,
			contains:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.GetKeys(tt.jsonStr, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != tt.expectedLen {
					t.Errorf("GetKeys() len = %v, want %v", len(result), tt.expectedLen)
				}
				for _, key := range tt.contains {
					if !containsStringInSlice(result, key) {
						t.Errorf("GetKeys() result should contain %q", key)
					}
				}
			}
		})
	}
}

func TestJSONEngine_GetSize(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name        string
		jsonStr     string
		path        string
		wantErr     bool
		expectedLen int
	}{
		{
			name:        "简单对象",
			jsonStr:     `{"name":"Alice","age":30,"city":"Beijing"}`,
			path:        "",
			wantErr:     false,
			expectedLen: 3,
		},
		{
			name:        "嵌套对象",
			jsonStr:     `{"user":{"name":"Alice","age":30}}`,
			path:        "user",
			wantErr:     false,
			expectedLen: 2,
		},
		{
			name:        "空对象",
			jsonStr:     `{"empty":{}}`,
			path:        "empty",
			wantErr:     false,
			expectedLen: 0,
		},
		{
			name:        "简单数组",
			jsonStr:     `{"items":[1,2,3,4,5]}`,
			path:        "items",
			wantErr:     false,
			expectedLen: 5,
		},
		{
			name:        "嵌套数组",
			jsonStr:     `{"data":{"items":[1,2,3]}}`,
			path:        "data.items",
			wantErr:     false,
			expectedLen: 3,
		},
		{
			name:        "空数组",
			jsonStr:     `{"empty":[]}`,
			path:        "empty",
			wantErr:     false,
			expectedLen: 0,
		},
		{
			name:        "对象数组",
			jsonStr:     `{"users":[{"name":"Alice"},{"name":"Bob"},{"name":"Charlie"}]}`,
			path:        "users",
			wantErr:     false,
			expectedLen: 3,
		},
		{
			name:        "根数组",
			jsonStr:     `[1,2,3,4]`,
			path:        "",
			wantErr:     false,
			expectedLen: 4,
		},
		{
			name:        "字符串-错误",
			jsonStr:     `{"name":"Alice"}`,
			path:        "name",
			wantErr:     true,
			expectedLen: 0,
		},
		{
			name:        "数字-错误",
			jsonStr:     `{"value":123}`,
			path:        "value",
			wantErr:     true,
			expectedLen: 0,
		},
		{
			name:        "布尔值-错误",
			jsonStr:     `{"active":true}`,
			path:        "active",
			wantErr:     true,
			expectedLen: 0,
		},
		{
			name:        "null-错误",
			jsonStr:     `{"value":null}`,
			path:        "value",
			wantErr:     true,
			expectedLen: 0,
		},
		{
			name:        "不存在的路径",
			jsonStr:     `{"name":"Alice"}`,
			path:        "age",
			wantErr:     true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.GetSize(tt.jsonStr, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expectedLen {
				t.Errorf("GetSize() = %v, want %v", result, tt.expectedLen)
			}
		})
	}
}

func TestJSONEngine_Contains(t *testing.T) {
	j := newJSONEngine()

	tests := []struct {
		name     string
		jsonStr  string
		path     string
		key      string
		wantErr  bool
		expected bool
	}{
		{
			name:     "存在的键",
			jsonStr:  `{"name":"Alice","age":30}`,
			path:     "",
			key:      "name",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "不存在的键",
			jsonStr:  `{"name":"Alice","age":30}`,
			path:     "",
			key:      "email",
			wantErr:  false,
			expected: false,
		},
		{
			name:     "嵌套对象-存在",
			jsonStr:  `{"user":{"name":"Alice","age":30}}`,
			path:     "user",
			key:      "name",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "嵌套对象-不存在",
			jsonStr:  `{"user":{"name":"Alice","age":30}}`,
			path:     "user",
			key:      "email",
			wantErr:  false,
			expected: false,
		},
		{
			name:     "空对象",
			jsonStr:  `{"empty":{}}`,
			path:     "empty",
			key:      "anything",
			wantErr:  false,
			expected: false,
		},
		{
			name:     "多种类型",
			jsonStr:  `{"data":{"str":"text","num":123,"bool":true}}`,
			path:     "data",
			key:      "num",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "深层嵌套",
			jsonStr:  `{"a":{"b":{"c":{"d":"value","e":"another"}}}}`,
			path:     "a.b.c",
			key:      "d",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "特殊字符键名",
			jsonStr:  `{"data":{"key-with-dash":"value","key_with_underscore":"value2"}}`,
			path:     "data",
			key:      "key-with-dash",
			wantErr:  false,
			expected: true,
		},
		{
			name:     "数组-错误",
			jsonStr:  `{"items":[1,2,3]}`,
			path:     "items",
			key:      "0",
			wantErr:  true,
			expected: false,
		},
		{
			name:     "字符串-错误",
			jsonStr:  `{"name":"Alice"}`,
			path:     "name",
			key:      "anything",
			wantErr:  true,
			expected: false,
		},
		{
			name:     "不存在的路径",
			jsonStr:  `{"name":"Alice"}`,
			path:     "age",
			key:      "anything",
			wantErr:  true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := j.Contains(tt.jsonStr, tt.path, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Contains() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// =========================================
// 综合测试
// =========================================

func TestJSONEngine_ComplexScenarios(t *testing.T) {
	j := newJSONEngine()

	t.Run("完整的API响应处理", func(t *testing.T) {
		apiResponse := `{
			"status": "success",
			"data": {
				"user": {
					"id": 123,
					"name": "Alice Johnson",
					"email": "alice@example.com",
					"profile": {
						"age": 30,
						"city": "Beijing",
						"tags": ["developer", "golang"]
					}
				},
				"meta": {
					"timestamp": 1234567890,
					"version": "1.0"
				}
			}
		}`

		// 测试IsValid
		if !j.IsValid(apiResponse) {
			t.Error("API response should be valid")
		}

		// 测试路径访问
		userName, err := j.GetString(apiResponse, "data.user.name")
		if err != nil || userName != "Alice Johnson" {
			t.Errorf("Expected 'Alice Johnson', got '%s', err: %v", userName, err)
		}

		userAge, err := j.GetFloat64(apiResponse, "data.user.profile.age")
		if err != nil || userAge != 30 {
			t.Errorf("Expected 30, got %v, err: %v", userAge, err)
		}

		// 测试GetType
		typeName, err := j.GetType(apiResponse, "data")
		if err != nil || typeName != "object" {
			t.Errorf("Expected 'object', got '%s', err: %v", typeName, err)
		}

		// 测试GetKeys
		keys, err := j.GetKeys(apiResponse, "data.user")
		if err != nil {
			t.Errorf("GetKeys error: %v", err)
		} else if !containsStringInSlice(keys, "name") {
			t.Error("Keys should contain 'name'")
		}
	})

	t.Run("配置文件合并", func(t *testing.T) {
		defaultConfig := `{
			"database": {
				"host": "localhost",
				"port": 3306,
				"ssl": false
			},
			"logging": {
				"level": "info",
				"format": "json"
			}
		}`

		userConfig := `{
			"database": {
				"host": "production.db.example.com",
				"password": "secret"
			},
			"logging": {
				"level": "debug"
			}
		}`

		merged, err := j.Merge(defaultConfig, userConfig)
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		// 验证合并结果
		host, _ := j.GetString(merged, "database.host")
		if host != "production.db.example.com" {
			t.Errorf("Expected production host, got %s", host)
		}

		port, _ := j.GetFloat64(merged, "database.port")
		if port != 3306 {
			t.Errorf("Expected port 3306, got %v", port)
		}

		// 验证嵌套合并 - ssl 应该保持默认值 false
		ssl, _ := j.GetBool(merged, "database.ssl")
		if ssl {
			t.Error("Expected ssl to be false from default")
		}

		// 验证密码被添加
		password, _ := j.GetString(merged, "database.password")
		if password != "secret" {
			t.Errorf("Expected password 'secret', got %s", password)
		}

		// 验证 logging 合并
		level, _ := j.GetString(merged, "logging.level")
		if level != "debug" {
			t.Errorf("Expected logging level 'debug', got %s", level)
		}

		format, _ := j.GetString(merged, "logging.format")
		if format != "json" {
			t.Errorf("Expected logging format 'json', got %s", format)
		}
	})

	t.Run("数据类型转换链", func(t *testing.T) {
		// Struct -> JSON -> Map
		person := TestPerson{
			Name:  "Bob",
			Age:   25,
			Email: "bob@example.com",
			Tags:  []string{"engineer", "python"},
		}
		person.Address.City = "Shanghai"
		person.Address.Street = "Nanjing Road"
		person.Address.Country = "China"

		// Struct to JSON
		jsonStr, err := j.FromStruct(person)
		if err != nil {
			t.Fatalf("FromStruct failed: %v", err)
		}

		// JSON to Map
		dataMap, err := j.ToMap(jsonStr)
		if err != nil {
			t.Fatalf("ToMap failed: %v", err)
		}

		// 验证map
		if dataMap["name"] != "Bob" {
			t.Error("Name should be Bob")
		}

		// Map back to JSON
		jsonStr2, err := j.FromMap(dataMap)
		if err != nil {
			t.Fatalf("FromMap failed: %v", err)
		}

		// Compare
		diff, err := j.Diff(jsonStr, jsonStr2)
		if err != nil {
			t.Fatalf("Diff failed: %v", err)
		}
		if diff {
			t.Error("Conversion should preserve data")
		}
	})
}

// =========================================
// 兼容性测试
// =========================================

func TestDefault(t *testing.T) {
	// 测试 Default() 函数
	// 注意：此函数已弃用，JSON 变量被标记为 removed
	j := Default()

	// 由于 JSON 变量未被初始化，返回的是 nil
	// 这是预期的行为，因为代码已被标记为 removed
	if j != nil {
		// 如果未来有人重新初始化了 JSON 变量，测试它是否工作
		testJSON := `{"name":"test","value":123}`
		if !j.IsValid(testJSON) {
			t.Error("Default() instance should work correctly when initialized")
		}
	}
	// nil 是可以接受的，因为这个功能已被标记为 removed
}

// =========================================
// 辅助函数
// =========================================

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOfString(s, substr) >= 0)
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// containsWhitespace 检查字符串是否包含空白字符
func containsWhitespace(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			return true
		}
	}
	return false
}

// containsStringInSlice 检查字符串切片是否包含指定字符串
func containsStringInSlice(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// deepEqual 比较两个值是否相等
func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for key, val := range av {
			if bVal, exists := bv[key]; !exists || !deepEqual(val, bVal) {
				return false
			}
		}
		return true
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !deepEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
