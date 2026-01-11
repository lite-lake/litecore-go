package drivers

import (
	"context"
	"fmt"
	"time"

	"com.litelake.litecore/common"
)

// NoneManager 空缓存管理器
// 用于降级场景，提供空实现以避免条件判断
type NoneManager struct {
	*BaseManager
}

// NewNoneManager 创建空缓存管理器
func NewNoneManager() *NoneManager {
	return &NoneManager{
		BaseManager: NewBaseManager("none-cache"),
	}
}

// Get 获取缓存值（返回错误）
func (m *NoneManager) Get(ctx context.Context, key string, dest any) error {
	return fmt.Errorf("cache not available")
}

// Set 设置缓存值（空操作）
func (m *NoneManager) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	// 空操作，不返回错误
	return nil
}

// SetNX 仅当键不存在时才设置值（返回 false，表示未设置）
func (m *NoneManager) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	// 返回 false 表示未设置成功
	return false, nil
}

// Delete 删除缓存值（空操作）
func (m *NoneManager) Delete(ctx context.Context, key string) error {
	return nil
}

// Exists 检查键是否存在（返回 false）
func (m *NoneManager) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// Expire 设置过期时间（空操作）
func (m *NoneManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

// TTL 获取剩余过期时间（返回 0）
func (m *NoneManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

// Clear 清空所有缓存（空操作）
func (m *NoneManager) Clear(ctx context.Context) error {
	return nil
}

// GetMultiple 批量获取（返回空 map）
func (m *NoneManager) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	return make(map[string]any), nil
}

// SetMultiple 批量设置（空操作）
func (m *NoneManager) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	return nil
}

// DeleteMultiple 批量删除（空操作）
func (m *NoneManager) DeleteMultiple(ctx context.Context, keys []string) error {
	return nil
}

// Increment 自增（返回 0）
func (m *NoneManager) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}

// Decrement 自减（返回 0）
func (m *NoneManager) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}

// Ensure NoneManager implements common.BaseManager interface
var _ common.BaseManager = (*NoneManager)(nil)
