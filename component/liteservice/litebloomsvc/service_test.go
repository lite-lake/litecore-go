package litebloomsvc

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLiteBloomService(t *testing.T) {
	t.Run("使用默认配置创建服务", func(t *testing.T) {
		svc := NewLiteBloomService()
		assert.NotNil(t, svc)

		filters := svc.ListFilters()
		assert.Empty(t, filters)
	})
}

func TestNewLiteBloomServiceWithConfig(t *testing.T) {
	t.Run("使用自定义配置创建服务", func(t *testing.T) {
		expectedItems := uint(5000)
		falsePositiveRate := 0.001
		ttl := time.Hour

		config := &Config{
			DefaultExpectedItems:     &expectedItems,
			DefaultFalsePositiveRate: &falsePositiveRate,
			DefaultTTL:               &ttl,
		}

		svc := NewLiteBloomServiceWithConfig(config)
		assert.NotNil(t, svc)

		err := svc.CreateFilter("test")
		assert.NoError(t, err)

		stats, err := svc.GetStats("test")
		assert.NoError(t, err)
		assert.Equal(t, expectedItems, stats.ExpectedItems)
		assert.Equal(t, falsePositiveRate, stats.FalsePositiveRate)
		assert.Equal(t, ttl, stats.TTL)
	})

	t.Run("配置为 nil 时使用默认配置", func(t *testing.T) {
		svc := NewLiteBloomServiceWithConfig(nil)
		assert.NotNil(t, svc)

		err := svc.CreateFilter("test")
		assert.NoError(t, err)

		stats, err := svc.GetStats("test")
		assert.NoError(t, err)
		assert.Equal(t, uint(defaultExpectedItems), stats.ExpectedItems)
		assert.Equal(t, defaultFalsePositiveRate, stats.FalsePositiveRate)
	})
}

func TestCreateFilter(t *testing.T) {
	t.Run("创建默认配置的过滤器", func(t *testing.T) {
		svc := NewLiteBloomService()

		err := svc.CreateFilter("users")
		assert.NoError(t, err)

		filters := svc.ListFilters()
		assert.Len(t, filters, 1)
		assert.Contains(t, filters, "users")
	})

	t.Run("创建自定义配置的过滤器", func(t *testing.T) {
		svc := NewLiteBloomService()
		expectedItems := uint(100000)
		falsePositiveRate := 0.001

		config := &FilterConfig{
			ExpectedItems:     &expectedItems,
			FalsePositiveRate: &falsePositiveRate,
		}

		err := svc.CreateFilterWithConfig("emails", config)
		assert.NoError(t, err)

		stats, err := svc.GetStats("emails")
		assert.NoError(t, err)
		assert.Equal(t, expectedItems, stats.ExpectedItems)
		assert.Equal(t, falsePositiveRate, stats.FalsePositiveRate)
	})

	t.Run("名称为空时返回错误", func(t *testing.T) {
		svc := NewLiteBloomService()

		err := svc.CreateFilter("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "过滤器名称不能为空")
	})

	t.Run("重复创建同名过滤器", func(t *testing.T) {
		svc := NewLiteBloomService()

		err := svc.CreateFilter("test")
		assert.NoError(t, err)

		err = svc.CreateFilter("test")
		assert.NoError(t, err)
	})
}

func TestDeleteFilter(t *testing.T) {
	t.Run("删除存在的过滤器", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		err := svc.DeleteFilter("test")
		assert.NoError(t, err)

		filters := svc.ListFilters()
		assert.Empty(t, filters)
	})

	t.Run("删除不存在的过滤器返回错误", func(t *testing.T) {
		svc := NewLiteBloomService()

		err := svc.DeleteFilter("notexist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "过滤器不存在")
	})
}

func TestAddAndContains(t *testing.T) {
	t.Run("添加和检查字节元素", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		err := svc.Add("test", []byte("hello"))
		assert.NoError(t, err)

		exists, err := svc.Contains("test", []byte("hello"))
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = svc.Contains("test", []byte("world"))
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("添加和检查字符串元素", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		err := svc.AddString("test", "hello")
		assert.NoError(t, err)

		exists, err := svc.ContainsString("test", "hello")
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = svc.ContainsString("test", "world")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("添加到不存在的过滤器返回错误", func(t *testing.T) {
		svc := NewLiteBloomService()

		err := svc.Add("notexist", []byte("test"))
		assert.Error(t, err)

		_, err = svc.Contains("notexist", []byte("test"))
		assert.Error(t, err)
	})
}

func TestAddBatch(t *testing.T) {
	t.Run("批量添加字节元素", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		dataList := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
		err := svc.AddBatch("test", dataList)
		assert.NoError(t, err)

		for _, data := range dataList {
			exists, err := svc.Contains("test", data)
			assert.NoError(t, err)
			assert.True(t, exists)
		}
	})

	t.Run("批量添加字符串元素", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		dataList := []string{"a", "b", "c"}
		err := svc.AddStringBatch("test", dataList)
		assert.NoError(t, err)

		for _, data := range dataList {
			exists, err := svc.ContainsString("test", data)
			assert.NoError(t, err)
			assert.True(t, exists)
		}
	})
}

func TestGetStats(t *testing.T) {
	t.Run("获取过滤器统计信息", func(t *testing.T) {
		svc := NewLiteBloomService()
		expectedItems := uint(10000)
		falsePositiveRate := 0.01

		err := svc.CreateFilterWithConfig("test", &FilterConfig{
			ExpectedItems:     &expectedItems,
			FalsePositiveRate: &falsePositiveRate,
		})
		assert.NoError(t, err)

		stats, err := svc.GetStats("test")
		assert.NoError(t, err)
		assert.Equal(t, "test", stats.Name)
		assert.Equal(t, expectedItems, stats.ExpectedItems)
		assert.Equal(t, falsePositiveRate, stats.FalsePositiveRate)
		assert.Greater(t, stats.BitSize, uint(0))
		assert.Greater(t, stats.HashFunctions, uint(0))
		assert.Equal(t, uint(0), stats.ElementCount)
	})

	t.Run("添加元素后统计信息更新", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		for i := 0; i < 100; i++ {
			svc.AddString("test", string(rune(i)))
		}

		stats, err := svc.GetStats("test")
		assert.NoError(t, err)
		assert.Greater(t, stats.ElementCount, uint(0))
		assert.Greater(t, stats.FillRatio, 0.0)
	})

	t.Run("获取不存在的过滤器统计返回错误", func(t *testing.T) {
		svc := NewLiteBloomService()

		_, err := svc.GetStats("notexist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "过滤器不存在")
	})
}

func TestRebuild(t *testing.T) {
	t.Run("重建过滤器", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		svc.AddString("test", "hello")
		exists, _ := svc.ContainsString("test", "hello")
		assert.True(t, exists)

		err := svc.Rebuild("test")
		assert.NoError(t, err)

		exists, _ = svc.ContainsString("test", "hello")
		assert.False(t, exists)

		stats, _ := svc.GetStats("test")
		assert.Equal(t, uint(0), stats.ElementCount)
	})

	t.Run("使用新配置重建过滤器", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		newExpectedItems := uint(50000)
		newFalsePositiveRate := 0.001

		err := svc.RebuildWithConfig("test", &FilterConfig{
			ExpectedItems:     &newExpectedItems,
			FalsePositiveRate: &newFalsePositiveRate,
		})
		assert.NoError(t, err)

		stats, _ := svc.GetStats("test")
		assert.Equal(t, newExpectedItems, stats.ExpectedItems)
		assert.Equal(t, newFalsePositiveRate, stats.FalsePositiveRate)
	})

	t.Run("重建不存在的过滤器返回错误", func(t *testing.T) {
		svc := NewLiteBloomService()

		err := svc.Rebuild("notexist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "过滤器不存在")
	})
}

func TestTTL(t *testing.T) {
	t.Run("过滤器带 TTL 创建", func(t *testing.T) {
		svc := NewLiteBloomService()
		ttl := time.Minute

		err := svc.CreateFilterWithConfig("test", &FilterConfig{
			TTL: &ttl,
		})
		assert.NoError(t, err)

		stats, err := svc.GetStats("test")
		assert.NoError(t, err)
		assert.Equal(t, ttl, stats.TTL)
		assert.False(t, stats.ExpiresAt.IsZero())
	})

	t.Run("过滤器永不过期", func(t *testing.T) {
		svc := NewLiteBloomService()

		err := svc.CreateFilter("test")
		assert.NoError(t, err)

		stats, err := svc.GetStats("test")
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), stats.TTL)
		assert.True(t, stats.ExpiresAt.IsZero())
	})
}

func TestLifecycle(t *testing.T) {
	t.Run("服务启动和停止", func(t *testing.T) {
		svc := NewLiteBloomService()

		err := svc.OnStart()
		assert.NoError(t, err)

		svc.CreateFilter("test")
		svc.AddString("test", "hello")

		err = svc.OnStop()
		assert.NoError(t, err)

		filters := svc.ListFilters()
		assert.Empty(t, filters)
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("并发添加元素", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				svc.AddString("test", string(rune(n)))
			}(i)
		}
		wg.Wait()

		stats, err := svc.GetStats("test")
		assert.NoError(t, err)
		assert.Greater(t, stats.ElementCount, uint(0))
	})

	t.Run("并发读写", func(t *testing.T) {
		svc := NewLiteBloomService()
		svc.CreateFilter("test")

		var wg sync.WaitGroup

		for i := 0; i < 50; i++ {
			wg.Add(2)
			go func(n int) {
				defer wg.Done()
				svc.AddString("test", string(rune(n)))
			}(i)
			go func(n int) {
				defer wg.Done()
				svc.ContainsString("test", string(rune(n)))
			}(i)
		}
		wg.Wait()
	})
}

func TestFalsePositiveRate(t *testing.T) {
	t.Run("验证误判率在预期范围内", func(t *testing.T) {
		svc := NewLiteBloomService()
		expectedItems := uint(10000)
		falsePositiveRate := 0.01

		err := svc.CreateFilterWithConfig("test", &FilterConfig{
			ExpectedItems:     &expectedItems,
			FalsePositiveRate: &falsePositiveRate,
		})
		assert.NoError(t, err)

		for i := 0; i < 1000; i++ {
			svc.AddString("test", string(rune(i)))
		}

		falsePositives := 0
		testCount := 1000
		for i := 1000; i < 1000+testCount; i++ {
			exists, _ := svc.ContainsString("test", string(rune(i)))
			if exists {
				falsePositives++
			}
		}

		actualRate := float64(falsePositives) / float64(testCount)
		assert.LessOrEqual(t, actualRate, falsePositiveRate*2)
	})
}
