package cachemgr

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// cacheManagerRedisImpl Redis 缓存实现
type cacheManagerRedisImpl struct {
	*cacheManagerBaseImpl
	client *redis.Client
	name   string
}

// NewCacheManagerRedisImpl 创建 Redis 实现
func NewCacheManagerRedisImpl(cfg *RedisConfig) (CacheManager, error) {
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

	impl := &cacheManagerRedisImpl{
		cacheManagerBaseImpl: newCacheManagerBaseImpl(),
		client:               client,
		name:                 "cacheManagerRedisImpl",
	}
	impl.initObservability()
	return impl, nil
}

// ManagerName 返回管理器名称
func (r *cacheManagerRedisImpl) ManagerName() string {
	return r.name
}

// Health 检查管理器健康状态
func (r *cacheManagerRedisImpl) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Ping(ctx).Err()
}

// OnStart 在服务器启动时触发
func (r *cacheManagerRedisImpl) OnStart() error {
	return nil // Redis 已在构造时连接
}

// OnStop 在服务器停止时触发
func (r *cacheManagerRedisImpl) OnStop() error {
	return r.Close()
}

// Get 获取缓存值
func (r *cacheManagerRedisImpl) Get(ctx context.Context, key string, dest any) error {
	var hit bool
	var getErr error

	err := r.recordOperation(ctx, r.name, "get", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		// 从 Redis 获取数据
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				return fmt.Errorf("key not found: %s", key)
			}
			return fmt.Errorf("failed to get key: %w", err)
		}

		// 反序列化
		getErr = deserialize(data, dest)
		if getErr != nil {
			return fmt.Errorf("failed to deserialize value: %w", getErr)
		}

		hit = true
		return nil
	})

	r.recordCacheHit(ctx, r.name, hit)
	return err
}

// Set 设置缓存值
func (r *cacheManagerRedisImpl) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return r.recordOperation(ctx, r.name, "set", key, func() error {
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
		return r.client.Set(ctx, key, data, expiration).Err()
	})
}

// SetNX 仅当键不存在时才设置值
func (r *cacheManagerRedisImpl) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	var result bool
	var resultErr error

	err := r.recordOperation(ctx, r.name, "setnx", key, func() error {
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
		result, resultErr = r.client.SetNX(ctx, key, data, expiration).Result()
		if resultErr != nil {
			return fmt.Errorf("failed to set key: %w", resultErr)
		}

		return nil
	})

	return result, err
}

// Delete 删除缓存值
func (r *cacheManagerRedisImpl) Delete(ctx context.Context, key string) error {
	return r.recordOperation(ctx, r.name, "delete", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		return r.client.Del(ctx, key).Err()
	})
}

// Exists 检查键是否存在
func (r *cacheManagerRedisImpl) Exists(ctx context.Context, key string) (bool, error) {
	var result bool

	err := r.recordOperation(ctx, r.name, "exists", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		r, err := r.client.Exists(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("failed to check key existence: %w", err)
		}

		result = r > 0
		return nil
	})

	return result, err
}

// Expire 设置过期时间
func (r *cacheManagerRedisImpl) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.recordOperation(ctx, r.name, "expire", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		return r.client.Expire(ctx, key, expiration).Err()
	})
}

// TTL 获取剩余过期时间
func (r *cacheManagerRedisImpl) TTL(ctx context.Context, key string) (time.Duration, error) {
	var result time.Duration

	err := r.recordOperation(ctx, r.name, "ttl", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("failed to get ttl: %w", err)
		}

		result = ttl
		return nil
	})

	return result, err
}

// Clear 清空所有缓存
func (r *cacheManagerRedisImpl) Clear(ctx context.Context) error {
	return r.recordOperation(ctx, r.name, "clear", "", func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}

		// 使用 FlushDB 清空当前数据库
		return r.client.FlushDB(ctx).Err()
	})
}

// GetMultiple 批量获取
func (r *cacheManagerRedisImpl) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	var result map[string]any

	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	err := r.recordOperation(ctx, r.name, "getmultiple", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}

		if len(keys) == 0 {
			result = make(map[string]any)
			return nil
		}

		// 使用 MGET 批量获取
		values, err := r.client.MGet(ctx, keys...).Result()
		if err != nil {
			return fmt.Errorf("failed to get multiple keys: %w", err)
		}

		result = make(map[string]any)
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

		return nil
	})

	return result, err
}

// SetMultiple 批量设置
func (r *cacheManagerRedisImpl) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	key := "batch"
	for k := range items {
		key = k
		break
	}

	return r.recordOperation(ctx, r.name, "setmultiple", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		// 使用 Pipeline 批量设置
		pipe := r.client.Pipeline()

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
	})
}

// DeleteMultiple 批量删除
func (r *cacheManagerRedisImpl) DeleteMultiple(ctx context.Context, keys []string) error {
	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	return r.recordOperation(ctx, r.name, "deletemultiple", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}

		if len(keys) == 0 {
			return nil
		}

		return r.client.Del(ctx, keys...).Err()
	})
}

// Increment 自增
func (r *cacheManagerRedisImpl) Increment(ctx context.Context, key string, value int64) (int64, error) {
	var result int64

	err := r.recordOperation(ctx, r.name, "increment", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		r, err := r.client.IncrBy(ctx, key, value).Result()
		if err != nil {
			return fmt.Errorf("failed to increment: %w", err)
		}

		result = r
		return nil
	})

	return result, err
}

// Decrement 自减
func (r *cacheManagerRedisImpl) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	var result int64

	err := r.recordOperation(ctx, r.name, "decrement", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		r, err := r.client.DecrBy(ctx, key, value).Result()
		if err != nil {
			return fmt.Errorf("failed to decrement: %w", err)
		}

		result = r
		return nil
	})

	return result, err
}

// Close 关闭 Redis 连接
func (r *cacheManagerRedisImpl) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

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

// 确保 cacheManagerRedisImpl 实现 CacheManager 接口
var _ CacheManager = (*cacheManagerRedisImpl)(nil)
