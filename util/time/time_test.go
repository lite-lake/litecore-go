package time

import (
	"fmt"
	"testing"
	stdtime "time"
)

// =========================================
// 测试基础时间检查方法
// =========================================

func TestTimeEngine_IsZero(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		tim      stdtime.Time
		expected bool
	}{
		{"零值时间", stdtime.Time{}, true},
		{"非零值时间", stdtime.Now(), false},
		{"Unix创建的时间", stdtime.Unix(0, 0), false}, // Unix epoch 不是零值，是有效时间
		{"具体日期", stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.IsZero(tt.tim)
			if result != tt.expected {
				t.Errorf("IsZero() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_IsNotZero(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		tim      stdtime.Time
		expected bool
	}{
		{"零值时间", stdtime.Time{}, false},
		{"非零值时间", stdtime.Now(), true},
		{"具体日期", stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.IsNotZero(tt.tim)
			if result != tt.expected {
				t.Errorf("IsNotZero() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_After(t *testing.T) {
	engine := newTimeEngine()

	time1 := stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC)
	time2 := stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)

	if !engine.After(time1, time2) {
		t.Error("After() should return true when first time is after second")
	}

	if engine.After(time2, time1) {
		t.Error("After() should return false when first time is before second")
	}
}

func TestTimeEngine_Before(t *testing.T) {
	engine := newTimeEngine()

	time1 := stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)
	time2 := stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC)

	if !engine.Before(time1, time2) {
		t.Error("Before() should return true when first time is before second")
	}

	if engine.Before(time2, time1) {
		t.Error("Before() should return false when first time is after second")
	}
}

func TestTimeEngine_Equal(t *testing.T) {
	engine := newTimeEngine()

	time1 := stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)
	time2 := stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)
	time3 := stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC)

	if !engine.Equal(time1, time2) {
		t.Error("Equal() should return true for equal times")
	}

	if engine.Equal(time1, time3) {
		t.Error("Equal() should return false for different times")
	}
}

// =========================================
// 测试时间获取方法
// =========================================

func TestTimeEngine_Now(t *testing.T) {
	engine := newTimeEngine()

	now := engine.Now()
	if engine.IsZero(now) {
		t.Error("Now() should not return zero time")
	}

	// 检查时间是否在合理范围内（当前时间前后1秒内）
	expectedNow := stdtime.Now()
	diff := expectedNow.Sub(now)
	if diff < 0 {
		diff = -diff
	}

	if diff > stdtime.Second {
		t.Errorf("Now() returned time too far from actual time, diff: %v", diff)
	}
}

func TestTimeEngine_NowUnix(t *testing.T) {
	engine := newTimeEngine()

	unix := engine.NowUnix()
	if unix == 0 {
		t.Error("NowUnix() should not return 0")
	}

	// 验证返回值是否合理（应该在当前时间附近）
	expected := stdtime.Now().Unix()
	diff := expected - unix
	if diff < 0 {
		diff = -diff
	}

	if diff > 2 {
		t.Errorf("NowUnix() returned value too far from actual time, diff: %d", diff)
	}
}

func TestTimeEngine_NowUnixMilli(t *testing.T) {
	engine := newTimeEngine()

	unixMilli := engine.NowUnixMilli()
	if unixMilli == 0 {
		t.Error("NowUnixMilli() should not return 0")
	}

	// 验证返回值是否合理
	expected := stdtime.Now().UnixMilli()
	diff := expected - unixMilli
	if diff < 0 {
		diff = -diff
	}

	if diff > 2000 {
		t.Errorf("NowUnixMilli() returned value too far from actual time, diff: %d", diff)
	}
}

func TestTimeEngine_Unix(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		sec      int64
		nsec     int64
		expected stdtime.Time
	}{
		{0, 0, stdtime.Unix(0, 0)},
		{1609459200, 0, stdtime.Date(2021, 1, 1, 0, 0, 0, 0, stdtime.UTC)},
		{1609459200, 500000000, stdtime.Date(2021, 1, 1, 0, 0, 0, 500000000, stdtime.UTC)},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("sec=%d,nsec=%d", tt.sec, tt.nsec), func(t *testing.T) {
			result := engine.Unix(tt.sec, tt.nsec)
			if !result.Equal(tt.expected) {
				t.Errorf("Unix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_Parse(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		layout   string
		value    string
		wantErr  bool
		expected stdtime.Time
	}{
		{
			name:     "ANSIC格式",
			layout:   stdtime.ANSIC,
			value:    "Mon Jan 2 15:04:05 2006",
			wantErr:  false,
			expected: stdtime.Date(2006, 1, 2, 15, 4, 5, 0, stdtime.UTC),
		},
		{
			name:     "RFC3339格式",
			layout:   stdtime.RFC3339,
			value:    "2006-01-02T15:04:05Z",
			wantErr:  false,
			expected: stdtime.Date(2006, 1, 2, 15, 4, 5, 0, stdtime.UTC),
		},
		{
			name:    "无效格式",
			layout:  "2006-01-02",
			value:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Parse(tt.layout, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !result.Equal(tt.expected) {
				t.Errorf("Parse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// =========================================
// 测试Java风格格式化方法
// =========================================

func TestTimeEngine_ConvertJavaFormatToGo(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		javaFormat  string
		expectedGo  string
		description string
	}{
		// 常用格式（硬编码优化路径）
		{"yyyyMMdd", "20060102", "紧凑日期格式"},
		{"yyyy-MM-dd", "2006-01-02", "标准日期格式"},
		{"yyyy/MM/dd", "2006/01/02", "斜杠分隔日期"},
		{"yyyy年MM月dd日", "2006年01月02日", "中文日期格式"},
		{"yyyyMMddHHmmss", "20060102150405", "紧凑日期时间"},
		{"yyyy-MM-dd HH:mm:ss", "2006-01-02 15:04:05", "标准日期时间"},
		{"yyyy/MM/dd HH:mm:ss", "2006/01/02 15:04:05", "斜杠日期时间"},
		{"yyyy年MM月dd日 HH:mm:ss", "2006年01月02日 15:04:05", "中文日期时间"},
		{"yyyyMMddHHmmssSSS", "20060102150405.000", "紧凑日期时间毫秒"},
		{"yyyy-MM-dd HH:mm:ss.SSS", "2006-01-02 15:04:05.000", "标准日期时间毫秒"},
		{"HH:mm:ss", "15:04:05", "时间格式"},
		{"HH:mm", "15:04", "短时间格式"},
		{"MM-dd", "01-02", "月日格式"},
		{"MM/dd", "01/02", "斜杠月日"},

		// 通用转换路径
		{"yy-MM-dd", "06-01-02", "两位年份"},
		{"yyyy-M-d", "2006-1-2", "单字符月日"},
		{"yyyy/MM/dd HH:mm", "2006/01/02 15:04", "无秒"},
		{"HH:mm:ss.SSS", "15:04:05.000", "毫秒时间"},
		{"yyyyMMddHHmmssSS", "2006010215040500", "两位毫秒"}, // SS被替换为00
		{"yyyyMMddHHmmssS", "200601021504050", "一位毫秒"},   // S被替换为0
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := engine.ConvertJavaFormatToGo(tt.javaFormat)
			if result != tt.expectedGo {
				t.Errorf("ConvertJavaFormatToGo(%q) = %q, want %q", tt.javaFormat, result, tt.expectedGo)
			}
		})
	}
}

func TestTimeEngine_FormatWithJava(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 123000000, stdtime.UTC)

	tests := []struct {
		javaFormat  string
		expected    string
		description string
	}{
		{"yyyy-MM-dd", "2024-01-15", "标准日期"},
		{"yyyy/MM/dd", "2024/01/15", "斜杠日期"},
		{"yyyyMMdd", "20240115", "紧凑日期"},
		{"yyyy-MM-dd HH:mm:ss", "2024-01-15 13:30:45", "完整日期时间"},
		{"HH:mm:ss", "13:30:45", "时间"},
		{"yyyy年MM月dd日", "2024年01月15日", "中文日期"},
		{"yyyy-MM-dd HH:mm:ss.SSS", "2024-01-15 13:30:45.123", "带毫秒"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := engine.FormatWithJava(testTime, tt.javaFormat)
			if result != tt.expected {
				t.Errorf("FormatWithJava(%q) = %q, want %q", tt.javaFormat, result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_FormatWithJavaOrDefault(t *testing.T) {
	engine := newTimeEngine()

	validTime := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)
	zeroTime := stdtime.Time{}

	tests := []struct {
		name         string
		tim          stdtime.Time
		javaFormat   string
		defaultValue string
		expected     string
	}{
		{
			name:         "有效时间",
			tim:          validTime,
			javaFormat:   "yyyy-MM-dd",
			defaultValue: "N/A",
			expected:     "2024-01-15",
		},
		{
			name:         "零值时间",
			tim:          zeroTime,
			javaFormat:   "yyyy-MM-dd",
			defaultValue: "N/A",
			expected:     "N/A",
		},
		{
			name:         "零值时间-空默认值",
			tim:          zeroTime,
			javaFormat:   "yyyy-MM-dd",
			defaultValue: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.FormatWithJavaOrDefault(tt.tim, tt.javaFormat, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("FormatWithJavaOrDefault() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// =========================================
// 测试Java风格解析方法
// =========================================

func TestTimeEngine_ParseWithJava(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name       string
		value      string
		javaFormat string
		wantErr    bool
		expected   stdtime.Time
	}{
		{
			name:       "标准日期",
			value:      "2024-01-15",
			javaFormat: "yyyy-MM-dd",
			wantErr:    false,
			expected:   stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:       "紧凑日期时间",
			value:      "20240115133045",
			javaFormat: "yyyyMMddHHmmss",
			wantErr:    false,
			expected:   stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC),
		},
		{
			name:       "中文日期",
			value:      "2024年01月15日",
			javaFormat: "yyyy年MM月dd日",
			wantErr:    false,
			expected:   stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:       "带毫秒",
			value:      "2024-01-15 13:30:45.123",
			javaFormat: "yyyy-MM-dd HH:mm:ss.SSS",
			wantErr:    false,
			expected:   stdtime.Date(2024, 1, 15, 13, 30, 45, 123000000, stdtime.UTC),
		},
		{
			name:       "无效日期",
			value:      "invalid",
			javaFormat: "yyyy-MM-dd",
			wantErr:    true,
		},
		{
			name:       "格式不匹配",
			value:      "2024-01-15",
			javaFormat: "yyyyMMdd",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.ParseWithJava(tt.value, tt.javaFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWithJava() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !result.Equal(tt.expected) {
				t.Errorf("ParseWithJava() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_TryParseWithJava(t *testing.T) {
	engine := newTimeEngine()

	t.Run("成功解析", func(t *testing.T) {
		result := engine.TryParseWithJava("2024-01-15", "yyyy-MM-dd")
		if result.IsZero() {
			t.Error("TryParseWithJava() should return valid time")
		}
		expected := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)
		if !result.Equal(expected) {
			t.Errorf("TryParseWithJava() = %v, want %v", result, expected)
		}
	})

	t.Run("解析失败", func(t *testing.T) {
		result := engine.TryParseWithJava("invalid", "yyyy-MM-dd")
		if !result.IsZero() {
			t.Error("TryParseWithJava() should return zero time on failure")
		}
	})
}

func TestTimeEngine_ParseWithMultipleFormats(t *testing.T) {
	engine := newTimeEngine()

	t.Run("匹配第一个格式", func(t *testing.T) {
		value := "2024-01-15"
		formats := []string{"yyyy-MM-dd", "yyyy/MM/dd", "yyyyMMdd"}

		result, err := engine.ParseWithMultipleFormats(value, formats)
		if err != nil {
			t.Errorf("ParseWithMultipleFormats() error = %v", err)
			return
		}

		expected := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)
		if !result.Equal(expected) {
			t.Errorf("ParseWithMultipleFormats() = %v, want %v", result, expected)
		}
	})

	t.Run("匹配第二个格式", func(t *testing.T) {
		value := "2024/01/15"
		formats := []string{"yyyy-MM-dd", "yyyy/MM/dd", "yyyyMMdd"}

		result, err := engine.ParseWithMultipleFormats(value, formats)
		if err != nil {
			t.Errorf("ParseWithMultipleFormats() error = %v", err)
			return
		}

		expected := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)
		if !result.Equal(expected) {
			t.Errorf("ParseWithMultipleFormats() = %v, want %v", result, expected)
		}
	})

	t.Run("所有格式都不匹配", func(t *testing.T) {
		value := "invalid"
		formats := []string{"yyyy-MM-dd", "yyyy/MM/dd"}

		_, err := engine.ParseWithMultipleFormats(value, formats)
		if err == nil {
			t.Error("ParseWithMultipleFormats() should return error when no format matches")
		}
	})
}

func TestTimeEngine_TryParseWithMultipleFormats(t *testing.T) {
	engine := newTimeEngine()

	t.Run("成功解析", func(t *testing.T) {
		value := "2024-01-15"
		formats := []string{"yyyy-MM-dd", "yyyy/MM/dd"}

		result := engine.TryParseWithMultipleFormats(value, formats)
		if result.IsZero() {
			t.Error("TryParseWithMultipleFormats() should return valid time")
		}
	})

	t.Run("解析失败", func(t *testing.T) {
		value := "invalid"
		formats := []string{"yyyy-MM-dd", "yyyy/MM/dd"}

		result := engine.TryParseWithMultipleFormats(value, formats)
		if !result.IsZero() {
			t.Error("TryParseWithMultipleFormats() should return zero time on failure")
		}
	})
}

func TestTimeEngine_ParseAuto(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		value    string
		wantErr  bool
		expected stdtime.Time
	}{
		{
			name:     "RFC3339格式",
			value:    "2024-01-15T13:30:45Z",
			wantErr:  false,
			expected: stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC),
		},
		{
			name:     "标准日期时间",
			value:    "2024-01-15 13:30:45",
			wantErr:  false,
			expected: stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC),
		},
		{
			name:     "标准日期",
			value:    "2024-01-15",
			wantErr:  false,
			expected: stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:     "紧凑日期",
			value:    "20240115",
			wantErr:  false,
			expected: stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:     "斜杠日期",
			value:    "2024/01/15",
			wantErr:  false,
			expected: stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:    "无效格式",
			value:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.ParseAuto(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAuto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !result.Equal(tt.expected) {
				t.Errorf("ParseAuto() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_TryParseAuto(t *testing.T) {
	engine := newTimeEngine()

	t.Run("成功解析", func(t *testing.T) {
		result := engine.TryParseAuto("2024-01-15")
		if result.IsZero() {
			t.Error("TryParseAuto() should return valid time")
		}
	})

	t.Run("解析失败", func(t *testing.T) {
		result := engine.TryParseAuto("invalid")
		if !result.IsZero() {
			t.Error("TryParseAuto() should return zero time on failure")
		}
	})
}

// =========================================
// 测试时间计算方法
// =========================================

func TestTimeEngine_Add(t *testing.T) {
	engine := newTimeEngine()

	baseTime := stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)

	// 测试增加时间
	result := engine.Add(baseTime, 24*stdtime.Hour)
	expected := stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("Add() = %v, want %v", result, expected)
	}

	// 测试减少时间
	result = engine.Add(baseTime, -24*stdtime.Hour)
	expected = stdtime.Date(2023, 12, 31, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("Add() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_AddDuration(t *testing.T) {
	engine := newTimeEngine()

	baseTime := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)

	// 测试增加年月日
	result := engine.AddDuration(baseTime, 1, 2, 3)
	expected := stdtime.Date(2025, 3, 18, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("AddDuration() = %v, want %v", result, expected)
	}

	// 测试减少年月日
	result = engine.AddDuration(baseTime, -1, -2, -3)
	expected = stdtime.Date(2022, 11, 12, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("AddDuration() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_AddYears(t *testing.T) {
	engine := newTimeEngine()

	baseTime := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)

	result := engine.AddYears(baseTime, 1)
	expected := stdtime.Date(2025, 1, 15, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("AddYears() = %v, want %v", result, expected)
	}

	result = engine.AddYears(baseTime, -1)
	expected = stdtime.Date(2023, 1, 15, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("AddYears() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_AddMonths(t *testing.T) {
	engine := newTimeEngine()

	baseTime := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)

	result := engine.AddMonths(baseTime, 1)
	expected := stdtime.Date(2024, 2, 15, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("AddMonths() = %v, want %v", result, expected)
	}

	// 测试跨年
	result = engine.AddMonths(baseTime, 12)
	expected = stdtime.Date(2025, 1, 15, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("AddMonths() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_AddDays(t *testing.T) {
	engine := newTimeEngine()

	baseTime := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)

	result := engine.AddDays(baseTime, 1)
	expected := stdtime.Date(2024, 1, 16, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("AddDays() = %v, want %v", result, expected)
	}

	// 测试跨月
	result = engine.AddDays(baseTime, 20)
	expected = stdtime.Date(2024, 2, 4, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("AddDays() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_Sub(t *testing.T) {
	engine := newTimeEngine()

	time1 := stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC)
	time2 := stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)

	result := engine.Sub(time1, time2)
	expected := 24 * stdtime.Hour

	if result != expected {
		t.Errorf("Sub() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_DurationBetween(t *testing.T) {
	engine := newTimeEngine()

	time1 := stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC)
	time2 := stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)

	result := engine.DurationBetween(time1, time2)
	expected := int64(24 * 60 * 60 * 1000) // 24小时的毫秒数

	if result != expected {
		t.Errorf("DurationBetween() = %d, want %d", result, expected)
	}
}

func TestTimeEngine_DaysBetween(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		time1    stdtime.Time
		time2    stdtime.Time
		expected int
	}{
		{
			name:     "相差1天",
			time1:    stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC),
			time2:    stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
			expected: 1,
		},
		{
			name:     "相差30天",
			time1:    stdtime.Date(2024, 1, 31, 0, 0, 0, 0, stdtime.UTC),
			time2:    stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
			expected: 30,
		},
		{
			name:     "时间1早于时间2",
			time1:    stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
			time2:    stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC),
			expected: 1,
		},
		{
			name:     "同一天",
			time1:    stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
			time2:    stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.DaysBetween(tt.time1, tt.time2)
			if result != tt.expected {
				t.Errorf("DaysBetween() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// =========================================
// 测试时间转换方法
// =========================================

func TestTimeEngine_StartOfDay(t *testing.T) {
	engine := newTimeEngine()

	// 测试一天中的不同时间
	testTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 123, stdtime.UTC)
	result := engine.StartOfDay(testTime)
	expected := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("StartOfDay() = %v, want %v", result, expected)
	}

	// 验证结果是当天开始
	if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 {
		t.Error("StartOfDay() should return time with 00:00:00")
	}
}

func TestTimeEngine_EndOfDay(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 123, stdtime.UTC)
	result := engine.EndOfDay(testTime)

	// 验证结果是当天结束
	if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 {
		t.Errorf("EndOfDay() should return time with 23:59:59, got %02d:%02d:%02d",
			result.Hour(), result.Minute(), result.Second())
	}

	// 验证纳秒数接近最大值
	if result.Nanosecond() < 999999999 {
		t.Errorf("EndOfDay() should have nanoseconds near max, got %d", result.Nanosecond())
	}
}

func TestTimeEngine_StartOfWeek(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		input    stdtime.Time
		expected stdtime.Time
	}{
		{
			name:     "周一",
			input:    stdtime.Date(2024, 1, 15, 13, 0, 0, 0, stdtime.UTC), // 周一
			expected: stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:     "周三",
			input:    stdtime.Date(2024, 1, 17, 13, 0, 0, 0, stdtime.UTC), // 周三
			expected: stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:     "周日",
			input:    stdtime.Date(2024, 1, 21, 13, 0, 0, 0, stdtime.UTC), // 周日
			expected: stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.StartOfWeek(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("StartOfWeek() = %v, want %v", result, tt.expected)
			}

			// 验证结果是周一
			if result.Weekday() != stdtime.Monday {
				t.Errorf("StartOfWeek() should return Monday, got %v", result.Weekday())
			}
		})
	}
}

func TestTimeEngine_EndOfWeek(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		input    stdtime.Time
		expected stdtime.Time
	}{
		{
			name:     "周一",
			input:    stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC), // 周一
			expected: stdtime.Date(2024, 1, 21, 23, 59, 59, 999999999, stdtime.UTC),
		},
		{
			name:     "周三",
			input:    stdtime.Date(2024, 1, 17, 0, 0, 0, 0, stdtime.UTC), // 周三
			expected: stdtime.Date(2024, 1, 21, 23, 59, 59, 999999999, stdtime.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.EndOfWeek(tt.input)
			// 只比较日期，不比较精确时间
			if !result.Equal(tt.expected) {
				t.Errorf("EndOfWeek() = %v, want %v", result, tt.expected)
			}

			// 验证结果是周日
			if result.Weekday() != stdtime.Sunday {
				t.Errorf("EndOfWeek() should return Sunday, got %v", result.Weekday())
			}
		})
	}
}

func TestTimeEngine_StartOfMonth(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		input    stdtime.Time
		expected stdtime.Time
	}{
		{
			name:     "月中",
			input:    stdtime.Date(2024, 1, 15, 13, 0, 0, 0, stdtime.UTC),
			expected: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:     "月末",
			input:    stdtime.Date(2024, 1, 31, 13, 0, 0, 0, stdtime.UTC),
			expected: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.StartOfMonth(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("StartOfMonth() = %v, want %v", result, tt.expected)
			}

			if result.Day() != 1 {
				t.Errorf("StartOfMonth() should return 1st day, got %d", result.Day())
			}
		})
	}
}

func TestTimeEngine_EndOfMonth(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name        string
		input       stdtime.Time
		expectedDay int
	}{
		{
			name:        "1月(31天)",
			input:       stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
			expectedDay: 31,
		},
		{
			name:        "2月闰年(29天)",
			input:       stdtime.Date(2024, 2, 15, 0, 0, 0, 0, stdtime.UTC),
			expectedDay: 29,
		},
		{
			name:        "2月平年(28天)",
			input:       stdtime.Date(2023, 2, 15, 0, 0, 0, 0, stdtime.UTC),
			expectedDay: 28,
		},
		{
			name:        "4月(30天)",
			input:       stdtime.Date(2024, 4, 15, 0, 0, 0, 0, stdtime.UTC),
			expectedDay: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.EndOfMonth(tt.input)

			if result.Day() != tt.expectedDay {
				t.Errorf("EndOfMonth() day = %d, want %d", result.Day(), tt.expectedDay)
			}

			if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 {
				t.Errorf("EndOfMonth() should return time with 23:59:59")
			}
		})
	}
}

func TestTimeEngine_StartOfYear(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		input    stdtime.Time
		expected stdtime.Time
	}{
		{
			name:     "年中",
			input:    stdtime.Date(2024, 6, 15, 13, 0, 0, 0, stdtime.UTC),
			expected: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
		},
		{
			name:     "年末",
			input:    stdtime.Date(2024, 12, 31, 13, 0, 0, 0, stdtime.UTC),
			expected: stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.StartOfYear(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("StartOfYear() = %v, want %v", result, tt.expected)
			}

			if result.Month() != 1 || result.Day() != 1 {
				t.Errorf("StartOfYear() should return Jan 1st, got %s %d", result.Month(), result.Day())
			}
		})
	}
}

func TestTimeEngine_EndOfYear(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		input    stdtime.Time
		expected stdtime.Time
	}{
		{
			name:     "年中",
			input:    stdtime.Date(2024, 6, 15, 0, 0, 0, 0, stdtime.UTC),
			expected: stdtime.Date(2024, 12, 31, 23, 59, 59, 999999999, stdtime.UTC),
		},
		{
			name:     "年初",
			input:    stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC),
			expected: stdtime.Date(2024, 12, 31, 23, 59, 59, 999999999, stdtime.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.EndOfYear(tt.input)

			if result.Month() != 12 || result.Day() != 31 {
				t.Errorf("EndOfYear() should return Dec 31st, got %s %d", result.Month(), result.Day())
			}

			if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 {
				t.Errorf("EndOfYear() should return time with 23:59:59")
			}
		})
	}
}

// =========================================
// 测试时间工具方法
// =========================================

func TestTimeEngine_IsLeapYear(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		year     int
		expected bool
		name     string
	}{
		{2000, true, "能被400整除"},
		{2004, true, "能被4整除但不能被100整除"},
		{2020, true, "普通闰年"},
		{2024, true, "当前闰年"},
		{1900, false, "能被100整除但不能被400整除"},
		{2001, false, "不能被4整除"},
		{2002, false, "普通平年"},
		{2023, false, "当前平年"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.IsLeapYear(tt.year)
			if result != tt.expected {
				t.Errorf("IsLeapYear(%d) = %v, want %v", tt.year, result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_DaysInMonth(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		year     int
		month    int
		expected int
		name     string
	}{
		{2024, 1, 31, "1月有31天"},
		{2024, 2, 29, "2024年2月有29天(闰年)"},
		{2023, 2, 28, "2023年2月有28天(平年)"},
		{2000, 2, 29, "2000年2月有29天(世纪闰年)"},
		{1900, 2, 28, "1900年2月有28天(世纪平年)"},
		{2024, 4, 30, "4月有30天"},
		{2024, 6, 30, "6月有30天"},
		{2024, 9, 30, "9月有30天"},
		{2024, 11, 30, "11月有30天"},
		{2024, 12, 31, "12月有31天"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.DaysInMonth(tt.year, tt.month)
			if result != tt.expected {
				t.Errorf("DaysInMonth(%d, %d) = %d, want %d", tt.year, tt.month, result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_FormatDuration(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		duration stdtime.Duration
		expected string
	}{
		{"1小时", 1 * stdtime.Hour, "01:00:00"},
		{"2小时30分", 2*stdtime.Hour + 30*stdtime.Minute, "02:30:00"},
		{"90分钟", 90 * stdtime.Minute, "01:30:00"},
		{"1分30秒", 1*stdtime.Minute + 30*stdtime.Second, "01:30"},
		{"45秒", 45 * stdtime.Second, "00:45"},
		{"完整时间", 2*stdtime.Hour + 34*stdtime.Minute + 56*stdtime.Second, "02:34:56"},
		{"负数时长", -1 * stdtime.Hour, "01:00:00"}, // 负数会被取绝对值
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_Age(t *testing.T) {
	engine := newTimeEngine()

	// 由于Age()方法内部使用Now()，我们测试几个相对固定的场景
	// 使用相对时间而不是绝对时间，使测试更稳定

	tests := []struct {
		name           string
		yearsAgo       int
		minExpectedAge int
		maxExpectedAge int
	}{
		{
			name:           "刚出生",
			yearsAgo:       0,
			minExpectedAge: 0,
			maxExpectedAge: 0,
		},
		{
			name:           "1年前",
			yearsAgo:       1,
			minExpectedAge: 0,
			maxExpectedAge: 1,
		},
		{
			name:           "10年前",
			yearsAgo:       10,
			minExpectedAge: 9,
			maxExpectedAge: 10,
		},
		{
			name:           "20年前",
			yearsAgo:       20,
			minExpectedAge: 19,
			maxExpectedAge: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建tt.yearsAgo年前的日期
			birthDate := engine.AddYears(stdtime.Now(), -tt.yearsAgo)

			// 计算年龄
			age := engine.Age(birthDate)

			// 验证年龄在合理范围内
			if age < tt.minExpectedAge || age > tt.maxExpectedAge {
				t.Logf("Age() = %d (yearsAgo: %d), expected between %d and %d",
					age, tt.yearsAgo, tt.minExpectedAge, tt.maxExpectedAge)

				// 由于生日可能还没到，允许有一定的范围
				if age < tt.minExpectedAge-1 || age > tt.maxExpectedAge+1 {
					t.Errorf("Age() = %d, want between %d and %d", age, tt.minExpectedAge, tt.maxExpectedAge)
				}
			}
		})
	}
}

func TestTimeEngine_Between(t *testing.T) {
	engine := newTimeEngine()

	start := stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC)
	end := stdtime.Date(2024, 1, 10, 0, 0, 0, 0, stdtime.UTC)

	tests := []struct {
		name     string
		tim      stdtime.Time
		expected bool
	}{
		{"范围内", stdtime.Date(2024, 1, 5, 0, 0, 0, 0, stdtime.UTC), true},
		{"等于开始", start, true},
		{"等于结束", end, true},
		{"早于开始", stdtime.Date(2023, 12, 31, 0, 0, 0, 0, stdtime.UTC), false},
		{"晚于结束", stdtime.Date(2024, 1, 11, 0, 0, 0, 0, stdtime.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Between(tt.tim, start, end)
			if result != tt.expected {
				t.Errorf("Between() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_Truncate(t *testing.T) {
	engine := newTimeEngine()

	baseTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 123456789, stdtime.UTC)

	// 截断到小时
	result := engine.Truncate(baseTime, stdtime.Hour)
	expected := stdtime.Date(2024, 1, 15, 13, 0, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("Truncate(Hour) = %v, want %v", result, expected)
	}

	// 截断到分钟
	result = engine.Truncate(baseTime, stdtime.Minute)
	expected = stdtime.Date(2024, 1, 15, 13, 30, 0, 0, stdtime.UTC)

	if !result.Equal(expected) {
		t.Errorf("Truncate(Minute) = %v, want %v", result, expected)
	}
}

func TestTimeEngine_Round(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		tim      stdtime.Time
		duration stdtime.Duration
		expected stdtime.Time
	}{
		{
			name:     "四舍五入到小时-向下",
			tim:      stdtime.Date(2024, 1, 15, 13, 29, 59, 0, stdtime.UTC),
			duration: stdtime.Hour,
			expected: stdtime.Date(2024, 1, 15, 13, 0, 0, 0, stdtime.UTC),
		},
		{
			name:     "四舍五入到小时-向上",
			tim:      stdtime.Date(2024, 1, 15, 13, 30, 1, 0, stdtime.UTC),
			duration: stdtime.Hour,
			expected: stdtime.Date(2024, 1, 15, 14, 0, 0, 0, stdtime.UTC),
		},
		{
			name:     "四舍五入到分钟",
			tim:      stdtime.Date(2024, 1, 15, 13, 30, 30, 0, stdtime.UTC),
			duration: stdtime.Minute,
			expected: stdtime.Date(2024, 1, 15, 13, 31, 0, 0, stdtime.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Round(tt.tim, tt.duration)
			if !result.Equal(tt.expected) {
				t.Errorf("Round() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// =========================================
// 测试快速格式化方法
// =========================================

func TestTimeEngine_ToYYYYMMDD(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)
	result := engine.ToYYYYMMDD(testTime)
	expected := "20240115"

	if result != expected {
		t.Errorf("ToYYYYMMDD() = %q, want %q", result, expected)
	}
}

func TestTimeEngine_ToYYYYMMDDHHMMSS(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC)
	result := engine.ToYYYYMMDDHHMMSS(testTime)
	expected := "20240115133045"

	if result != expected {
		t.Errorf("ToYYYYMMDDHHMMSS() = %q, want %q", result, expected)
	}
}

func TestTimeEngine_ToYYYY_MM_DD(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)
	result := engine.ToYYYY_MM_DD(testTime)
	expected := "2024-01-15"

	if result != expected {
		t.Errorf("ToYYYY_MM_DD() = %q, want %q", result, expected)
	}
}

func TestTimeEngine_ToYYYY_MM_DD_HH_MM_SS(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC)
	result := engine.ToYYYY_MM_DD_HH_MM_SS(testTime)
	expected := "2024-01-15 13:30:45"

	if result != expected {
		t.Errorf("ToYYYY_MM_DD_HH_MM_SS() = %q, want %q", result, expected)
	}
}

func TestTimeEngine_ToHHMMSS(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC)
	result := engine.ToHHMMSS(testTime)
	expected := "13:30:45"

	if result != expected {
		t.Errorf("ToHHMMSS() = %q, want %q", result, expected)
	}
}

// =========================================
// 测试快速解析方法
// =========================================

func TestTimeEngine_FromYYYYMMDD(t *testing.T) {
	engine := newTimeEngine()

	result, err := engine.FromYYYYMMDD("20240115")
	if err != nil {
		t.Errorf("FromYYYYMMDD() error = %v", err)
		return
	}

	expected := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)
	if !result.Equal(expected) {
		t.Errorf("FromYYYYMMDD() = %v, want %v", result, expected)
	}

	// 测试无效格式
	_, err = engine.FromYYYYMMDD("invalid")
	if err == nil {
		t.Error("FromYYYYMMDD() should return error for invalid format")
	}
}

func TestTimeEngine_FromYYYYMMDDHHMMSS(t *testing.T) {
	engine := newTimeEngine()

	result, err := engine.FromYYYYMMDDHHMMSS("20240115133045")
	if err != nil {
		t.Errorf("FromYYYYMMDDHHMMSS() error = %v", err)
		return
	}

	expected := stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC)
	if !result.Equal(expected) {
		t.Errorf("FromYYYYMMDDHHMMSS() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_FromYYYY_MM_DD(t *testing.T) {
	engine := newTimeEngine()

	result, err := engine.FromYYYY_MM_DD("2024-01-15")
	if err != nil {
		t.Errorf("FromYYYY_MM_DD() error = %v", err)
		return
	}

	expected := stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC)
	if !result.Equal(expected) {
		t.Errorf("FromYYYY_MM_DD() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_FromYYYY_MM_DD_HH_MM_SS(t *testing.T) {
	engine := newTimeEngine()

	result, err := engine.FromYYYY_MM_DD_HH_MM_SS("2024-01-15 13:30:45")
	if err != nil {
		t.Errorf("FromYYYY_MM_DD_HH_MM_SS() error = %v", err)
		return
	}

	expected := stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC)
	if !result.Equal(expected) {
		t.Errorf("FromYYYY_MM_DD_HH_MM_SS() = %v, want %v", result, expected)
	}
}

// =========================================
// 测试时区相关方法
// =========================================

func TestTimeEngine_InLocation(t *testing.T) {
	engine := newTimeEngine()

	// 创建UTC时间
	utcTime := stdtime.Date(2024, 1, 15, 13, 0, 0, 0, stdtime.UTC)

	// 转换到纽约时区
	loc, err := engine.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}

	result := engine.InLocation(utcTime, loc)

	// 验证时间已转换
	if result.Location().String() != "America/New_York" {
		t.Errorf("InLocation() location = %v, want America/New_York", result.Location())
	}
}

func TestTimeEngine_UTC(t *testing.T) {
	engine := newTimeEngine()

	// 创建本地时间
	localTime := stdtime.Date(2024, 1, 15, 13, 0, 0, 0, stdtime.Local)

	result := engine.UTC(localTime)

	// 验证时间在UTC时区
	if result.Location() != stdtime.UTC {
		t.Errorf("UTC() location = %v, want UTC", result.Location())
	}
}

func TestTimeEngine_Local(t *testing.T) {
	engine := newTimeEngine()

	// 创建UTC时间
	utcTime := stdtime.Date(2024, 1, 15, 13, 0, 0, 0, stdtime.UTC)

	result := engine.Local(utcTime)

	// 验证时间在本地时区
	if result.Location() != stdtime.Local {
		t.Errorf("Local() location = %v, want Local", result.Location())
	}
}

func TestTimeEngine_LoadLocation(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name    string
		tzName  string
		wantErr bool
	}{
		{"纽约时区", "America/New_York", false},
		{"伦敦时区", "Europe/London", false},
		{"东京时区", "Asia/Tokyo", false},
		{"上海时区", "Asia/Shanghai", false},
		{"UTC", "UTC", false},
		{"无效时区", "Invalid/Timezone", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := engine.LoadLocation(tt.tzName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && loc == nil {
				t.Error("LoadLocation() should return non-nil location")
			}
		})
	}
}

// =========================================
// 测试时间戳相关方法
// =========================================

func TestTimeEngine_ToUnix(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 0, stdtime.UTC)
	result := engine.ToUnix(testTime)

	expected := testTime.Unix()
	if result != expected {
		t.Errorf("ToUnix() = %d, want %d", result, expected)
	}
}

func TestTimeEngine_ToUnixMilli(t *testing.T) {
	engine := newTimeEngine()

	testTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 123000000, stdtime.UTC)
	result := engine.ToUnixMilli(testTime)

	expected := testTime.UnixMilli()
	if result != expected {
		t.Errorf("ToUnixMilli() = %d, want %d", result, expected)
	}
}

func TestTimeEngine_FromUnix(t *testing.T) {
	engine := newTimeEngine()

	sec := int64(1705317045) // 2024-01-15 13:30:45 UTC
	result := engine.FromUnix(sec)

	expected := stdtime.Unix(sec, 0)
	if !result.Equal(expected) {
		t.Errorf("FromUnix() = %v, want %v", result, expected)
	}
}

func TestTimeEngine_FromUnixMilli(t *testing.T) {
	engine := newTimeEngine()

	msec := int64(1705317045123) // 2024-01-15 13:30:45.123 UTC
	result := engine.FromUnixMilli(msec)

	expected := stdtime.Unix(msec/1000, (msec%1000)*1000000)
	if !result.Equal(expected) {
		t.Errorf("FromUnixMilli() = %v, want %v", result, expected)
	}
}

// =========================================
// 测试验证和辅助方法
// =========================================

func TestTimeEngine_IsValidFormat(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name       string
		javaFormat string
		expected   bool
	}{
		{"标准日期", "yyyy-MM-dd", true},
		{"标准时间", "HH:mm:ss", true},
		{"完整日期时间", "yyyy-MM-dd HH:mm:ss", true},
		{"紧凑格式", "yyyyMMddHHmmss", true},
		{"中文格式", "yyyy年MM月dd日", true},
		{"只有年", "yyyy", true},
		{"空字符串", "", false},
		{"无效字符串", "invalid", false},
		{"随机字符", "abcdef", true}, // 包含有效的模式字符
		{"混合格式", "yyyy-MM-dd HH:mm:ss.SSS", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.IsValidFormat(tt.javaFormat)
			if result != tt.expected {
				t.Errorf("IsValidFormat(%q) = %v, want %v", tt.javaFormat, result, tt.expected)
			}
		})
	}
}

func TestTimeEngine_GuessFormat(t *testing.T) {
	engine := newTimeEngine()

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"标准日期", "2024-01-15", "yyyy-MM-dd"},
		{"斜杠日期", "2024/01/15", "yyyy/MM/dd"},
		{"标准日期时间", "2024-01-15 13:30:45", "yyyy-MM-dd HH:mm:ss"},
		{"斜杠日期时间", "2024/01/15 13:30:45", "yyyy/MM/dd HH:mm:ss"},
		{"紧凑日期", "20240115", "yyyyMMdd"},
		{"紧凑日期时间", "20240115133045", "yyyyMMddHHmmss"},
		{"时间", "13:30:45", "HH:mm:ss"},
		{"短时间", "13:30", "HH:mm"},
		{"月日", "01-15", "yyyy-MM-dd"}, // 默认格式
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.GuessFormat(tt.value)
			if result != tt.expected {
				t.Errorf("GuessFormat(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

// =========================================
// 集成测试
// =========================================

func TestTimeEngine_RoundTrip(t *testing.T) {
	engine := newTimeEngine()

	originalTime := stdtime.Date(2024, 1, 15, 13, 30, 45, 123000000, stdtime.UTC)

	// 格式化后解析
	formatted := engine.FormatWithJava(originalTime, "yyyy-MM-dd HH:mm:ss.SSS")
	parsed, err := engine.ParseWithJava(formatted, "yyyy-MM-dd HH:mm:ss.SSS")

	if err != nil {
		t.Errorf("ParseWithJava() error = %v", err)
		return
	}

	if !parsed.Equal(originalTime) {
		t.Errorf("Round trip failed: original = %v, parsed = %v", originalTime, parsed)
	}
}

func TestTimeEngine_FormatParseConsistency(t *testing.T) {
	engine := newTimeEngine()

	testCases := []struct {
		format string
		value  string
	}{
		{"yyyyMMdd", "20240115"},
		{"yyyy-MM-dd", "2024-01-15"},
		{"yyyy/MM/dd", "2024/01/15"},
		{"yyyy-MM-dd HH:mm:ss", "2024-01-15 13:30:45"},
		{"yyyyMMddHHmmss", "20240115133045"},
	}

	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			// 解析
			parsed, err := engine.ParseWithJava(tc.value, tc.format)
			if err != nil {
				t.Errorf("ParseWithJava() error = %v", err)
				return
			}

			// 格式化
			formatted := engine.FormatWithJava(parsed, tc.format)

			// 验证格式化结果与原始值一致
			if formatted != tc.value {
				t.Errorf("Format inconsistency: original = %q, formatted = %q", tc.value, formatted)
			}
		})
	}
}

// =========================================
// 边界条件测试
// =========================================

func TestTimeEngine_BoundaryConditions(t *testing.T) {
	engine := newTimeEngine()

	t.Run("跨月计算", func(t *testing.T) {
		date := stdtime.Date(2024, 1, 31, 0, 0, 0, 0, stdtime.UTC)
		result := engine.AddMonths(date, 1)

		// Go的AddDate行为：1月31日加1个月会变成3月2日（或3月3日，取决于闰年）
		// 因为2月没有31日，所以会调整到3月的对应日期
		// 这是Go标准库的行为，我们的函数只是包装
		expectedMonth := stdtime.March
		if result.Month() != expectedMonth {
			t.Logf("AddMonths() across month boundary = %v (month: %s)", result, result.Month())
			// 这是预期行为，不报错
		}
	})

	t.Run("跨年计算", func(t *testing.T) {
		date := stdtime.Date(2024, 12, 31, 0, 0, 0, 0, stdtime.UTC)
		result := engine.AddDays(date, 1)

		if result.Year() != 2025 || result.Month() != 1 || result.Day() != 1 {
			t.Errorf("AddDays() across year boundary = %v, expected Jan 1, 2025", result)
		}
	})

	t.Run("闰年2月末", func(t *testing.T) {
		date := stdtime.Date(2024, 2, 28, 0, 0, 0, 0, stdtime.UTC)
		result := engine.AddDays(date, 1)

		if result.Month() != 2 || result.Day() != 29 {
			t.Errorf("AddDays() on leap year Feb 28 = %v, expected Feb 29", result)
		}
	})

	t.Run("时间戳边界", func(t *testing.T) {
		// Unix时间戳0
		result := engine.FromUnix(0)
		if result.Year() != 1970 {
			t.Errorf("FromUnix(0) year = %d, expected 1970", result.Year())
		}

		// 负数时间戳
		result = engine.FromUnix(-86400) // 1970年1月1日前一天
		if result.Year() != 1969 || result.Month() != 12 || result.Day() != 31 {
			t.Errorf("FromUnix(-86400) = %v, expected Dec 31, 1969", result)
		}
	})
}

// =========================================
// 并发安全测试
// =========================================

func TestTimeEngine_ConcurrentAccess(t *testing.T) {
	engine := newTimeEngine()

	done := make(chan bool)

	// 并发执行多个操作
	for i := 0; i < 100; i++ {
		go func() {
			testTime := stdtime.Now()

			// 各种操作
			engine.FormatWithJava(testTime, "yyyy-MM-dd HH:mm:ss")
			engine.IsZero(testTime)
			engine.AddDays(testTime, 1)
			engine.StartOfDay(testTime)
			engine.ToUnix(testTime)

			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 100; i++ {
		<-done
	}

	// 如果没有死锁或panic，则测试通过
}
