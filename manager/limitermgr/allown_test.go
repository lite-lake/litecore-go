package limitermgr

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lite-lake/litecore-go/manager/cachemgr"
)

// AllowN 成本语义测试（内部PRD §2.2/§4）：按记录计数、拒绝整批不消耗、
// 空批零消耗、精确 429 分界；memory 与 redis 双实现对拍（同一用例体跑两实现，
// 语义必须一致）。redis 实现按既有测试先例以 memory cachemgr 为底座。

func newAllowNManagers(t *testing.T) map[string]ILimiterManager {
	t.Helper()
	redis := NewLimiterManagerRedisImpl(nil, nil, cachemgr.NewCacheManagerMemoryImpl(time.Minute, time.Minute, nil, nil))
	require.NoError(t, redis.OnStart())
	memory := NewLimiterManagerMemoryImpl(nil, nil)
	require.NoError(t, memory.OnStart())
	return map[string]ILimiterManager{"memory": memory, "redis": redis}
}

// TestAllowN_按记录计数精确分界 单批 460 条耗 460 配额：3600 配额下前 7 批通过
// （3220），第 8 批整批 429 分界精确落在记录数而非批数
func TestAllowN_按记录计数精确分界(t *testing.T) {
	for name, mgr := range newAllowNManagers(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			key := "allown-boundary"
			for i := 0; i < 7; i++ {
				allowed, err := mgr.AllowN(ctx, key, 3600, time.Minute, 460)
				require.NoError(t, err)
				assert.True(t, allowed, "第%d批(460条)应允许", i+1)
			}
			// 3220 + 460 = 3680 > 3600 → 整批拒绝
			allowed, err := mgr.AllowN(ctx, key, 3600, time.Minute, 460)
			require.NoError(t, err)
			assert.False(t, allowed, "第8批应整批拒绝")

			// 拒绝不消耗：剩余 380 → cost=380 通过、381 拒绝
			allowed, err = mgr.AllowN(ctx, key, 3600, time.Minute, 380)
			require.NoError(t, err)
			assert.True(t, allowed, "拒绝路径零消耗：剩余380应允许")
			allowed, err = mgr.AllowN(ctx, key, 3600, time.Minute, 1)
			require.NoError(t, err)
			assert.False(t, allowed, "3600 配额耗尽后应拒绝")
		})
	}
}

// TestAllowN_空批与跨批累计 cost=0 恒允许且不消耗；cost 跨批累计
func TestAllowN_空批与跨批累计(t *testing.T) {
	for name, mgr := range newAllowNManagers(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			key := "allown-empty"

			// 空批：未建键时也允许
			allowed, err := mgr.AllowN(ctx, key, 10, time.Minute, 0)
			require.NoError(t, err)
			assert.True(t, allowed, "空批(未建键)应允许")

			// 跨批累计：5+5 用满 10，第 11 条拒绝
			allowed, err = mgr.AllowN(ctx, key, 10, time.Minute, 5)
			require.NoError(t, err)
			assert.True(t, allowed)
			allowed, err = mgr.AllowN(ctx, key, 10, time.Minute, 5)
			require.NoError(t, err)
			assert.True(t, allowed)

			// 配额已尽：空批仍允许且不消耗（不因满额拒绝空批），实体记录被拒
			allowed, err = mgr.AllowN(ctx, key, 10, time.Minute, 0)
			require.NoError(t, err)
			assert.True(t, allowed, "满额时空批仍应允许")
			allowed, err = mgr.AllowN(ctx, key, 10, time.Minute, 1)
			require.NoError(t, err)
			assert.False(t, allowed, "配额尽后实体记录应拒绝")
			remaining, err := mgr.GetRemaining(ctx, key, 10, time.Minute)
			require.NoError(t, err)
			assert.Equal(t, 0, remaining, "空批不得消耗配额")
		})
	}
}

// TestAllowN_窗口恢复 短窗口过期后配额恢复（memory 滑动窗口 / redis 固定窗口
// 恢复时点不同，取足够长的 sleep 保证双实现都过期）
func TestAllowN_窗口恢复(t *testing.T) {
	for name, mgr := range newAllowNManagers(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			key := "allown-recover"
			allowed, err := mgr.AllowN(ctx, key, 3, 80*time.Millisecond, 3)
			require.NoError(t, err)
			assert.True(t, allowed)
			allowed, err = mgr.AllowN(ctx, key, 3, 80*time.Millisecond, 1)
			require.NoError(t, err)
			assert.False(t, allowed, "窗口内超配应拒绝")

			time.Sleep(150 * time.Millisecond)

			allowed, err = mgr.AllowN(ctx, key, 3, 80*time.Millisecond, 3)
			require.NoError(t, err)
			assert.True(t, allowed, "窗口过期后配额应恢复")
		})
	}
}

// TestAllowN_等价Allow Allow ≡ AllowN(cost=1)；cost<0 参数非法
func TestAllowN_等价Allow(t *testing.T) {
	for name, mgr := range newAllowNManagers(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			key := "allown-equiv"
			for i := 0; i < 3; i++ {
				allowed, err := mgr.Allow(ctx, key, 3, time.Minute)
				require.NoError(t, err)
				assert.True(t, allowed)
				allowedN, err := mgr.AllowN(ctx, key+"-n", 3, time.Minute, 1)
				require.NoError(t, err)
				assert.True(t, allowedN)
			}
			allowed, err := mgr.Allow(ctx, key, 3, time.Minute)
			require.NoError(t, err)
			assert.False(t, allowed)
			allowed, err = mgr.AllowN(ctx, key+"-n", 3, time.Minute, 1)
			require.NoError(t, err)
			assert.False(t, allowed)

			_, err = mgr.AllowN(ctx, key+"-neg", 3, time.Minute, -1)
			assert.Error(t, err, "cost<0 应参数非法")
		})
	}
}
