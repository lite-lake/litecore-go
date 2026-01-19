package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/cachemgr"
)

// NewCacheManager 创建缓存管理器
func NewCacheManager(configProvider common.BaseConfigProvider) (cachemgr.CacheManager, error) {
	return cachemgr.BuildWithConfigProvider(configProvider)
}
