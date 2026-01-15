// Package application 管理应用容器和依赖注入
package application

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/config"
	"com.litelake.litecore/container"
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

	// 创建容器
	configContainer := container.NewConfigContainer()
	entityContainer := container.NewEntityContainer()
	managerContainer := container.NewManagerContainer(configContainer)
	repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
	serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
	controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
	middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

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

	// 第三步：注册所有组件到容器

	// 注册配置
	container.RegisterConfig[common.BaseConfigProvider](configContainer, configProvider)

	// 注册管理器
	container.RegisterManager[databasemgr.DatabaseManager](managerContainer, dbMgr)
	container.RegisterManager[cachemgr.CacheManager](managerContainer, cacheMgr)
	container.RegisterManager[loggermgr.LoggerManager](managerContainer, loggerMgr)

	// 注册实体
	container.RegisterEntity[common.BaseEntity](entityContainer, &entities.Message{})

	// 注册仓储
	container.RegisterRepository[repositories.IMessageRepository](repositoryContainer, repositories.NewMessageRepository())

	// 注册服务
	container.RegisterService[services.ISessionService](serviceContainer, services.NewSessionService())
	container.RegisterService[services.IAuthService](serviceContainer, services.NewAuthService())
	container.RegisterService[services.IMessageService](serviceContainer, services.NewMessageService())

	// 注册中间件
	container.RegisterMiddleware[middlewares.IAuthMiddleware](middlewareContainer, middlewares.NewAuthMiddleware())

	// 注册控制器
	container.RegisterController[controllers.IGetMessagesController](controllerContainer, controllers.NewGetMessagesController())
	container.RegisterController[controllers.ICreateMessageController](controllerContainer, controllers.NewCreateMessageController())
	container.RegisterController[controllers.IAdminLoginController](controllerContainer, controllers.NewAdminLoginController())
	container.RegisterController[controllers.IGetAllMessagesController](controllerContainer, controllers.NewGetAllMessagesController())
	container.RegisterController[controllers.IUpdateStatusController](controllerContainer, controllers.NewUpdateStatusController())
	container.RegisterController[controllers.IDeleteMessageController](controllerContainer, controllers.NewDeleteMessageController())

	// 第四步：创建引擎，传入容器
	return server.NewEngine(
		configContainer,
		entityContainer,
		managerContainer,
		repositoryContainer,
		serviceContainer,
		controllerContainer,
		middlewareContainer,
	), nil
}
