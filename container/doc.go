// Package container 提供依赖注入容器功能，支持分层架构的依赖管理。
//
// 核心特性：
//   - 分层容器：支持 Entity、Manager、Repository、Service、Controller、Middleware 六层容器
//   - 类型安全：使用泛型确保类型安全，编译时检查
//   - 自动注入：通过结构体标签 `inject:""` 自动注入依赖
//   - 循环检测：使用拓扑排序检测循环依赖
//   - 线程安全：所有容器操作都使用读写锁保护
//
// 基本用法：
//
//	// 创建容器链
//	entityContainer := container.NewEntityContainer()
//	managerContainer := container.NewManagerContainer()
//	repositoryContainer := container.NewRepositoryContainer(entityContainer)
//	serviceContainer := container.NewServiceContainer(repositoryContainer)
//	controllerContainer := container.NewControllerContainer(serviceContainer)
//	middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
//
//	// 设置容器链
//	repositoryContainer.SetManagerContainer(managerContainer)
//	serviceContainer.SetManagerContainer(managerContainer)
//	controllerContainer.SetManagerContainer(managerContainer)
//	middlewareContainer.SetManagerContainer(managerContainer)
//
//	// 注册实例
//	container.RegisterEntity(entityContainer, &UserEntity{})
//	container.RegisterManager(managerContainer, &ConfigManager{})
//	container.RegisterRepository(repositoryContainer, &UserRepository{})
//	container.RegisterService(serviceContainer, &UserService{})
//	container.RegisterController(controllerContainer, &UserController{})
//
//	// 执行依赖注入
//	serviceContainer.InjectAll()
//	controllerContainer.InjectAll()
//
// 服务层依赖：
//
// 服务层的依赖注入支持拓扑排序，确保依赖按正确顺序注入。
// 例如：ServiceA 依赖 ServiceB，ServiceB 依赖 ServiceC，
// 注入顺序为：ServiceC → ServiceB → ServiceA。
//
// 错误处理：
//
// 包中定义了多种错误类型：
//   - DependencyNotFoundError：依赖未找到
//   - CircularDependencyError：循环依赖
//   - AmbiguousMatchError：多重匹配
//   - DuplicateRegistrationError：重复注册
//   - InstanceNotFoundError：实例未找到
//   - InterfaceAlreadyRegisteredError：接口已注册
//   - ImplementationDoesNotImplementInterfaceError：实现未实现接口
package container
