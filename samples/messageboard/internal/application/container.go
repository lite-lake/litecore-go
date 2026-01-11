// Package application 管理应用容器和依赖注入
package application

import (
	"com.litelake.litecore/config"
	"com.litelake.litecore/manager/cachemgr"
	"com.litelake.litecore/manager/databasemgr"
	"com.litelake.litecore/manager/loggermgr"
	"com.litelake.litecore/server"
	"com.litelake.litecore/samples/messageboard/internal/controllers"
	"com.litelake.litecore/samples/messageboard/internal/entities"
	"com.litelake.litecore/samples/messageboard/internal/middlewares"
	"com.litelake.litecore/samples/messageboard/internal/repositories"
	"com.litelake.litecore/samples/messageboard/internal/services"
)

// NewEngine 创建并配置留言板应用引擎
func NewEngine() (*server.Engine, error) {
	// 第一步：创建配置提供者
	configProvider, err := config.NewConfigProvider("yaml", "configs/config.yaml")
	if err != nil {
		return nil, err
	}

	// 第二步：使用工厂方法创建 managers
	dbMgr, err := databasemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}

	cacheMgr, err := cachemgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}

	loggerMgr, err := loggermgr.BuildWithConfigProvider(configProvider)
	if err != nil {
		return nil, err
	}

	// 第三步：一次性创建引擎并注册所有组件
	return server.NewEngine(
		server.WithConfig(configProvider),
		server.RegisterManagers(dbMgr, cacheMgr, loggerMgr),
		server.RegisterEntities(
			&entities.Message{},
		),
		server.RegisterRepositories(
			repositories.NewMessageRepository(),
		),
		server.RegisterServices(
			services.NewSessionService(),
			services.NewAuthService(),
			services.NewMessageService(),
		),
		server.RegisterMiddlewares(
			middlewares.NewAuthMiddleware(),
		),
		server.RegisterControllers(
			controllers.NewGetMessagesController(),
			controllers.NewCreateMessageController(),
			controllers.NewAdminLoginController(),
			controllers.NewGetAllMessagesController(),
			controllers.NewUpdateStatusController(),
			controllers.NewDeleteMessageController(),
		),
	)
}
