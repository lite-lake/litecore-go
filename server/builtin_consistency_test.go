package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/manager/configmgr"
)

// mockBuiltinConfigManager 一致性校验测试用的配置管理器桩
type mockBuiltinConfigManager struct {
	data map[string]any
}

func (m *mockBuiltinConfigManager) ManagerName() string { return "mockBuiltinConfigManager" }
func (m *mockBuiltinConfigManager) Health() error       { return nil }
func (m *mockBuiltinConfigManager) OnStart() error      { return nil }
func (m *mockBuiltinConfigManager) OnStop() error       { return nil }
func (m *mockBuiltinConfigManager) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}
func (m *mockBuiltinConfigManager) Get(key string) (any, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

var _ configmgr.IConfigManager = (*mockBuiltinConfigManager)(nil)

// redisSection 组装 cache/lock/limiter 三段 redis 配置的测试数据
func redisSection(cacheDriver string, cache map[string]any, lockDriver string, lock map[string]any, limiterDriver string, limiter map[string]any) *mockBuiltinConfigManager {
	data := map[string]any{
		"cache.driver":   cacheDriver,
		"lock.driver":    lockDriver,
		"limiter.driver": limiterDriver,
	}
	if cache != nil {
		data["cache.redis_config"] = cache
	}
	if lock != nil {
		data["lock.redis_config"] = lock
	}
	if limiter != nil {
		data["limiter.redis_config"] = limiter
	}
	return &mockBuiltinConfigManager{data: data}
}

func redisMap(host string, port int, db int) map[string]any {
	return map[string]any{"host": host, "port": port, "db": db, "password": "${REDIS_PASSWORD}"}
}

// TestValidateRedisConfigConsistency 测试 lock/limiter 与 cache 的 redis_config 一致性校验（内部PRD C组）：
// lock/limiter 驱动实际复用全局 cacheMgr 连接，redis_config 声明不一致即「配置说谎」，启动期拒绝。
func TestValidateRedisConfigConsistency(t *testing.T) {
	t.Run("三方完全一致校验通过", func(t *testing.T) {
		cfg := redisSection("redis", redisMap("redis.example.com", 6379, 2),
			"redis", redisMap("redis.example.com", 6379, 2),
			"redis", redisMap("redis.example.com", 6379, 2))
		assert.NoError(t, validateRedisConfigConsistency(cfg))
	})

	t.Run("cache 非 redis 时跳过比对（契约违反由工厂层拦截）", func(t *testing.T) {
		cfg := redisSection("memory", nil, "memory", nil, "memory", nil)
		assert.NoError(t, validateRedisConfigConsistency(cfg))
	})

	t.Run("lock 驱动为 memory 时跳过 lock 比对", func(t *testing.T) {
		cfg := redisSection("redis", redisMap("h", 6379, 1),
			"memory", nil,
			"redis", redisMap("h", 6379, 1))
		assert.NoError(t, validateRedisConfigConsistency(cfg))
	})

	t.Run("lock 与 cache 的 db 不一致报错并列出两侧实际值", func(t *testing.T) {
		cfg := redisSection("redis", redisMap("h1", 6379, 1),
			"redis", redisMap("h1", 6379, 2),
			"memory", nil)
		err := validateRedisConfigConsistency(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock.redis_config 与 cache.redis_config 不一致")
		assert.Contains(t, err.Error(), "db: lock=2 cache=1")
	})

	t.Run("lock 与 cache 的 host 不一致报错", func(t *testing.T) {
		cfg := redisSection("redis", redisMap("host-a", 6379, 1),
			"redis", redisMap("host-b", 6379, 1),
			"memory", nil)
		err := validateRedisConfigConsistency(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `host: lock="host-b" cache="host-a"`)
	})

	t.Run("lock 与 cache 的 port 不一致报错", func(t *testing.T) {
		cfg := redisSection("redis", redisMap("h", 6379, 1),
			"redis", redisMap("h", 6380, 1),
			"memory", nil)
		err := validateRedisConfigConsistency(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port: lock=6380 cache=6379")
	})

	t.Run("limiter 与 cache 的 db 不一致报错", func(t *testing.T) {
		cfg := redisSection("redis", redisMap("h", 6379, 8),
			"memory", nil,
			"redis", redisMap("h", 6379, 9))
		err := validateRedisConfigConsistency(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limiter.redis_config 与 cache.redis_config 不一致")
		assert.Contains(t, err.Error(), "db: limiter=9 cache=8")
	})

	t.Run("缺省键回落默认值不误报（lock 未写 port/db）", func(t *testing.T) {
		cfg := redisSection("redis", map[string]any{"host": "h", "port": 6379, "db": 0},
			"redis", map[string]any{"host": "h"},
			"memory", nil)
		assert.NoError(t, validateRedisConfigConsistency(cfg))
	})

	t.Run("secrets 占位符两侧字面一致即通过", func(t *testing.T) {
		cfg := redisSection("redis", redisMap("${REDIS_HOST}", 6379, 1),
			"redis", redisMap("${REDIS_HOST}", 6379, 1),
			"redis", redisMap("${REDIS_HOST}", 6379, 1))
		assert.NoError(t, validateRedisConfigConsistency(cfg))
	})

	t.Run("错误信息含修复指引", func(t *testing.T) {
		cfg := redisSection("redis", redisMap("h", 6379, 1),
			"redis", redisMap("h", 6379, 2),
			"memory", nil)
		err := validateRedisConfigConsistency(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "middleware-redis.md")
	})
}
