package infras

import (
	"fmt"

	"com.litelake.litecore/common"
	"com.litelake.litecore/component/manager/databasemgr"
)

// DatabaseManager 数据库管理器接口
type DatabaseManager interface {
	databasemgr.DatabaseManager
}

// databaseManagerImpl 数据库管理器实现
type databaseManagerImpl struct {
	databasemgr.DatabaseManager
}

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager(configProvider common.BaseConfigProvider) (DatabaseManager, error) {
	fmt.Println("[DEBUG] NewDatabaseManager called")
	mgr, err := databasemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		fmt.Printf("[DEBUG] NewDatabaseManager error: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] NewDatabaseManager success: type=%T\n", mgr)
	return &databaseManagerImpl{DatabaseManager: mgr}, nil
}
