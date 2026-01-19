package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/cachemgr"
)

// CacheManager 缓存管理器接口
type CacheManager interface {
	cachemgr.CacheManager
}

// cacheManagerImpl 缓存管理器实现
type cacheManagerImpl struct {
	cachemgr.CacheManager
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(configProvider common.BaseConfigProvider) (CacheManager, error) {
	mgr, err := cachemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &cacheManagerImpl{CacheManager: mgr}, nil
}
