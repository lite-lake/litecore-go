package litebloomsvc

import (
	"time"
)

// FilterConfig 单个布隆过滤器的配置
type FilterConfig struct {
	ExpectedItems     *uint          // 预期元素数量
	FalsePositiveRate *float64       // 误判率（0-1之间）
	TTL               *time.Duration // 生存时间（0 表示永不过期）
}

// FilterStats 布隆过滤器的统计信息
type FilterStats struct {
	Name              string        // 过滤器名称
	ExpectedItems     uint          // 预期元素数量
	FalsePositiveRate float64       // 误判率
	BitSize           uint          // 位数组大小
	HashFunctions     uint          // 哈希函数数量
	ElementCount      uint          // 已添加元素数量
	FillRatio         float64       // 填充率（0-1）
	CreatedAt         time.Time     // 创建时间
	ExpiresAt         time.Time     // 过期时间（零值表示永不过期）
	TTL               time.Duration // 生存时间
}

// Config 布隆过滤器服务的全局配置
type Config struct {
	DefaultExpectedItems     *uint          // 默认预期元素数量
	DefaultFalsePositiveRate *float64       // 默认误判率
	DefaultTTL               *time.Duration // 默认生存时间
}

// 默认配置值
const (
	defaultExpectedItems     uint    = 10000
	defaultFalsePositiveRate float64 = 0.01
)

// DefaultConfig 返回默认服务配置
func DefaultConfig() *Config {
	return &Config{}
}

// DefaultFilterConfig 返回默认过滤器配置
func DefaultFilterConfig() *FilterConfig {
	return &FilterConfig{}
}

// getExpectedItems 获取预期元素数量，如果未配置则使用默认值或服务默认值
func (c *Config) getExpectedItems(filterCfg *FilterConfig) uint {
	if filterCfg != nil && filterCfg.ExpectedItems != nil {
		return *filterCfg.ExpectedItems
	}
	if c != nil && c.DefaultExpectedItems != nil {
		return *c.DefaultExpectedItems
	}
	return defaultExpectedItems
}

// getFalsePositiveRate 获取误判率，如果未配置则使用默认值或服务默认值
func (c *Config) getFalsePositiveRate(filterCfg *FilterConfig) float64 {
	if filterCfg != nil && filterCfg.FalsePositiveRate != nil {
		return *filterCfg.FalsePositiveRate
	}
	if c != nil && c.DefaultFalsePositiveRate != nil {
		return *c.DefaultFalsePositiveRate
	}
	return defaultFalsePositiveRate
}

// getTTL 获取 TTL，如果未配置则使用服务默认值
func (c *Config) getTTL(filterCfg *FilterConfig) time.Duration {
	if filterCfg != nil && filterCfg.TTL != nil {
		return *filterCfg.TTL
	}
	if c != nil && c.DefaultTTL != nil {
		return *c.DefaultTTL
	}
	return 0
}

// 辅助函数：创建指针
func ptrUint(v uint) *uint {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrDuration(v time.Duration) *time.Duration {
	return &v
}
