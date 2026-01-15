package server

import (
	"context"
	"fmt"

	"com.litelake.litecore/common"
)

// startManagers 启动所有管理器
func (e *Engine) startManagers() error {
	managers := e.manager.GetAll()
	for _, mgr := range managers {
		if err := mgr.OnStart(); err != nil {
			return fmt.Errorf("failed to start manager %s: %w", mgr.ManagerName(), err)
		}
	}
	return nil
}

// startRepositories 启动所有仓储
func (e *Engine) startRepositories() error {
	repositories := e.repository.GetAll()
	for _, repo := range repositories {
		if err := repo.OnStart(); err != nil {
			return fmt.Errorf("failed to start repository %s: %w", repo.RepositoryName(), err)
		}
	}
	return nil
}

// startServices 启动所有服务
func (e *Engine) startServices() error {
	services := e.service.GetAll()
	for _, svc := range services {
		if err := svc.OnStart(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", svc.ServiceName(), err)
		}
	}
	return nil
}

// stopManagers 停止所有管理器
func (e *Engine) stopManagers() error {
	managers := e.manager.GetAll()
	for i := len(managers) - 1; i >= 0; i-- {
		if err := managers[i].OnStop(); err != nil {
			fmt.Printf("warning: failed to stop manager %s: %v\n", managers[i].ManagerName(), err)
		}
	}
	return nil
}

// stopServices 停止所有服务
func (e *Engine) stopServices() error {
	services := e.service.GetAll()
	for i := len(services) - 1; i >= 0; i-- {
		if err := services[i].OnStop(); err != nil {
			fmt.Printf("warning: failed to stop service %s: %v\n", services[i].ServiceName(), err)
		}
	}
	return nil
}

// stopRepositories 停止所有仓储
func (e *Engine) stopRepositories() error {
	repositories := e.repository.GetAll()
	for i := len(repositories) - 1; i >= 0; i-- {
		if err := repositories[i].OnStop(); err != nil {
			fmt.Printf("warning: failed to stop repository %s: %v\n", repositories[i].RepositoryName(), err)
		}
	}
	return nil
}

// Stop 停止引擎（实现 LiteServer 接口）
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.shutdownTimeout)
	defer cancel()

	if err := e.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	e.stopServices()
	e.stopRepositories()
	e.stopManagers()

	e.started = false
	return nil
}

// Restart 重启引擎
func (e *Engine) Restart() error {
	if err := e.Stop(); err != nil {
		return fmt.Errorf("stop failed: %w", err)
	}
	if err := e.Start(); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}
	return nil
}

// GetManagers 获取所有管理器
func (e *Engine) GetManagers() []common.BaseManager {
	return e.manager.GetAll()
}

// GetServices 获取所有服务
func (e *Engine) GetServices() []common.BaseService {
	return e.service.GetAll()
}

// GetControllers 获取所有控制器
func (e *Engine) GetControllers() []common.BaseController {
	return e.controller.GetAll()
}

// GetMiddlewares 获取所有中间件
func (e *Engine) GetMiddlewares() []common.BaseMiddleware {
	return e.middleware.GetAll()
}
