package managers

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/cachemgr"
)

type ICacheManager interface {
	cachemgr.ICacheManager
}

type cacheManagerImpl struct {
	cachemgr.ICacheManager
}

func NewCacheManager(configProvider common.IBaseConfigProvider) (ICacheManager, error) {
	mgr, err := cachemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}
	return &cacheManagerImpl{mgr}, nil
}
