// Package metrics 提供零第三方依赖的进程内指标注册表，按 Prometheus 文本格式输出。
//
// 设计要点：
//   - 指标按 name 注册，支持 counter / gauge 两种类型；未注册的 name 在首次
//     Add/Set 时自动按 counter 注册（help 为空）；
//   - 每条时间序列（name + 标签集合）各自持有独立原子值，累加采用 CAS 重试，
//     注册表与序列集合使用 sync.Map，读取路径不存在全局写锁；
//   - 默认标签在序列创建时合并（显式标签优先），因此 SetDefaultLabels 应在
//     打点之前调用（产品进程在启动期由 privilegemetrics.Init 完成）；
//   - Render 输出 `# HELP` / `# TYPE` 与样本行：指标名升序、标签按字典序，
//     标签值转义反斜杠、双引号与换行，浮点使用最小表示；
//   - 指标名与标签键按 Prometheus 规范校验，非法名称/标签键直接忽略，不产生非法样本。
package metrics

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Type 指标类型。
type Type string

const (
	// Counter 单调递增计数器。
	Counter Type = "counter"
	// Gauge 可增可减的瞬时值。
	Gauge Type = "gauge"
)

// Labels 指标标签集合（只允许有界枚举值，禁止 IP、request_id 等无界维度）。
type Labels map[string]string

// series 单条时间序列：标签集合不可变，值以 float64 位模式原子存储。
type series struct {
	labels Labels
	value  atomic.Uint64
}

// add 以 CAS 方式累加值，避免全局写锁。
func (s *series) add(delta float64) {
	for {
		old := s.value.Load()
		next := math.Float64bits(math.Float64frombits(old) + delta)
		if s.value.CompareAndSwap(old, next) {
			return
		}
	}
}

// store 直接覆盖值。
func (s *series) store(v float64) {
	s.value.Store(math.Float64bits(v))
}

// load 读取当前值。
func (s *series) load() float64 {
	return math.Float64frombits(s.value.Load())
}

// metric 一个已注册指标：name/help/type 在发布到注册表后不再修改。
type metric struct {
	name   string
	help   string
	typ    Type
	series sync.Map // string（标签签名） -> *series
}

// registry 全局指标注册表：name -> *metric。
var registry sync.Map

// defaults 全局默认标签（atomic.Value，存储 Labels；写入时整体替换副本）。
var defaults atomic.Value

// Register 注册指标。
//
// name 不符合 Prometheus 指标名规范（`[a-zA-Z_:][a-zA-Z0-9_:]*`）时直接忽略。
// 重复注册同类型为幂等操作（保留首次的 help 与全部序列）；类型冲突时忽略本次
// 注册、保留首次注册的类型。typ 仅接受 Counter / Gauge，其他值按 Counter 处理。
func Register(name, help string, typ Type) {
	if !validMetricName(name) {
		return
	}
	if typ != Counter && typ != Gauge {
		typ = Counter
	}

	m := &metric{name: name, help: help, typ: typ}
	// 已存在（含 Add/Set 触发的自动注册）时保留首次注册结果。
	registry.LoadOrStore(name, m)
}

// Add 对指标累加 v，labels 与默认标签合并后定位序列；name 未注册时按 counter 自动注册。
//
// name 非法（空或不匹配 `[a-zA-Z_:][a-zA-Z0-9_:]*`）时直接忽略。
func Add(name string, v float64, labels Labels) {
	m := lookupOrRegister(name)
	if m == nil {
		return
	}
	m.seriesFor(labels).add(v)
}

// Set 将指标设为 v，labels 与默认标签合并后定位序列；name 未注册时按 counter 自动注册。
//
// name 非法（空或不匹配 `[a-zA-Z_:][a-zA-Z0-9_:]*`）时直接忽略。
func Set(name string, v float64, labels Labels) {
	m := lookupOrRegister(name)
	if m == nil {
		return
	}
	m.seriesFor(labels).store(v)
}

// SetDefaultLabels 设置全局默认标签（整体替换，传入 map 会被复制）。
//
// 默认标签在序列首次创建时合并，显式传入的同名标签优先；应在打点之前调用。
func SetDefaultLabels(labels Labels) {
	defaults.Store(copyLabels(labels))
}

// ResetForTest 清空注册表与默认标签，仅供测试使用。
func ResetForTest() {
	registry.Clear()
	defaults.Store(Labels{})
}

// Render 输出全部指标的 Prometheus 文本格式内容，空注册表返回空字符串。
func Render() string {
	names := make([]string, 0, 8)
	registry.Range(func(key, _ any) bool {
		if name, ok := key.(string); ok {
			names = append(names, name)
		}
		return true
	})
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		value, ok := registry.Load(name)
		if !ok {
			continue
		}
		m := value.(*metric)

		b.WriteString("# HELP ")
		b.WriteString(m.name)
		b.WriteByte(' ')
		b.WriteString(escapeHelp(m.help))
		b.WriteByte('\n')
		b.WriteString("# TYPE ")
		b.WriteString(m.name)
		b.WriteByte(' ')
		b.WriteString(string(m.typ))
		b.WriteByte('\n')

		type sample struct {
			key    string
			value  float64
			labels Labels
		}
		samples := make([]sample, 0, 8)
		m.series.Range(func(key, value any) bool {
			s, ok := value.(*series)
			if !ok {
				return true
			}
			sig, ok := key.(string)
			if !ok {
				return true
			}
			samples = append(samples, sample{key: sig, value: s.load(), labels: s.labels})
			return true
		})
		sort.Slice(samples, func(i, j int) bool { return samples[i].key < samples[j].key })

		for _, sm := range samples {
			b.WriteString(m.name)
			writeLabels(&b, sm.labels)
			b.WriteByte(' ')
			b.WriteString(formatFloat(sm.value))
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// lookupOrRegister 返回已注册指标；不存在时按 counter（help 为空）自动注册。
// name 非法时返回 nil。
func lookupOrRegister(name string) *metric {
	if !validMetricName(name) {
		return nil
	}
	if value, ok := registry.Load(name); ok {
		return value.(*metric)
	}
	created := &metric{name: name, typ: Counter}
	actual, _ := registry.LoadOrStore(name, created)
	return actual.(*metric)
}

// seriesFor 返回标签对应的序列，不存在时创建；创建时合并默认标签并过滤非法标签键。
func (m *metric) seriesFor(labels Labels) *series {
	merged := mergeLabels(currentDefaults(), labels)
	key, stored := canonicalLabels(merged)

	if value, ok := m.series.Load(key); ok {
		return value.(*series)
	}
	created := &series{labels: stored}
	actual, _ := m.series.LoadOrStore(key, created)
	return actual.(*series)
}

// currentDefaults 返回当前默认标签（只读，不得修改返回值）。
func currentDefaults() Labels {
	value := defaults.Load()
	if value == nil {
		return nil
	}
	labels, _ := value.(Labels)
	return labels
}

// copyLabels 复制标签集合（过滤非法标签键），避免调用方后续修改影响注册表。
func copyLabels(labels Labels) Labels {
	copied := make(Labels, len(labels))
	for key, value := range labels {
		if !validLabelName(key) {
			continue
		}
		copied[key] = value
	}
	return copied
}

// mergeLabels 合并默认标签与显式标签（显式优先），过滤非法标签键。
func mergeLabels(base, override Labels) Labels {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	merged := make(Labels, len(base)+len(override))
	for key, value := range base {
		if !validLabelName(key) {
			continue
		}
		merged[key] = value
	}
	for key, value := range override {
		if !validLabelName(key) {
			continue
		}
		merged[key] = value
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

// canonicalLabels 生成稳定的标签签名与副本：键升序，长度前缀避免拼接歧义。
func canonicalLabels(labels Labels) (string, Labels) {
	if len(labels) == 0 {
		return "", nil
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	stored := make(Labels, len(keys))
	for _, key := range keys {
		value := labels[key]
		stored[key] = value
		b.WriteString(strconv.Itoa(len(key)))
		b.WriteByte(':')
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(strconv.Itoa(len(value)))
		b.WriteByte(':')
		b.WriteString(value)
		b.WriteByte(';')
	}
	return b.String(), stored
}

// writeLabels 按字典序写出 {k="v",...}；无标签时不输出。
func writeLabels(b *strings.Builder, labels Labels) {
	if len(labels) == 0 {
		return
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	b.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(key)
		b.WriteString(`="`)
		b.WriteString(escapeLabelValue(labels[key]))
		b.WriteByte('"')
	}
	b.WriteByte('}')
}

// escapeLabelValue 转义标签值中的反斜杠、双引号与换行。
func escapeLabelValue(value string) string {
	if !strings.ContainsAny(value, "\\\"\n") {
		return value
	}
	var b strings.Builder
	b.Grow(len(value) + 8)
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		default:
			b.WriteByte(value[i])
		}
	}
	return b.String()
}

// escapeHelp 转义 HELP 文本中的反斜杠与换行。
func escapeHelp(help string) string {
	if !strings.ContainsAny(help, "\\\n") {
		return help
	}
	var b strings.Builder
	b.Grow(len(help) + 8)
	for i := 0; i < len(help); i++ {
		switch help[i] {
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		default:
			b.WriteByte(help[i])
		}
	}
	return b.String()
}

// formatFloat 输出最小表示的浮点数；特殊值使用 Prometheus 约定写法。
func formatFloat(value float64) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "+Inf"
	case math.IsInf(value, -1):
		return "-Inf"
	}
	if value == math.Trunc(value) && math.Abs(value) < 1e15 {
		if value == 0 {
			return "0"
		}
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}

// validMetricName 校验 Prometheus 指标名：`[a-zA-Z_:][a-zA-Z0-9_:]*`。
func validMetricName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if i == 0 {
			if !isAlpha(c) && c != '_' && c != ':' {
				return false
			}
			continue
		}
		if !isAlpha(c) && !isDigit(c) && c != '_' && c != ':' {
			return false
		}
	}
	return true
}

// validLabelName 校验 Prometheus 标签键：`[a-zA-Z_][a-zA-Z0-9_]*`。
func validLabelName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if i == 0 {
			if !isAlpha(c) && c != '_' {
				return false
			}
			continue
		}
		if !isAlpha(c) && !isDigit(c) && c != '_' {
			return false
		}
	}
	return true
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
