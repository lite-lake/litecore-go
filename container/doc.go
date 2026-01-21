// Package container 提供分层依赖注入容器，严格约束架构分层并管理组件生命周期。
//
// 核心特性：
//   - 分层架构：定义 Entity/Repository/Service/Controller/Middleware 五层容器，严格单向依赖
//   - 依赖注入：通过 inject 标签自动注入依赖，支持接口类型匹配和可选依赖
//   - 同层依赖：Service 层支持同层依赖，自动拓扑排序确定注入顺序
//   - 类型安全：使用泛型 API 注册，编译时类型检查，接口实现校验
//   - 错误检测：自动检测循环依赖、依赖缺失、接口未实现等错误
//   - 并发安全：使用 RWMutex 保护容器内部状态，支持多线程并发读取
//   - 内置组件：Config 和 Manager 作为服务器内置组件，由引擎自动初始化和注入
//
// 基本用法：
//
//	// 1. 创建容器（按依赖顺序）
//	entityContainer := container.NewEntityContainer()
//	repositoryContainer := container.NewRepositoryContainer(entityContainer)
//	serviceContainer := container.NewServiceContainer(repositoryContainer)
//	controllerContainer := container.NewControllerContainer(serviceContainer)
//	middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
//
//	// 2. 注册实例（使用泛型 API）
//	userService := &UserServiceImpl{}
//	container.RegisterService[UserService](serviceContainer, userService)
//
//	// 3. 执行依赖注入（按层次从下到上）
//	repositoryContainer.InjectAll()
//	serviceContainer.InjectAll()
//	controllerContainer.InjectAll()
//	middlewareContainer.InjectAll()
//
// 依赖声明：
//
//	// 在结构体中使用 inject 标签声明依赖
//	// Config 和 Manager 由引擎自动注入
//	type UserServiceImpl struct {
//		Config    common.IBaseConfigProvider `inject:""`  // 内置组件
//		DBManager DatabaseManager           `inject:""`  // 内置组件
//		UserRepo  UserRepository            `inject:""`
//		OrderSvc  OrderService              `inject:""`   // 同层依赖
//		CacheMgr  CacheManager              `inject:"optional"` // 可选依赖（内置组件）
//	}
package container
