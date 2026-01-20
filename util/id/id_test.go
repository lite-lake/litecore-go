package id

import (
	"regexp"
	"sync"
	"testing"
)

// TestNewCUID2_Basic 测试基本的 CUID2 生成功能
func TestNewCUID2_Basic(t *testing.T) {
	t.Run("生成单个ID", func(t *testing.T) {
		id, err := NewCUID2()
		if err != nil {
			t.Fatalf("NewCUID2() 返回错误: %v", err)
		}

		// 检查长度
		if len(id) != 25 {
			t.Errorf("NewCUID2() 长度 = %d, 期望 25", len(id))
		}

		// 检查不为空
		if id == "" {
			t.Error("NewCUID2() 返回了空字符串")
		}
	})

	t.Run("生成多个ID", func(t *testing.T) {
		ids := make([]string, 10)
		for i := 0; i < 10; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			ids[i] = id
		}

		// 检查所有ID都有正确长度
		for i, id := range ids {
			if len(id) != 25 {
				t.Errorf("ID[%d] 长度 = %d, 期望 25", i, len(id))
			}
			if id == "" {
				t.Errorf("ID[%d] 为空字符串", i)
			}
		}
	})
}

// TestNewCUID2_Format 测试 CUID2 格式
func TestNewCUID2_Format(t *testing.T) {
	// CUID2 应该只包含小写字母和数字
	cuidRegex := regexp.MustCompile(`^[0-9a-z]{25}$`)

	id, err := NewCUID2()
	if err != nil {
		t.Fatalf("NewCUID2() 返回错误: %v", err)
	}

	if !cuidRegex.MatchString(id) {
		t.Errorf("NewCUID2() = %s, 格式不正确，应该只包含小写字母和数字", id)
	}

	// 生成更多ID验证格式一致性
	for i := 0; i < 100; i++ {
		id, err := NewCUID2()
		if err != nil {
			t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
		}
		if !cuidRegex.MatchString(id) {
			t.Errorf("NewCUID2() 第%d次生成 = %s, 格式不正确", i, id)
		}
	}
}

// TestNewCUID2_Uniqueness 测试 CUID2 唯一性
func TestNewCUID2_Uniqueness(t *testing.T) {
	t.Run("小批量唯一性", func(t *testing.T) {
		ids := make(map[string]bool)
		count := 1000

		for i := 0; i < count; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			if ids[id] {
				t.Errorf("生成重复ID: %s (第%d次)", id, i)
			}
			ids[id] = true
		}

		if len(ids) != count {
			t.Errorf("生成了 %d 个ID，但只有 %d 个唯一ID", count, len(ids))
		}

		t.Logf("成功生成 %d 个唯一的 CUID2", len(ids))
	})

	t.Run("大批量唯一性", func(t *testing.T) {
		ids := make(map[string]bool)
		count := 10000

		for i := 0; i < count; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			if ids[id] {
				t.Errorf("生成重复ID: %s (第%d次)", id, i)
				return
			}
			ids[id] = true
		}

		if len(ids) != count {
			t.Errorf("生成了 %d 个ID，但只有 %d 个唯一ID", count, len(ids))
		}

		t.Logf("成功生成 %d 个唯一的 CUID2", len(ids))
	})
}

// TestNewCUID2_Randomness 测试随机性
func TestNewCUID2_Randomness(t *testing.T) {
	ids := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		id, err := NewCUID2()
		if err != nil {
			t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
		}
		ids[i] = id
	}

	// 检查是否有重复
	uniqueIds := make(map[string]bool)
	for _, id := range ids {
		uniqueIds[id] = true
	}

	if len(uniqueIds) != 1000 {
		t.Errorf("随机性不足：1000次生成了 %d 个唯一ID", len(uniqueIds))
	}

	// 检查字符分布是否合理（不应该所有ID都相同或非常相似）
	firstChars := make(map[rune]bool)
	for _, id := range ids {
		firstChars[rune(id[0])] = true
	}

	// 第一个字符应该有多种可能
	if len(firstChars) < 5 {
		t.Logf("警告：首字符分布较窄，只有 %d 种不同字符", len(firstChars))
	}
}

// TestNewCUID2_Concurrency 测试并发安全性
func TestNewCUID2_Concurrency(t *testing.T) {
	const goroutines = 100
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	ids := make(chan string, goroutines*idsPerGoroutine)
	uniqueIds := sync.Map{} // 用于并发安全的去重

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := NewCUID2()
				if err != nil {
					t.Errorf("NewCUID2() 返回错误: %v", err)
					return
				}
				ids <- id

				// 检查重复
				if _, exists := uniqueIds.LoadOrStore(id, true); exists {
					t.Errorf("并发生成重复ID: %s", id)
				}
			}
		}()
	}

	wg.Wait()
	close(ids)

	totalCount := 0
	for range ids {
		totalCount++
	}

	expectedCount := goroutines * idsPerGoroutine
	if totalCount != expectedCount {
		t.Errorf("期望生成 %d 个ID，实际生成 %d 个", expectedCount, totalCount)
	}

	t.Logf("并发测试完成：%d 个goroutine，每个生成 %d 个ID，共 %d 个唯一ID",
		goroutines, idsPerGoroutine, totalCount)
}

// TestNewCUID2_Consistency 测试一致性（相同的输入不一定产生相同的输出，因为是随机的）
func TestNewCUID2_Consistency(t *testing.T) {
	t.Run("不同ID应该不同", func(t *testing.T) {
		id1, err := NewCUID2()
		if err != nil {
			t.Fatalf("NewCUID2() 第一次返回错误: %v", err)
		}
		id2, err := NewCUID2()
		if err != nil {
			t.Fatalf("NewCUID2() 第二次返回错误: %v", err)
		}

		if id1 == id2 {
			t.Errorf("连续两次生成应该产生不同的ID，但得到: %s", id1)
		}
	})

	t.Run("快速连续生成", func(t *testing.T) {
		ids := make([]string, 100)
		for i := 0; i < 100; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			ids[i] = id
		}

		// 检查是否有重复
		seen := make(map[string]bool)
		for i, id := range ids {
			if seen[id] {
				t.Errorf("快速连续生成了重复ID: %s (索引 %d)", id, i)
			}
			seen[id] = true
		}
	})
}

// TestNewCUID2_CharacterSet 测试字符集
func TestNewCUID2_CharacterSet(t *testing.T) {
	validCharset := "0123456789abcdefghijklmnopqrstuvwxyz"

	t.Run("验证所有字符都在有效字符集中", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}

			for j, char := range id {
				found := false
				for _, validChar := range validCharset {
					if char == validChar {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ID[%d][%d] = %c 包含无效字符，ID: %s", i, j, char, id)
				}
			}
		}
	})

	t.Run("验证不包含大写字母", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			for j, char := range id {
				if char >= 'A' && char <= 'Z' {
					t.Errorf("ID[%d][%d] = %c 包含大写字母，ID: %s", i, j, char, id)
				}
			}
		}
	})

	t.Run("验证不包含特殊字符", func(t *testing.T) {
		specialChars := "-_+!@#$%^&*()=[]{}|;:',.<>?/~` "
		for i := 0; i < 100; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			for j, char := range id {
				for _, special := range specialChars {
					if char == special {
						t.Errorf("ID[%d][%d] = %c 包含特殊字符 %c, ID: %s", i, j, char, special, id)
					}
				}
			}
		}
	})
}

// TestNewCUID2_EdgeCases 边界情况测试
func TestNewCUID2_EdgeCases(t *testing.T) {
	t.Run("连续大量生成", func(t *testing.T) {
		ids := make(map[string]bool)
		count := 50000

		for i := 0; i < count; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			if ids[id] {
				t.Errorf("在 %d 次生成中发现重复ID: %s", i+1, id)
				return
			}
			ids[id] = true
		}

		t.Logf("连续生成 %d 个ID，全部唯一", count)
	})

	t.Run("验证固定长度", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			if len(id) != 25 {
				t.Errorf("第 %d 次生成长度不正确: %d, ID: %s", i, len(id), id)
				return
			}
		}
	})
}

// TestNewCUID2_Properties 测试 CUID2 的其他属性
func TestNewCUID2_Properties(t *testing.T) {
	t.Run("排序性-时间戳在前", func(t *testing.T) {
		// 快速连续生成两个ID，第一个应该"较小"（因为时间戳在前）
		id1, err := NewCUID2()
		if err != nil {
			t.Fatalf("NewCUID2() 第一次返回错误: %v", err)
		}
		id2, err := NewCUID2()
		if err != nil {
			t.Fatalf("NewCUID2() 第二次返回错误: %v", err)
		}

		// 由于时间戳在前，通常后生成的ID应该更大（或相等）
		// 但这取决于生成速度，所以这个测试只是验证它们不同
		if id1 == id2 {
			t.Error("时间戳没有起到作用，生成了相同的ID")
		}
	})

	t.Run("前缀变化测试", func(t *testing.T) {
		// 间隔一段时间生成ID，前缀应该变化
		ids := make([]string, 100)
		prefixes := make(map[string]bool)

		for i := 0; i < 100; i++ {
			id, err := NewCUID2()
			if err != nil {
				t.Fatalf("NewCUID2() 第%d次返回错误: %v", i, err)
			}
			ids[i] = id
			// 取前5个字符作为前缀
			prefix := ids[i][:5]
			prefixes[prefix] = true
		}

		// 应该有多个不同的前缀（因为在100次生成中时间会流逝）
		if len(prefixes) < 2 {
			t.Logf("警告：100次生成中只有 %d 个不同前缀，可能生成速度太快", len(prefixes))
		}

		t.Logf("100个ID中有 %d 个不同的前缀", len(prefixes))
	})
}

// BenchmarkNewCUID2 性能基准测试
func BenchmarkNewCUID2(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewCUID2()
	}
}

// BenchmarkNewCUID2_Parallel 并发性能基准测试
func BenchmarkNewCUID2_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewCUID2()
		}
	})
}

// BenchmarkNewCUID2_Batch 批量生成基准测试
func BenchmarkNewCUID2_Batch(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次生成100个ID
		for j := 0; j < 100; j++ {
			NewCUID2()
		}
	}
}

// ExampleNewCUID2 示例代码
func ExampleNewCUID2() {
	id, _ := NewCUID2()
	_ = id // 使用生成的ID
}
