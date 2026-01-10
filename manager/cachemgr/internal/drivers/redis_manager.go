package drivers

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/cachemgr/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisManager Redis 缓存管理器
// 使用 go-redis v9 客户端连接 Redis 服务器
type RedisManager struct {
	*BaseManager
	client *redis.Client
}

// NewRedisManager 创建 Redis 缓存管理器
func NewRedisManager(cfg *config.RedisConfig) (*RedisManager, error) {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxIdleConns:    cfg.MaxIdleConns,
		MaxActiveConns:  cfg.MaxOpenConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisManager{
		BaseManager: NewBaseManager("redis-cache"),
		client:      client,
	}, nil
}

// Get 获取缓存值
func (m *RedisManager) Get(ctx context.Context, key string, dest any) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}

	// 从 Redis 获取数据
	data, err := m.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key: %w", err)
	}

	// 反序列化
	if err := deserialize(data, dest); err != nil {
		return fmt.Errorf("failed to deserialize value: %w", err)
	}

	return nil
}

// Set 设置缓存值
func (m *RedisManager) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}

	// 序列化
	data, err := serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// 设置到 Redis
	return m.client.Set(ctx, key, data, expiration).Err()
}

// SetNX 仅当键不存在时才设置值
func (m *RedisManager) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	if err := ValidateContext(ctx); err != nil {
		return false, err
	}
	if err := ValidateKey(key); err != nil {
		return false, err
	}

	// 序列化
	data, err := serialize(value)
	if err != nil {
		return false, fmt.Errorf("failed to serialize value: %w", err)
	}

	// 设置到 Redis
	result, err := m.client.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set key: %w", err)
	}

	return result, nil
}

// Delete 删除缓存值
func (m *RedisManager) Delete(ctx context.Context, key string) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}

	return m.client.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func (m *RedisManager) Exists(ctx context.Context, key string) (bool, error) {
	if err := ValidateContext(ctx); err != nil {
		return false, err
	}
	if err := ValidateKey(key); err != nil {
		return false, err
	}

	result, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return result > 0, nil
}

// Expire 设置过期时间
func (m *RedisManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}

	return m.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (m *RedisManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	if err := ValidateContext(ctx); err != nil {
		return 0, err
	}
	if err := ValidateKey(key); err != nil {
		return 0, err
	}

	ttl, err := m.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get ttl: %w", err)
	}

	return ttl, nil
}

// Clear 清空所有缓存（慎用）
func (m *RedisManager) Clear(ctx context.Context) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}

	// 使用 FlushDB 清空当前数据库
	return m.client.FlushDB(ctx).Err()
}

// GetMultiple 批量获取
func (m *RedisManager) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	if err := ValidateContext(ctx); err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return make(map[string]any), nil
	}

	// 使用 MGET 批量获取
	values, err := m.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	result := make(map[string]any)
	for i, key := range keys {
		value := values[i]
		if value != nil {
			// 反序列化
			if data, ok := value.([]byte); ok {
				var dest any
				if err := deserialize(data, &dest); err == nil {
					result[key] = dest
				} else {
					result[key] = value
				}
			} else if str, ok := value.(string); ok {
				result[key] = str
			} else {
				result[key] = value
			}
		}
	}

	return result, nil
}

// SetMultiple 批量设置
func (m *RedisManager) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	// 使用 Pipeline 批量设置
	pipe := m.client.Pipeline()

	for key, value := range items {
		data, err := serialize(value)
		if err != nil {
			return fmt.Errorf("failed to serialize value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, expiration)
	}

	// 执行所有命令
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to set multiple keys: %w", err)
	}

	return nil
}

// DeleteMultiple 批量删除
func (m *RedisManager) DeleteMultiple(ctx context.Context, keys []string) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return m.client.Del(ctx, keys...).Err()
}

// Increment 自增
func (m *RedisManager) Increment(ctx context.Context, key string, value int64) (int64, error) {
	if err := ValidateContext(ctx); err != nil {
		return 0, err
	}
	if err := ValidateKey(key); err != nil {
		return 0, err
	}

	result, err := m.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}

	return result, nil
}

// Decrement 自减
func (m *RedisManager) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	if err := ValidateContext(ctx); err != nil {
		return 0, err
	}
	if err := ValidateKey(key); err != nil {
		return 0, err
	}

	result, err := m.client.DecrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement: %w", err)
	}

	return result, nil
}

// Close 关闭 Redis 连接
func (m *RedisManager) Close() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}

// Health 检查 Redis 健康状态
func (m *RedisManager) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Ping(ctx).Err()
}

// Ensure RedisManager implements common.Manager interface
var _ common.Manager = (*RedisManager)(nil)

// 序列化函数
func serialize(value any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 反序列化函数
func deserialize(data []byte, dest any) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(dest); err != nil {
		return err
	}
	return nil
}
