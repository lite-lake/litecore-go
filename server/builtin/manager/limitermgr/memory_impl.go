package limitermgr

import (
	"context"
	"sync"
	"time"
)

type limiterEntry struct {
	mu        sync.RWMutex
	window    []time.Time
	limit     int
	windowDur time.Duration
}

type limiterManagerMemoryImpl struct {
	*limiterManagerBaseImpl
	limiters sync.Map
	name     string
}

func NewLimiterManagerMemoryImpl() ILimiterManager {
	impl := &limiterManagerMemoryImpl{
		limiterManagerBaseImpl: newILimiterManagerBaseImpl(),
		limiters:               sync.Map{},
		name:                   "limiterManagerMemoryImpl",
	}
	impl.initObservability()
	return impl
}

func (m *limiterManagerMemoryImpl) ManagerName() string {
	return m.name
}

func (m *limiterManagerMemoryImpl) Health() error {
	return nil
}

func (m *limiterManagerMemoryImpl) OnStart() error {
	return nil
}

func (m *limiterManagerMemoryImpl) OnStop() error {
	return nil
}

func (m *limiterManagerMemoryImpl) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	var result bool

	err := m.recordOperation(ctx, "memory", "allow", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}
		if err := ValidateLimit(limit); err != nil {
			return err
		}
		if err := ValidateWindow(window); err != nil {
			return err
		}

		value, _ := m.limiters.LoadOrStore(key, &limiterEntry{
			limit:     limit,
			windowDur: window,
		})
		entry := value.(*limiterEntry)

		entry.mu.Lock()
		now := time.Now()
		entry.limit = limit
		entry.windowDur = window
		cutoff := now.Add(-window)

		validWindow := make([]time.Time, 0, len(entry.window))
		for _, t := range entry.window {
			if t.After(cutoff) {
				validWindow = append(validWindow, t)
			}
		}
		entry.window = validWindow

		if len(entry.window) < limit {
			entry.window = append(entry.window, now)
			result = true
			m.recordAllowance(ctx, "memory", true)
		} else {
			result = false
			m.recordAllowance(ctx, "memory", false)
		}
		entry.mu.Unlock()

		return nil
	})

	return result, err
}

func (m *limiterManagerMemoryImpl) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	var result int

	err := m.recordOperation(ctx, "memory", "get_remaining", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}
		if err := ValidateLimit(limit); err != nil {
			return err
		}
		if err := ValidateWindow(window); err != nil {
			return err
		}

		value, ok := m.limiters.Load(key)
		if !ok {
			result = limit
			return nil
		}

		entry := value.(*limiterEntry)

		entry.mu.RLock()
		now := time.Now()
		entry.limit = limit
		entry.windowDur = window
		cutoff := now.Add(-window)

		count := 0
		for _, t := range entry.window {
			if t.After(cutoff) {
				count++
			}
		}
		result = limit - count
		entry.mu.RUnlock()

		return nil
	})

	return result, err
}

var _ ILimiterManager = (*limiterManagerMemoryImpl)(nil)
