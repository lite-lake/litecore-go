// Package application 管理应用容器和依赖注入
package application

import (
	"com.litelake.litecore/config"
	"com.litelake.litecore/manager/cachemgr"
	"com.litelake.litecore/manager/databasemgr"
	"com.litelake.litecore/manager/loggermgr"
	"com.litelake.litecore/samples/messageboard/internal/controllers"
	"com.litelake.litecore/samples/messageboard/internal/entities"
	"com.litelake.litecore/samples/messageboard/internal/middlewares"
	"com.litelake.litecore/samples/messageboard/internal/repositories"
	"com.litelake.litecore/samples/messageboard/internal/services"
	"com.litelake.litecore/server"
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
		server.RegisterManagers(
			server.Register[databasemgr.DatabaseManager](dbMgr),
			server.Register[cachemgr.CacheManager](cacheMgr),
			server.Register[loggermgr.LoggerManager](loggerMgr),
		),
		server.RegisterEntities(
			&entities.Message{},
		),
		server.RegisterRepositories(
			server.Register[repositories.IMessageRepository](repositories.NewMessageRepository()),
		),
		server.RegisterServices(
			server.Register[services.ISessionService](services.NewSessionService()),
			server.Register[services.IAuthService](services.NewAuthService()),
			server.Register[services.IMessageService](services.NewMessageService()),
		),
		server.RegisterMiddlewares(
			server.Register[middlewares.IAuthMiddleware](middlewares.NewAuthMiddleware()),
		),
		server.RegisterControllers(
			server.Register[controllers.IGetMessagesController](controllers.NewGetMessagesController()),
			server.Register[controllers.ICreateMessageController](controllers.NewCreateMessageController()),
			server.Register[controllers.IAdminLoginController](controllers.NewAdminLoginController()),
			server.Register[controllers.IGetAllMessagesController](controllers.NewGetAllMessagesController()),
			server.Register[controllers.IUpdateStatusController](controllers.NewUpdateStatusController()),
			server.Register[controllers.IDeleteMessageController](controllers.NewDeleteMessageController()),
		),
	)
}
