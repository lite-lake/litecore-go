package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/logger"
)

// logStartup 记录启动日志
func (e *Engine) logStartup(phase StartupPhase, msg string, fields ...logger.Field) {
	if !e.startupLogConfig.Enabled {
		return
	}

	if e.startupLogConfig.Async && e.asyncLogger != nil {
		e.asyncLogger.Log(phase, msg, fields...)
	} else {
		e.getLogger().Info(msg, fields...)
	}
}

// logPhaseStart 记录阶段开始
func (e *Engine) logPhaseStart(phase StartupPhase, msg string, fields ...logger.Field) {
	e.phaseStartTimes[phase] = time.Now()
	e.logStartup(phase, msg, fields...)
}

// logPhaseEnd 记录阶段结束（带耗时）
func (e *Engine) logPhaseEnd(phase StartupPhase, msg string, extraFields ...logger.Field) {
	duration := time.Since(e.phaseStartTimes[phase])
	e.phaseDurations[phase] = duration

	fields := append(extraFields,
		logger.F("duration", duration.String()),
		logger.F("phase", phase.String()))
	e.logStartup(phase, msg, fields...)
}

// startManagers 启动所有管理器
func (e *Engine) startManagers() error {
	e.logPhaseStart(PhaseStartup, "开始启动 Manager 层")
	managers := e.Manager.GetAll()

	for _, mgr := range managers {
		if err := mgr.(common.IBaseManager).OnStart(); err != nil {
			return fmt.Errorf("failed to start manager %s: %w", mgr.(common.IBaseManager).ManagerName(), err)
		}
		e.logStartup(PhaseStartup, mgr.(common.IBaseManager).ManagerName()+": 启动完成")
	}

	e.logPhaseEnd(PhaseStartup, "Manager 层启动完成", logger.F("count", len(managers)))
	return nil
}

// startRepositories 启动所有仓储
func (e *Engine) startRepositories() error {
	e.logPhaseStart(PhaseStartup, "开始启动 Repository 层")
	repositories := e.Repository.GetAll()

	for _, repo := range repositories {
		if err := repo.OnStart(); err != nil {
			return fmt.Errorf("failed to start repository %s: %w", repo.RepositoryName(), err)
		}
		e.logStartup(PhaseStartup, repo.RepositoryName()+": 启动完成")
	}

	e.logPhaseEnd(PhaseStartup, "Repository 层启动完成", logger.F("count", len(repositories)))
	return nil
}

// startServices 启动所有服务
func (e *Engine) startServices() error {
	e.logPhaseStart(PhaseStartup, "开始启动 Service 层")
	services := e.Service.GetAll()

	for _, svc := range services {
		if err := svc.OnStart(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", svc.ServiceName(), err)
		}
		e.logStartup(PhaseStartup, svc.ServiceName()+": 启动完成")
	}

	e.logPhaseEnd(PhaseStartup, "Service 层启动完成", logger.F("count", len(services)))
	return nil
}

// startMiddlewares 启动所有中间件
func (e *Engine) startMiddlewares() error {
	e.logPhaseStart(PhaseStartup, "开始启动 Middleware 层")
	middlewares := e.Middleware.GetAll()

	for _, mw := range middlewares {
		if err := mw.OnStart(); err != nil {
			return fmt.Errorf("failed to start middleware %s: %w", mw.MiddlewareName(), err)
		}
		e.logStartup(PhaseStartup, mw.MiddlewareName()+": 启动完成")
	}

	e.logPhaseEnd(PhaseStartup, "Middleware 层启动完成", logger.F("count", len(middlewares)))
	return nil
}

// stopManagers 停止所有管理器
func (e *Engine) stopManagers() []error {
	managers := e.Manager.GetAll()
	var errors []error
	for i := len(managers) - 1; i >= 0; i-- {
		if err := managers[i].(common.IBaseManager).OnStop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop manager %s: %w", managers[i].(common.IBaseManager).ManagerName(), err))
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

// stopMiddlewares 停止所有中间件
func (e *Engine) stopMiddlewares() []error {
	middlewares := e.Middleware.GetAll()
	var errors []error
	for i := len(middlewares) - 1; i >= 0; i-- {
		if err := middlewares[i].OnStop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop middleware %s: %w", middlewares[i].MiddlewareName(), err))
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

	e.logStartup(PhaseShutdown, "HTTP 服务器关闭...")

	ctx, cancel := context.WithTimeout(context.Background(), e.shutdownTimeout)
	defer cancel()

	if e.httpServer != nil {
		if err := e.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
	}

	e.logPhaseStart(PhaseShutdown, "开始停止各层组件")

	middlewareErrors := e.stopMiddlewares()
	e.logStartup(PhaseShutdown, "Middleware 层停止完成")

	serviceErrors := e.stopServices()
	e.logStartup(PhaseShutdown, "Service 层停止完成")

	repositoryErrors := e.stopRepositories()
	e.logStartup(PhaseShutdown, "Repository 层停止完成")

	managerErrors := e.stopManagers()
	e.logStartup(PhaseShutdown, "Manager 层停止完成")

	allErrors := make([]error, 0, len(middlewareErrors)+len(serviceErrors)+len(repositoryErrors)+len(managerErrors))
	allErrors = append(allErrors, middlewareErrors...)
	allErrors = append(allErrors, serviceErrors...)
	allErrors = append(allErrors, repositoryErrors...)
	allErrors = append(allErrors, managerErrors...)

	totalDuration := time.Since(e.startupStartTime)
	e.logPhaseEnd(PhaseShutdown, "关闭完成",
		logger.F("error_count", len(allErrors)),
		logger.F("total_duration", totalDuration.String()))

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
