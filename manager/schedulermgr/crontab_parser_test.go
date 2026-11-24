package schedulermgr

import (
	"testing"
	"time"
)

func TestParseStep_StarSlash(t *testing.T) {
	t.Run("*/10 秒字段", func(t *testing.T) {
		cf, err := parseField("*/10", 0, 59, nil)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if cf.allowedAll {
			t.Fatal("*/10 不应设置 allowedAll，步长值被忽略")
		}
		expected := []int{0, 10, 20, 30, 40, 50}
		if len(cf.allowed) != len(expected) {
			t.Fatalf("期望 %d 个值，实际 %d 个值: %v", len(expected), len(cf.allowed), cf.allowed)
		}
		for i, v := range expected {
			if cf.allowed[i] != v {
				t.Errorf("期望 [%d]=%d，实际 %d", i, v, cf.allowed[i])
			}
		}
	})
}

func TestParseStep_RangeSlash(t *testing.T) {
	t.Run("0-59/10 秒字段", func(t *testing.T) {
		cf, err := parseField("0-59/10", 0, 59, nil)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		expected := []int{0, 10, 20, 30, 40, 50}
		if len(cf.allowed) != len(expected) {
			t.Fatalf("期望 %d 个值，实际 %d 个值: %v", len(expected), len(cf.allowed), cf.allowed)
		}
	})
}

func TestParseStep_ValueSlash(t *testing.T) {
	t.Run("5/15 秒字段", func(t *testing.T) {
		cf, err := parseField("5/15", 0, 59, nil)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		expected := []int{5, 20, 35, 50}
		if len(cf.allowed) != len(expected) {
			t.Fatalf("期望 %d 个值，实际 %d 个值: %v", len(expected), len(cf.allowed), cf.allowed)
		}
	})
}

func TestParseField_Star(t *testing.T) {
	t.Run("* 应设置 allowedAll", func(t *testing.T) {
		cf, err := parseField("*", 0, 59, nil)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if !cf.allowedAll {
			t.Fatal("* 应设置 allowedAll=true")
		}
	})
}

func TestParseField_Comma(t *testing.T) {
	t.Run("逗号分隔值", func(t *testing.T) {
		cf, err := parseField("0,15,30,45", 0, 59, nil)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		expected := []int{0, 15, 30, 45}
		if len(cf.allowed) != len(expected) {
			t.Fatalf("期望 %d 个值，实际 %d 个值: %v", len(expected), len(cf.allowed), cf.allowed)
		}
	})
}

func TestParseField_Range(t *testing.T) {
	t.Run("范围值", func(t *testing.T) {
		cf, err := parseField("10-12", 0, 59, nil)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		expected := []int{10, 11, 12}
		if len(cf.allowed) != len(expected) {
			t.Fatalf("期望 %d 个值，实际 %d 个值: %v", len(expected), len(cf.allowed), cf.allowed)
		}
	})
}

func TestCronExpression_Every10Seconds(t *testing.T) {
	t.Run("*/10 * * * * * 应每 10 秒执行", func(t *testing.T) {
		ce, err := parseCrontab("*/10 * * * * *", "Asia/Shanghai")
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		base := time.Date(2026, 7, 22, 14, 30, 0, 0, time.UTC)
		next := ce.getNextExecutionTime(base)

		// base+1s 是 14:30:01，不匹配 */10
		// 下一个匹配是 14:30:10
		if next.Second() != 10 {
			t.Errorf("期望秒=10，实际秒=%d", next.Second())
		}
	})
}

func TestCronExpression_Every5Minutes(t *testing.T) {
	t.Run("0 */5 * * * * 应每 5 分钟执行", func(t *testing.T) {
		ce, err := parseCrontab("0 */5 * * * *", "Asia/Shanghai")
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		base := time.Date(2026, 7, 22, 14, 2, 0, 0, time.UTC)
		next := ce.getNextExecutionTime(base)

		// base+1s 是 14:02:01，秒=0 不匹配
		// 下一个匹配是 14:05:00
		if next.Minute() != 5 {
			t.Errorf("期望分钟=5，实际分钟=%d", next.Minute())
		}
		if next.Second() != 0 {
			t.Errorf("期望秒=0，实际秒=%d", next.Second())
		}
	})
}
