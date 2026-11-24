package metrics

import (
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddConcurrent 并发累加不丢计数。
func TestAddConcurrent(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	Register("concurrent_total", "并发累加测试", Counter)

	const goroutines = 32
	const times = 2000
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < times; j++ {
				Add("concurrent_total", 1, nil)
			}
		}()
	}
	wg.Wait()

	output := Render()
	assert.Contains(t, output, "concurrent_total 64000\n")
}

// TestLabelEscaping 标签值中的反斜杠、双引号与换行必须转义。
func TestLabelEscaping(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	Register("escape_total", "转义测试", Counter)
	Add("escape_total", 1, Labels{"path": "a\"b\\c\nd"})

	output := Render()
	assert.Contains(t, output, `escape_total{path="a\"b\\c\nd"} 1`+"\n")
}

// TestRenderCounterAndGauge counter/gauge 的 HELP、TYPE、样本与排序。
func TestRenderCounterAndGauge(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	Register("demo_requests_total", "请求总数", Counter)
	Register("demo_inflight", "在途请求数", Gauge)

	Add("demo_requests_total", 2, nil)
	Add("demo_requests_total", 3, nil)
	Set("demo_inflight", 2.5, nil)

	expected := "# HELP demo_inflight 在途请求数\n" +
		"# TYPE demo_inflight gauge\n" +
		"demo_inflight 2.5\n" +
		"# HELP demo_requests_total 请求总数\n" +
		"# TYPE demo_requests_total counter\n" +
		"demo_requests_total 5\n"
	assert.Equal(t, expected, Render())
}

// TestRenderLabelOrder 标签按字典序输出，默认标签在序列创建时合并且显式标签优先。
func TestRenderLabelOrder(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	SetDefaultLabels(Labels{"product": "whois", "env": "dev"})
	Add("requests_total", 1, Labels{"tier": "free", "decision": "allow"})
	Set("quota_points_total", 3, Labels{"product": "override"})

	output := Render()
	assert.Contains(t, output, `requests_total{decision="allow",env="dev",product="whois",tier="free"} 1`)
	assert.Contains(t, output, `quota_points_total{env="dev",product="override"} 3`)
}

// TestEmptyRegistry 空注册表输出空字符串。
func TestEmptyRegistry(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	assert.Equal(t, "", Render())
}

// TestRegisterDuplicate 同类型重复注册幂等，类型冲突保留首次注册。
func TestRegisterDuplicate(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	Register("repeat_total", "首次 help", Counter)
	Register("repeat_total", "第二次 help", Counter)
	output := Render()
	assert.Contains(t, output, "# HELP repeat_total 首次 help\n")
	assert.NotContains(t, output, "第二次 help")
	assert.Equal(t, 1, strings.Count(output, "# TYPE repeat_total counter\n"))

	Register("conflict_total", "counter 版", Counter)
	Register("conflict_total", "gauge 版", Gauge)
	output = Render()
	assert.Contains(t, output, "# TYPE conflict_total counter\n")
	assert.NotContains(t, output, "conflict_total gauge")
}

// TestAutoRegister 未注册 name 在 Add/Set 时自动按 counter 注册（help 为空）。
func TestAutoRegister(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	Add("auto_total", 2, nil)
	Set("auto_gauge_like", 7, nil)

	output := Render()
	assert.Contains(t, output, "# HELP auto_total \n")
	assert.Contains(t, output, "# TYPE auto_total counter\n")
	assert.Contains(t, output, "auto_total 2\n")
	assert.Contains(t, output, "# TYPE auto_gauge_like counter\n")
	assert.Contains(t, output, "auto_gauge_like 7\n")
}

// TestFormatFloat 浮点使用最小表示，特殊值使用 Prometheus 约定写法。
func TestFormatFloat(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		want  string
	}{
		{"零", 0, "0"},
		{"整数", 1, "1"},
		{"负整数", -12, "-12"},
		{"小数", 1.5, "1.5"},
		{"大整数", 1e6, "1000000"},
		{"超出整型阈值", 1e15, "1e+15"},
		{"正无穷", math.Inf(1), "+Inf"},
		{"负无穷", math.Inf(-1), "-Inf"},
		{"非数", math.NaN(), "NaN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, formatFloat(tc.value))
		})
	}
}

// TestHelpEscaping HELP 文本中的反斜杠与换行转义。
func TestHelpEscaping(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	Register("help_total", "第一行\\反斜杠\n第二行", Counter)
	require.Contains(t, Render(), `# HELP help_total 第一行\\反斜杠\n第二行`)
}

// TestInvalidMetricNameIgnored 空名或非法指标名不产生任何样本/定义。
func TestInvalidMetricNameIgnored(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	for _, name := range []string{"", "1abc", "bad-name", "bad name", "bad.name", "中文名"} {
		Register(name, "非法名", Counter)
		Add(name, 1, nil)
		Set(name, 1, nil)
		assert.Nil(t, lookupOrRegister(name), "非法名 %q 不应返回指标", name)
	}

	assert.Equal(t, "", Render())
	assert.Empty(t, registryNames())
}

// TestInvalidLabelKeyFiltered 非法标签键被忽略，合法键正常输出。
func TestInvalidLabelKeyFiltered(t *testing.T) {
	ResetForTest()
	t.Cleanup(ResetForTest)

	Register("label_total", "标签过滤", Counter)
	Add("label_total", 1, Labels{
		"good_key": "v",
		"bad-key":  "x",
		"1bad":     "y",
		"bad key":  "z",
		"中文键":      "w",
	})

	output := Render()
	assert.Contains(t, output, `label_total{good_key="v"} 1`)
	assert.NotContains(t, output, "bad-key")
	assert.NotContains(t, output, "1bad")
	assert.NotContains(t, output, "bad key")
	assert.NotContains(t, output, "中文键")
}

// TestValidMetricName 指标名规范：允许字母/数字/下划线/冒号，且不能以数字开头。
func TestValidMetricName(t *testing.T) {
	valid := []string{"a", "_hidden", ":recording_rule", "A1_b:c", "requests_total", "ns:metric_name_2"}
	for _, name := range valid {
		assert.True(t, validMetricName(name), "应合法: %q", name)
	}
	invalid := []string{"", "1abc", "a-b", "a b", "a.b", "a/b", "é"}
	for _, name := range invalid {
		assert.False(t, validMetricName(name), "应非法: %q", name)
	}

	ResetForTest()
	t.Cleanup(ResetForTest)
	Add(":recording_rule_total", 2, Labels{"_label": "x", "1bad": "y"})
	output := Render()
	assert.Contains(t, output, "# TYPE :recording_rule_total counter\n")
	assert.Contains(t, output, `:recording_rule_total{_label="x"} 2`)
	assert.NotContains(t, output, "1bad")
}

// registryNames 返回当前注册表指标名（测试辅助）。
func registryNames() []string {
	names := make([]string, 0, 8)
	registry.Range(func(key, _ any) bool {
		if name, ok := key.(string); ok {
			names = append(names, name)
		}
		return true
	})
	return names
}
