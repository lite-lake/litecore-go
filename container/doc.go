// Package container 提供依赖注入容器功能，支持 7 层分层架构的自动依赖管理。
//
// 核心特性：
//   - 7 层容器：支持 Entity、Manager、Repository、Service、Controller、Middleware、Scheduler、Listener 八层容器
//   - 类型安全：使用泛型确保类型安全，编译时检查
//   - 自动注入：通过结构体标签 `inject:""` 自动注入依赖
//   - 拓扑排序：Service 层使用 Kahn 算法检测并解决循环依赖
//   - 线程安全：所有容器操作都使用读写锁保护
//   - 分层依赖：严格的依赖层级，Controller/Middleware/Scheduler/Listener 禁止直接注入 Repository
//
// 基本用法：
//
//	// 创建容器链（按依赖顺序）
//	entityContainer := container.NewEntityContainer()
//	repositoryContainer := container.NewRepositoryContainer(entityContainer)
//	serviceContainer := container.NewServiceContainer(repositoryContainer)
//	controllerContainer := container.NewControllerContainer(serviceContainer)
//	middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
//	schedulerContainer := container.NewSchedulerContainer(serviceContainer)
//	listenerContainer := container.NewListenerContainer(serviceContainer)
//
//	// 注册实例
//	container.RegisterEntity(entityContainer, &Message{})
//	container.RegisterRepository[IMessageRepository](repositoryContainer, repo)
//	container.RegisterService[IMessageService](serviceContainer, svc)
//	container.RegisterController[IMessageController](controllerContainer, ctrl)
//	container.RegisterMiddleware[IAuthMiddleware](middlewareContainer, mw)
//
// 服务层拓扑排序：
//
// 服务层的依赖注入支持拓扑排序，确保依赖按正确顺序注入。
// 例如：ServiceA 依赖 ServiceB，ServiceB 依赖 ServiceC，
// 注入顺序为：ServiceC → ServiceB → ServiceA。循环依赖会触发 CircularDependencyError。
//
// 错误处理：
//
// 包中定义了多种错误类型：
//   - DependencyNotFoundError：依赖未找到
//   - CircularDependencyError：Service 层循环依赖
//   - AmbiguousMatchError：多重匹配（Entity 层同类型多个实例）
//   - DuplicateRegistrationError：重复注册
//   - InstanceNotFoundError：实例未找到
//   - InterfaceAlreadyRegisteredError：接口已注册
//   - ImplementationDoesNotImplementInterfaceError：实现未实现接口
//   - InterfaceNotRegisteredError：接口未注册
//   - ManagerContainerNotSetError：ManagerContainer 未设置
//   - UninjectedFieldError：标记 inject:"" 的字段注入后仍为 nil
package container
