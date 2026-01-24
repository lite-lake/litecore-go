package server

// StartupPhase 启动阶段枚举
type StartupPhase int

const (
	PhaseConfig     StartupPhase = iota // 配置加载
	PhaseManagers                       // 管理器初始化
	PhaseValidation                     // 配置验证
	PhaseInjection                      // 依赖注入
	PhaseRouter                         // 路由注册
	PhaseStartup                        // 组件启动
	PhaseRunning                        // 运行中
	PhaseShutdown                       // 关闭中
)

// String 返回阶段的中文描述
func (p StartupPhase) String() string {
	switch p {
	case PhaseConfig:
		return "配置加载"
	case PhaseManagers:
		return "管理器初始化"
	case PhaseValidation:
		return "配置验证"
	case PhaseInjection:
		return "依赖注入"
	case PhaseRouter:
		return "路由注册"
	case PhaseStartup:
		return "组件启动"
	case PhaseRunning:
		return "运行中"
	case PhaseShutdown:
		return "关闭中"
	default:
		return "未知阶段"
	}
}
