package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/cachemgr"
)

// ICacheManager 缓存管理器接口
type ICacheManager interface {
	cachemgr.ICacheManager
}

// cacheManagerImpl 缓存管理器实现
type cacheManagerImpl struct {
	cachemgr.ICacheManager
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(configProvider common.IBaseConfigProvider) (ICacheManager, error) {
	mgr, err := cachemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &cacheManagerImpl{mgr}, nil
}
