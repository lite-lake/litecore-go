package managers

import (
	"fmt"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/manager/databasemgr"
)

type IDatabaseManager interface {
	databasemgr.IDatabaseManager
}

type databaseManagerImpl struct {
	databasemgr.IDatabaseManager
}

func NewDatabaseManager(configProvider common.IBaseConfigProvider) (IDatabaseManager, error) {
	fmt.Println("[DEBUG] NewDatabaseManager called")
	mgr, err := databasemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		fmt.Printf("[DEBUG] NewDatabaseManager error: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] NewDatabaseManager success: type=%T\n", mgr)
	return &databaseManagerImpl{mgr}, nil
}
