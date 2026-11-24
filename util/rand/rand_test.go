package rand

import (
	"regexp"
	"strings"
	"testing"
)

// 获取随机数工具实例
func getRandEngine() ILiteUtilRand {
	return Rand
}

// TestRandomInt 测试随机整数生成
func TestRandomInt(t *testing.T) {
	rand := getRandEngine()

	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"正常范围", 1, 100},
		{"相同值", 50, 50},
		{"反转范围", 100, 1},
		{"负数范围", -100, -1},
		{"混合范围", -50, 50},
		{"零范围", 0, 0},
		{"大范围", 0, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rand.RandomInt(tt.min, tt.max)

			// 确定实际的最小值和最大值
			actualMin, actualMax := tt.min, tt.max
			if actualMin > actualMax {
				actualMin, actualMax = actualMax, actualMin
			}

			// 验证结果在范围内
			if result < actualMin || result > actualMax {
				t.Errorf("RandomInt(%d, %d) = %d, 结果超出范围 [%d, %d]",
					tt.min, tt.max, result, actualMin, actualMax)
			}

			// 如果 min == max，结果应该等于该值
			if tt.min == tt.max && result != tt.min {
				t.Errorf("RandomInt(%d, %d) = %d, 期望 %d",
					tt.min, tt.max, result, tt.min)
			}
		})
	}

	// 测试随机性 - 生成多次，检查是否会产生不同的值
	t.Run("随机性测试", func(t *testing.T) {
		results := make(map[int]bool)
		for i := 0; i < 100; i++ {
			val := rand.RandomInt(1, 50)
			results[val] = true
		}
		// 至少应该生成10个不同的值
		if len(results) < 10 {
			t.Errorf("随机性不足，100次生成了 %d 个不同的值", len(results))
		}
	})
}

// TestRandomInt64 测试随机int64整数生成
func TestRandomInt64(t *testing.T) {
	rand := getRandEngine()

	tests := []struct {
		name string
		min  int64
		max  int64
	}{
		{"正常范围", 1, 100},
		{"相同值", 50, 50},
		{"反转范围", 100, 1},
		{"负数范围", -100, -1},
		{"混合范围", -50, 50},
		{"大范围", 0, 1000000},
		{"极大范围", -1000000000, 1000000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rand.RandomInt64(tt.min, tt.max)

			// 确定实际的最小值和最大值
			actualMin, actualMax := tt.min, tt.max
			if actualMin > actualMax {
				actualMin, actualMax = actualMax, actualMin
			}

			// 验证结果在范围内
			if result < actualMin || result > actualMax {
				t.Errorf("RandomInt64(%d, %d) = %d, 结果超出范围 [%d, %d]",
					tt.min, tt.max, result, actualMin, actualMax)
			}

			// 如果 min == max，结果应该等于该值
			if tt.min == tt.max && result != tt.min {
				t.Errorf("RandomInt64(%d, %d) = %d, 期望 %d",
					tt.min, tt.max, result, tt.min)
			}
		})
	}
}

// TestRandomFloat 测试随机浮点数生成
func TestRandomFloat(t *testing.T) {
	rand := getRandEngine()

	tests := []struct {
		name string
		min  float64
		max  float64
	}{
		{"正常范围", 0.0, 100.0},
		{"相同值", 50.0, 50.0},
		{"反转范围", 100.0, 0.0},
		{"负数范围", -100.0, -1.0},
		{"混合范围", -50.0, 50.0},
		{"小范围", 0.0, 1.0},
		{"大范围", 0.0, 10000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rand.RandomFloat(tt.min, tt.max)

			// 确定实际的最小值和最大值
			actualMin, actualMax := tt.min, tt.max
			if actualMin > actualMax {
				actualMin, actualMax = actualMax, actualMin
			}

			// 如果 min == max，结果应该等于该值
			if actualMin == actualMax {
				if result != actualMin {
					t.Errorf("RandomFloat(%f, %f) = %f, 期望 %f",
						tt.min, tt.max, result, actualMin)
				}
				return
			}

			// 验证结果在范围内 [min, max)
			if result < actualMin || result >= actualMax {
				t.Errorf("RandomFloat(%f, %f) = %f, 结果超出范围 [%f, %f)",
					tt.min, tt.max, result, actualMin, actualMax)
			}
		})
	}
}

// TestRandomBool 测试随机布尔值生成
func TestRandomBool(t *testing.T) {
	rand := getRandEngine()

	// 生成多次，统计 true 和 false 的数量
	trueCount := 0
	falseCount := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		result := rand.RandomBool()
		if result {
			trueCount++
		} else {
			falseCount++
		}
	}

	// 两种值都应该出现过
	if trueCount == 0 {
		t.Error("RandomBool() 没有生成过 true 值")
	}
	if falseCount == 0 {
		t.Error("RandomBool() 没有生成过 false 值")
	}

	// 理论上，两种值应该接近均衡（允许一定偏差）
	// 在1000次中，每种值至少应该出现300次
	minExpected := iterations * 30 / 100
	if trueCount < minExpected {
		t.Errorf("RandomBool() 的 true 值出现次数过低: %d (期望至少 %d)", trueCount, minExpected)
	}
	if falseCount < minExpected {
		t.Errorf("RandomBool() 的 false 值出现次数过低: %d (期望至少 %d)", falseCount, minExpected)
	}

	t.Logf("RandomBool 统计: true=%d, false=%d", trueCount, falseCount)
}

// TestRandomStringFromCharset 测试从指定字符集生成随机字符串
func TestRandomStringFromCharset(t *testing.T) {
	rand := getRandEngine()

	tests := []struct {
		name    string
		length  int
		charset string
		wantErr bool
	}{
		{"正常字母数字", 10, "abc123", false},
		{"特殊字符", 15, "!@#$%^&*()", false},
		{"单一字符", 5, "a", false},
		{"空字符集", 10, "", true},
		{"零长度", 0, "abc", true},
		{"负长度", -5, "abc", true},
		{"大长度", 1000, "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rand.RandomStringFromCharset(tt.length, tt.charset)

			// 检查长度
			if tt.length > 0 && tt.charset != "" {
				if len(result) != tt.length {
					t.Errorf("RandomStringFromCharset(%d, %q) 长度 = %d, 期望 %d",
						tt.length, tt.charset, len(result), tt.length)
				}

				// 检查所有字符都在字符集中
				for _, char := range result {
					found := false
					for _, c := range tt.charset {
						if char == c {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("RandomStringFromCharset(%d, %q) 包含不在字符集中的字符: %c",
							tt.length, tt.charset, char)
					}
				}
			} else if tt.wantErr && result != "" {
				t.Errorf("RandomStringFromCharset(%d, %q) 应该返回空字符串，但得到: %s",
					tt.length, tt.charset, result)
			}
		})
	}
}

// TestRandomString 测试随机字母数字字符串生成
func TestRandomString(t *testing.T) {
	rand := getRandEngine()

	tests := []struct {
		name   string
		length int
	}{
		{"短字符串", 5},
		{"中等字符串", 20},
		{"长字符串", 100},
		{"零长度", 0},
		{"单字符", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rand.RandomString(tt.length)

			if tt.length == 0 {
				if result != "" {
					t.Errorf("RandomString(0) 应该返回空字符串，但得到: %s", result)
				}
				return
			}

			if len(result) != tt.length {
				t.Errorf("RandomString(%d) 长度 = %d, 期望 %d", tt.length, len(result), tt.length)
			}

			// 检查只包含字母和数字
			validCharset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, char := range result {
				found := false
				for _, c := range validCharset {
					if char == c {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("RandomString(%d) 包含无效字符: %c", tt.length, char)
				}
			}
		})
	}
}

// TestRandomLetters 测试随机字母字符串生成
func TestRandomLetters(t *testing.T) {
	rand := getRandEngine()

	length := 50
	result := rand.RandomLetters(length)

	if len(result) != length {
		t.Errorf("RandomLetters(%d) 长度 = %d, 期望 %d", length, len(result), length)
	}

	// 检查只包含字母
	validCharset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, char := range result {
		found := false
		for _, c := range validCharset {
			if char == c {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomLetters(%d) 包含非字母字符: %c", length, char)
		}
	}
}

// TestRandomDigits 测试随机数字字符串生成
func TestRandomDigits(t *testing.T) {
	rand := getRandEngine()

	length := 30
	result := rand.RandomDigits(length)

	if len(result) != length {
		t.Errorf("RandomDigits(%d) 长度 = %d, 期望 %d", length, len(result), length)
	}

	// 检查只包含数字
	validCharset := "0123456789"
	for _, char := range result {
		found := false
		for _, c := range validCharset {
			if char == c {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomDigits(%d) 包含非数字字符: %c", length, char)
		}
	}
}

// TestRandomLowercase 测试随机小写字母字符串生成
func TestRandomLowercase(t *testing.T) {
	rand := getRandEngine()

	length := 40
	result := rand.RandomLowercase(length)

	if len(result) != length {
		t.Errorf("RandomLowercase(%d) 长度 = %d, 期望 %d", length, len(result), length)
	}

	// 检查只包含小写字母
	validCharset := "abcdefghijklmnopqrstuvwxyz"
	for _, char := range result {
		found := false
		for _, c := range validCharset {
			if char == c {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomLowercase(%d) 包含非小写字母字符: %c", length, char)
		}
	}
}

// TestRandomUppercase 测试随机大写字母字符串生成
func TestRandomUppercase(t *testing.T) {
	rand := getRandEngine()

	length := 35
	result := rand.RandomUppercase(length)

	if len(result) != length {
		t.Errorf("RandomUppercase(%d) 长度 = %d, 期望 %d", length, len(result), length)
	}

	// 检查只包含大写字母
	validCharset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, char := range result {
		found := false
		for _, c := range validCharset {
			if char == c {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomUppercase(%d) 包含非大写字母字符: %c", length, char)
		}
	}
}

// TestRandomUUID 测试UUID生成
func TestRandomUUID(t *testing.T) {
	rand := getRandEngine()

	// UUID v4 格式正则表达式
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	// 生成多个UUID并验证
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		uuid := rand.RandomUUID()

		// 检查格式
		if !uuidRegex.MatchString(uuid) {
			t.Errorf("RandomUUID() = %s, 格式不正确（应符合 UUID v4 格式）", uuid)
		}

		// 检查长度（标准UUID是36个字符，包含4个连字符）
		if len(uuid) != 36 {
			t.Errorf("RandomUUID() = %s, 长度 = %d, 期望 36", uuid, len(uuid))
		}

		// 检查唯一性
		if uuids[uuid] {
			t.Errorf("RandomUUID() 生成了重复的UUID: %s", uuid)
		}
		uuids[uuid] = true

		// 验证结构：8-4-4-4-12
		parts := strings.Split(uuid, "-")
		if len(parts) != 5 {
			t.Errorf("RandomUUID() = %s, 结构不正确", uuid)
		} else {
			if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 ||
				len(parts[3]) != 4 || len(parts[4]) != 12 {
				t.Errorf("RandomUUID() = %s, 各段长度不正确", uuid)
			}
		}
	}

	t.Logf("成功生成 %d 个唯一的UUID", len(uuids))
}

// TestRandomChoice 测试泛型函数 - 随机选择一个元素
func TestRandomChoice(t *testing.T) {
	// 测试整数切片
	t.Run("整数切片", func(t *testing.T) {
		options := []int{1, 2, 3, 4, 5}
		result := RandomChoice(options)

		found := false
		for _, opt := range options {
			if result == opt {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomChoice() = %d, 不在选项中 %v", result, options)
		}
	})

	// 测试字符串切片
	t.Run("字符串切片", func(t *testing.T) {
		options := []string{"apple", "banana", "cherry"}
		result := RandomChoice(options)

		found := false
		for _, opt := range options {
			if result == opt {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomChoice() = %s, 不在选项中 %v", result, options)
		}
	})

	// 测试空切片
	t.Run("空整数切片", func(t *testing.T) {
		options := []int{}
		result := RandomChoice(options)

		if result != 0 {
			t.Errorf("RandomChoice([]) = %d, 期望零值 0", result)
		}
	})

	// 测试空字符串切片
	t.Run("空字符串切片", func(t *testing.T) {
		options := []string{}
		result := RandomChoice(options)

		if result != "" {
			t.Errorf("RandomChoice([]) = %s, 期望空字符串", result)
		}
	})

	// 测试单个元素
	t.Run("单个元素", func(t *testing.T) {
		options := []int{42}
		result := RandomChoice(options)

		if result != 42 {
			t.Errorf("RandomChoice([42]) = %d, 期望 42", result)
		}
	})

	// 测试随机性
	t.Run("随机性测试", func(t *testing.T) {
		options := []int{1, 2, 3, 4, 5}
		results := make(map[int]bool)

		for i := 0; i < 100; i++ {
			result := RandomChoice(options)
			results[result] = true
		}

		// 至少应该选择过3个不同的元素
		if len(results) < 3 {
			t.Errorf("RandomChoice() 随机性不足，100次只选择了 %d 个不同的元素", len(results))
		}
	})
}

// TestRandomChoices 测试泛型函数 - 随机选择多个元素
func TestRandomChoices(t *testing.T) {
	// 测试正常情况
	t.Run("正常选择", func(t *testing.T) {
		options := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		count := 3

		results := RandomChoices(options, count)

		if len(results) != count {
			t.Errorf("RandomChoices(..., %d) 返回了 %d 个元素", count, len(results))
		}

		// 检查没有重复
		unique := make(map[int]bool)
		for _, r := range results {
			if unique[r] {
				t.Errorf("RandomChoices() 返回了重复的元素: %d", r)
			}
			unique[r] = true
		}

		// 检查所有元素都在原切片中
		for _, r := range results {
			found := false
			for _, opt := range options {
				if r == opt {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("RandomChoices() 返回了不在原切片中的元素: %d", r)
			}
		}
	})

	// 测试请求数量等于或超过切片长度
	t.Run("请求全部元素", func(t *testing.T) {
		options := []int{1, 2, 3, 4, 5}
		count := 5

		results := RandomChoices(options, count)

		if len(results) != len(options) {
			t.Errorf("RandomChoices(..., %d) 返回了 %d 个元素，期望 %d",
				count, len(results), len(options))
		}

		// 验证包含所有元素（可能顺序不同）
		resultMap := make(map[int]bool)
		for _, r := range results {
			resultMap[r] = true
		}

		for _, opt := range options {
			if !resultMap[opt] {
				t.Errorf("RandomChoices() 缺少元素: %d", opt)
			}
		}
	})

	// 测试请求数量超过切片长度
	t.Run("请求超过切片长度", func(t *testing.T) {
		options := []int{1, 2, 3}
		count := 10

		results := RandomChoices(options, count)

		// 应该返回所有元素
		if len(results) != len(options) {
			t.Errorf("RandomChoices(..., %d) 返回了 %d 个元素，期望 %d",
				count, len(results), len(options))
		}
	})

	// 测试空切片
	t.Run("空切片", func(t *testing.T) {
		options := []int{}
		count := 3

		results := RandomChoices(options, count)

		if results != nil {
			t.Errorf("RandomChoices([], %d) 应该返回 nil，得到 %v", count, results)
		}
	})

	// 测试count为零
	t.Run("count为零", func(t *testing.T) {
		options := []int{1, 2, 3}
		count := 0

		results := RandomChoices(options, count)

		if results != nil {
			t.Errorf("RandomChoices(..., 0) 应该返回 nil，得到 %v", results)
		}
	})

	// 测试负数count
	t.Run("负数count", func(t *testing.T) {
		options := []int{1, 2, 3}
		count := -5

		results := RandomChoices(options, count)

		if results != nil {
			t.Errorf("RandomChoices(..., -5) 应该返回 nil，得到 %v", results)
		}
	})

	// 测试count为1
	t.Run("单个元素", func(t *testing.T) {
		options := []int{1, 2, 3, 4, 5}
		count := 1

		results := RandomChoices(options, count)

		if len(results) != 1 {
			t.Errorf("RandomChoices(..., 1) 返回了 %d 个元素", len(results))
		}

		// 检查元素在原切片中
		found := false
		for _, opt := range options {
			if results[0] == opt {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomChoices()[0] = %d 不在原切片中", results[0])
		}
	})

	// 测试字符串切片
	t.Run("字符串切片", func(t *testing.T) {
		options := []string{"apple", "banana", "cherry", "date", "elderberry"}
		count := 3

		results := RandomChoices(options, count)

		if len(results) != count {
			t.Errorf("RandomChoices(..., %d) 返回了 %d 个元素", count, len(results))
		}

		// 检查没有重复
		unique := make(map[string]bool)
		for _, r := range results {
			if unique[r] {
				t.Errorf("RandomChoices() 返回了重复的元素: %s", r)
			}
			unique[r] = true
		}
	})

	// 测试随机性
	t.Run("随机性测试", func(t *testing.T) {
		options := []int{1, 2, 3, 4, 5}
		count := 3

		// 多次选择，统计组合的多样性
		combinations := make(map[string]bool)

		for i := 0; i < 50; i++ {
			results := RandomChoices(options, count)
			// 创建一个排序后的键来表示组合
			key := ""
			for j, r := range results {
				if j > 0 {
					key += ","
				}
				key += string(rune('0' + r))
			}
			combinations[key] = true
		}

		// 至少应该产生10种不同的组合
		if len(combinations) < 10 {
			t.Errorf("RandomChoices() 随机性不足，50次只产生了 %d 种不同的组合", len(combinations))
		}

		t.Logf("产生了 %d 种不同的组合", len(combinations))
	})
}

// BenchmarkRandomInt 性能测试 - 随机整数
func BenchmarkRandomInt(b *testing.B) {
	rand := getRandEngine()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rand.RandomInt(1, 100)
	}
}

// BenchmarkRandomInt64 性能测试 - 随机int64整数
func BenchmarkRandomInt64(b *testing.B) {
	rand := getRandEngine()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rand.RandomInt64(1, 100)
	}
}

// BenchmarkRandomFloat 性能测试 - 随机浮点数
func BenchmarkRandomFloat(b *testing.B) {
	rand := getRandEngine()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rand.RandomFloat(0.0, 100.0)
	}
}

// BenchmarkRandomString 性能测试 - 随机字符串
func BenchmarkRandomString(b *testing.B) {
	rand := getRandEngine()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rand.RandomString(20)
	}
}

// BenchmarkRandomUUID 性能测试 - UUID生成
func BenchmarkRandomUUID(b *testing.B) {
	rand := getRandEngine()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rand.RandomUUID()
	}
}

// BenchmarkRandomChoice 性能测试 - 随机选择单个元素
func BenchmarkRandomChoice(b *testing.B) {
	options := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		RandomChoice(options)
	}
}

// BenchmarkRandomChoices 性能测试 - 随机选择多个元素
func BenchmarkRandomChoices(b *testing.B) {
	options := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		RandomChoices(options, 5)
	}
}
