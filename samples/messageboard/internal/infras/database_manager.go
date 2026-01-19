package infras

import (
	"fmt"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/databasemgr"
)

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager(configProvider common.BaseConfigProvider) (databasemgr.DatabaseManager, error) {
	fmt.Println("[DEBUG] NewDatabaseManager called")
	mgr, err := databasemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		fmt.Printf("[DEBUG] NewDatabaseManager error: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] NewDatabaseManager success: type=%T\n", mgr)
	return mgr, nil
}
