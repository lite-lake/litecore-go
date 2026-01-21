package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/lite-lake/litecore-go/common"
)

// startManagers 启动所有管理器
func (e *Engine) startManagers() error {
	managers := e.Manager.GetAll()
	for _, mgr := range managers {
		if err := mgr.OnStart(); err != nil {
			return fmt.Errorf("failed to start manager %s: %w", mgr.ManagerName(), err)
		}
	}
	return nil
}

// startRepositories 启动所有仓储
func (e *Engine) startRepositories() error {
	repositories := e.Repository.GetAll()
	for _, repo := range repositories {
		if err := repo.OnStart(); err != nil {
			return fmt.Errorf("failed to start repository %s: %w", repo.RepositoryName(), err)
		}
	}
	return nil
}

// startServices 启动所有服务
func (e *Engine) startServices() error {
	services := e.Service.GetAll()
	for _, svc := range services {
		if err := svc.OnStart(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", svc.ServiceName(), err)
		}
	}
	return nil
}

// stopManagers 停止所有管理器
func (e *Engine) stopManagers() []error {
	managers := e.Manager.GetAll()
	var errors []error
	for i := len(managers) - 1; i >= 0; i-- {
		if err := managers[i].OnStop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop manager %s: %w", managers[i].ManagerName(), err))
		}
	}
	return errors
}

// stopServices 停止所有服务
func (e *Engine) stopServices() []error {
	services := e.Service.GetAll()
	var errors []error
	for i := len(services) - 1; i >= 0; i-- {
		if err := services[i].OnStop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop service %s: %w", services[i].ServiceName(), err))
		}
	}
	return errors
}

// stopRepositories 停止所有仓储
func (e *Engine) stopRepositories() []error {
	repositories := e.Repository.GetAll()
	var errors []error
	for i := len(repositories) - 1; i >= 0; i-- {
		if err := repositories[i].OnStop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop repository %s: %w", repositories[i].RepositoryName(), err))
		}
	}
	return errors
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

	if e.httpServer != nil {
		if err := e.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
	}

	serviceErrors := e.stopServices()
	repositoryErrors := e.stopRepositories()
	managerErrors := e.stopManagers()

	allErrors := make([]error, 0, len(serviceErrors)+len(repositoryErrors)+len(managerErrors))
	allErrors = append(allErrors, serviceErrors...)
	allErrors = append(allErrors, repositoryErrors...)
	allErrors = append(allErrors, managerErrors...)

	if len(allErrors) > 0 {
		errorMessages := make([]string, len(allErrors))
		for i, err := range allErrors {
			errorMessages[i] = err.Error()
		}
		return fmt.Errorf("shutdown completed with %d error(s): %s", len(allErrors), strings.Join(errorMessages, "; "))
	}

	e.started = false
	return nil
}

// getServices 获取所有服务
func (e *Engine) getServices() []common.IBaseService {
	return e.Service.GetAll()
}
