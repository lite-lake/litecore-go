package litebloomsvc

import (
	"fmt"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"

	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

// ILiteBloomService 布隆过滤器服务接口
type ILiteBloomService interface {
	// 创建过滤器
	CreateFilter(name string) error
	CreateFilterWithConfig(name string, config *FilterConfig) error

	// 删除过滤器
	DeleteFilter(name string) error

	// 获取过滤器统计信息
	GetStats(name string) (*FilterStats, error)

	// 列出所有过滤器
	ListFilters() []string

	// 添加元素
	Add(name string, data []byte) error
	AddString(name string, data string) error
	AddBatch(name string, dataList [][]byte) error
	AddStringBatch(name string, dataList []string) error

	// 检查元素
	Contains(name string, data []byte) (bool, error)
	ContainsString(name string, data string) (bool, error)

	// 重建过滤器
	Rebuild(name string) error
	RebuildWithConfig(name string, config *FilterConfig) error

	// 生命周期
	OnStart() error
	OnStop() error
}

// filterWrapper 包装布隆过滤器和元数据
type filterWrapper struct {
	filter            *bloom.BloomFilter
	expectedItems     uint
	falsePositiveRate float64
	createdAt         time.Time
	expiresAt         time.Time
	ttl               time.Duration
	mu                sync.RWMutex
}

// liteBloomServiceImpl 布隆过滤器服务实现
type liteBloomServiceImpl struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
	config    *Config
	filters   map[string]*filterWrapper
	mu        sync.RWMutex
	stopCh    chan struct{}
}

// NewLiteBloomService 使用默认配置创建布隆过滤器服务
func NewLiteBloomService() ILiteBloomService {
	return &liteBloomServiceImpl{
		config:  DefaultConfig(),
		filters: make(map[string]*filterWrapper),
		stopCh:  make(chan struct{}),
	}
}

// NewLiteBloomServiceWithConfig 使用自定义配置创建布隆过滤器服务
func NewLiteBloomServiceWithConfig(config *Config) ILiteBloomService {
	cfg := config
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &liteBloomServiceImpl{
		config:  cfg,
		filters: make(map[string]*filterWrapper),
		stopCh:  make(chan struct{}),
	}
}

// CreateFilter 使用默认配置创建过滤器
func (s *liteBloomServiceImpl) CreateFilter(name string) error {
	return s.CreateFilterWithConfig(name, nil)
}

// CreateFilterWithConfig 使用自定义配置创建过滤器
func (s *liteBloomServiceImpl) CreateFilterWithConfig(name string, config *FilterConfig) error {
	if name == "" {
		return fmt.Errorf("过滤器名称不能为空")
	}

	expectedItems := s.config.getExpectedItems(config)
	falsePositiveRate := s.config.getFalsePositiveRate(config)
	ttl := s.config.getTTL(config)

	filter := bloom.NewWithEstimates(expectedItems, falsePositiveRate)

	now := time.Now()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}

	wrapper := &filterWrapper{
		filter:            filter,
		expectedItems:     expectedItems,
		falsePositiveRate: falsePositiveRate,
		createdAt:         now,
		expiresAt:         expiresAt,
		ttl:               ttl,
	}

	s.mu.Lock()
	s.filters[name] = wrapper
	s.mu.Unlock()

	if s.LoggerMgr != nil {
		s.LoggerMgr.Ins().Info("布隆过滤器已创建",
			"name", name,
			"expected_items", expectedItems,
			"false_positive_rate", falsePositiveRate,
			"ttl", ttl,
		)
	}

	return nil
}

// DeleteFilter 删除过滤器
func (s *liteBloomServiceImpl) DeleteFilter(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.filters[name]; !exists {
		return fmt.Errorf("过滤器不存在: %s", name)
	}

	delete(s.filters, name)

	if s.LoggerMgr != nil {
		s.LoggerMgr.Ins().Info("布隆过滤器已删除", "name", name)
	}

	return nil
}

// GetStats 获取过滤器统计信息
func (s *liteBloomServiceImpl) GetStats(name string) (*FilterStats, error) {
	s.mu.RLock()
	wrapper, exists := s.filters[name]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("过滤器不存在: %s", name)
	}

	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()

	elementCount := uint(wrapper.filter.ApproximatedSize())
	return &FilterStats{
		Name:              name,
		ExpectedItems:     wrapper.expectedItems,
		FalsePositiveRate: wrapper.falsePositiveRate,
		BitSize:           wrapper.filter.Cap(),
		HashFunctions:     wrapper.filter.K(),
		ElementCount:      elementCount,
		FillRatio:         float64(elementCount) / float64(wrapper.expectedItems),
		CreatedAt:         wrapper.createdAt,
		ExpiresAt:         wrapper.expiresAt,
		TTL:               wrapper.ttl,
	}, nil
}

// ListFilters 列出所有过滤器名称
func (s *liteBloomServiceImpl) ListFilters() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.filters))
	for name := range s.filters {
		names = append(names, name)
	}
	return names
}

// Add 添加元素到过滤器
func (s *liteBloomServiceImpl) Add(name string, data []byte) error {
	s.mu.RLock()
	wrapper, exists := s.filters[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("过滤器不存在: %s", name)
	}

	if err := s.checkAndRebuildIfNeeded(name, wrapper); err != nil {
		return err
	}

	wrapper.mu.Lock()
	wrapper.filter.Add(data)
	wrapper.mu.Unlock()

	return nil
}

// AddString 添加字符串元素到过滤器
func (s *liteBloomServiceImpl) AddString(name string, data string) error {
	return s.Add(name, []byte(data))
}

// AddBatch 批量添加元素到过滤器
func (s *liteBloomServiceImpl) AddBatch(name string, dataList [][]byte) error {
	s.mu.RLock()
	wrapper, exists := s.filters[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("过滤器不存在: %s", name)
	}

	if err := s.checkAndRebuildIfNeeded(name, wrapper); err != nil {
		return err
	}

	wrapper.mu.Lock()
	for _, data := range dataList {
		wrapper.filter.Add(data)
	}
	wrapper.mu.Unlock()

	return nil
}

// AddStringBatch 批量添加字符串元素到过滤器
func (s *liteBloomServiceImpl) AddStringBatch(name string, dataList []string) error {
	byteList := make([][]byte, len(dataList))
	for i, data := range dataList {
		byteList[i] = []byte(data)
	}
	return s.AddBatch(name, byteList)
}

// Contains 检查元素是否可能存在于过滤器中
func (s *liteBloomServiceImpl) Contains(name string, data []byte) (bool, error) {
	s.mu.RLock()
	wrapper, exists := s.filters[name]
	s.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("过滤器不存在: %s", name)
	}

	if err := s.checkAndRebuildIfNeeded(name, wrapper); err != nil {
		return false, err
	}

	wrapper.mu.RLock()
	result := wrapper.filter.Test(data)
	wrapper.mu.RUnlock()

	return result, nil
}

// ContainsString 检查字符串元素是否可能存在于过滤器中
func (s *liteBloomServiceImpl) ContainsString(name string, data string) (bool, error) {
	return s.Contains(name, []byte(data))
}

// Rebuild 重建过滤器
func (s *liteBloomServiceImpl) Rebuild(name string) error {
	return s.RebuildWithConfig(name, nil)
}

// RebuildWithConfig 使用新配置重建过滤器
func (s *liteBloomServiceImpl) RebuildWithConfig(name string, config *FilterConfig) error {
	s.mu.RLock()
	_, exists := s.filters[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("过滤器不存在: %s", name)
	}

	return s.CreateFilterWithConfig(name, config)
}

// checkAndRebuildIfNeeded 检查是否需要重建过滤器
func (s *liteBloomServiceImpl) checkAndRebuildIfNeeded(name string, wrapper *filterWrapper) error {
	if wrapper.ttl <= 0 || wrapper.expiresAt.IsZero() {
		return nil
	}

	if time.Now().After(wrapper.expiresAt) {
		if s.LoggerMgr != nil {
			s.LoggerMgr.Ins().Info("布隆过滤器已过期，自动重建", "name", name)
		}
		return s.CreateFilterWithConfig(name, &FilterConfig{
			ExpectedItems:     &wrapper.expectedItems,
			FalsePositiveRate: &wrapper.falsePositiveRate,
			TTL:               &wrapper.ttl,
		})
	}

	return nil
}

// OnStart 服务启动
func (s *liteBloomServiceImpl) OnStart() error {
	if s.LoggerMgr != nil {
		s.LoggerMgr.Ins().Info("布隆过滤器服务已启动")
	}
	return nil
}

// OnStop 服务停止
func (s *liteBloomServiceImpl) OnStop() error {
	close(s.stopCh)
	s.mu.Lock()
	s.filters = make(map[string]*filterWrapper)
	s.mu.Unlock()

	if s.LoggerMgr != nil {
		s.LoggerMgr.Ins().Info("布隆过滤器服务已停止")
	}
	return nil
}
