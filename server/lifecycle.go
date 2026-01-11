package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"com.litelake.litecore/common"
)

// startManagers 启动所有管理器
func (e *Engine) startManagers() error {
	managers := e.containers.manager.GetAll()

	// 按注册顺序启动（容器已保证拓扑顺序）
	for _, mgr := range managers {
		fmt.Printf("[DEBUG] Starting manager: %s...\n", mgr.ManagerName())
		if err := mgr.OnStart(); err != nil {
			return fmt.Errorf("failed to start manager %s: %w",
				mgr.ManagerName(), err)
		}
		fmt.Printf("[DEBUG] Manager %s started\n", mgr.ManagerName())
	}
	return nil
}

// stopManagers 停止所有管理器
func (e *Engine) stopManagers() error {
	managers := e.containers.manager.GetAll()

	// 逆序停止
	for i := len(managers) - 1; i >= 0; i-- {
		if err := managers[i].OnStop(); err != nil {
			// 记录错误但继续停止其他 Manager
			fmt.Printf("warning: failed to stop manager %s: %v\n",
				managers[i].ManagerName(), err)
		}
	}
	return nil
}

// Stop 停止引擎（实现 LiteServer 接口）
// - 停止 HTTP 服务器
// - 停止所有 Manager
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	// 1. 停止 HTTP 服务器（超时控制）
	ctx, cancel := context.WithTimeout(context.Background(), e.shutdownTimeout)
	defer cancel()

	if err := e.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	// 2. 停止所有 Manager
	if err := e.stopManagers(); err != nil {
		return err
	}

	e.started = false
	return nil
}

// waitForShutdown 等待关闭信号
func (e *Engine) waitForShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigs

	// 执行优雅关闭
	if err := e.Stop(); err != nil {
		fmt.Printf("shutdown error: %v\n", err)
		os.Exit(1)
	}
}

// Restart 重启引擎
func (e *Engine) Restart() error {
	// 停止
	if err := e.Stop(); err != nil {
		return fmt.Errorf("stop failed: %w", err)
	}

	// 启动
	if err := e.Start(); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}

	return nil
}

// GracefulShutdown 优雅关闭（带超时控制）
func (e *Engine) GracefulShutdown(timeout time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	// 1. 停止 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := e.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	// 2. 停止所有 Manager
	if err := e.stopManagers(); err != nil {
		return err
	}

	e.started = false
	return nil
}

// GetManagers 获取所有管理器
func (e *Engine) GetManagers() []common.BaseManager {
	return e.containers.manager.GetAll()
}

// GetServices 获取所有服务
func (e *Engine) GetServices() []common.BaseService {
	return e.containers.service.GetAll()
}

// GetControllers 获取所有控制器
func (e *Engine) GetControllers() []common.BaseController {
	return e.containers.controller.GetAll()
}

// GetMiddlewares 获取所有中间件
func (e *Engine) GetMiddlewares() []common.BaseMiddleware {
	return e.containers.middleware.GetAll()
}
