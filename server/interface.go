package server

// LiteServer 服务器接口定义
type LiteServer interface {
	// Initialize 初始化服务器
	Initialize() error
	// Start 启动服务器
	Start() error
	// Stop 停止服务器
	Stop() error
}
