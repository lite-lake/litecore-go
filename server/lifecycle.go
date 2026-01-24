package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/mqmgr"
	"github.com/lite-lake/litecore-go/manager/schedulermgr"
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
	e.logPhaseStart(PhaseStartup, "Starting Manager layer")
	managers := e.Manager.GetAll()

	for _, mgr := range managers {
		if err := mgr.(common.IBaseManager).OnStart(); err != nil {
			return fmt.Errorf("failed to start manager %s: %w", mgr.(common.IBaseManager).ManagerName(), err)
		}
		e.logStartup(PhaseStartup, mgr.(common.IBaseManager).ManagerName()+": started")
	}

	e.logPhaseEnd(PhaseStartup, "Manager layer started", logger.F("count", len(managers)))
	return nil
}

// startRepositories 启动所有仓储
func (e *Engine) startRepositories() error {
	e.logPhaseStart(PhaseStartup, "Starting Repository layer")
	repositories := e.Repository.GetAll()

	for _, repo := range repositories {
		if err := repo.OnStart(); err != nil {
			return fmt.Errorf("failed to start repository %s: %w", repo.RepositoryName(), err)
		}
		e.logStartup(PhaseStartup, repo.RepositoryName()+": started")
	}

	e.logPhaseEnd(PhaseStartup, "Repository layer started", logger.F("count", len(repositories)))
	return nil
}

// startServices 启动所有服务
func (e *Engine) startServices() error {
	e.logPhaseStart(PhaseStartup, "Starting Service layer")
	services := e.Service.GetAll()

	for _, svc := range services {
		if err := svc.OnStart(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", svc.ServiceName(), err)
		}
		e.logStartup(PhaseStartup, svc.ServiceName()+": started")
	}

	e.logPhaseEnd(PhaseStartup, "Service layer started", logger.F("count", len(services)))
	return nil
}

// startMiddlewares 启动所有中间件
func (e *Engine) startMiddlewares() error {
	e.logPhaseStart(PhaseStartup, "Starting Middleware layer")
	middlewares := e.Middleware.GetAll()

	for _, mw := range middlewares {
		if err := mw.OnStart(); err != nil {
			return fmt.Errorf("failed to start middleware %s: %w", mw.MiddlewareName(), err)
		}
		e.logStartup(PhaseStartup, mw.MiddlewareName()+": started")
	}

	e.logPhaseEnd(PhaseStartup, "Middleware layer started", logger.F("count", len(middlewares)))
	return nil
}

// startListeners 启动所有监听器
func (e *Engine) startListeners() error {
	e.logPhaseStart(PhaseStartup, "Starting Listener layer")

	if e.Listener == nil {
		e.getLogger().Info("Listener layer not configured, skipping")
		return nil
	}

	listeners := e.Listener.GetAll()
	if len(listeners) == 0 {
		e.getLogger().Info("No registered Listener, skipping")
		return nil
	}

	mqManager, err := container.GetManager[mqmgr.IMQManager](e.Manager)
	if err != nil {
		return fmt.Errorf("MQManager 未初始化，但存在 %d 个 Listener: %w", len(listeners), err)
	}

	startedCount := 0

	for _, listener := range listeners {
		queue := listener.GetQueue()
		opts := listener.GetSubscribeOptions()

		e.getLogger().Info("Starting message listener",
			logger.F("listener", listener.ListenerName()),
			logger.F("queue", queue))

		var subscribeOpts []mqmgr.SubscribeOption
		for _, opt := range opts {
			if so, ok := opt.(mqmgr.SubscribeOption); ok {
				subscribeOpts = append(subscribeOpts, so)
			}
		}

		wrapper := func(ctx context.Context, msg mqmgr.Message) error {
			return listener.Handle(ctx, msg)
		}

		err := mqManager.SubscribeWithCallback(
			e.ctx,
			queue,
			wrapper,
			subscribeOpts...,
		)
		if err != nil {
			return fmt.Errorf("Failed to start listener %s: %w", listener.ListenerName(), err)
		}

		e.logStartup(PhaseStartup, listener.ListenerName()+": started")
		startedCount++
	}

	e.logPhaseEnd(PhaseStartup, "Listener layer started", logger.F("count", startedCount))
	return nil
}

// startSchedulers 启动所有定时器
func (e *Engine) startSchedulers() error {
	e.logPhaseStart(PhaseStartup, "Starting Scheduler layer")

	if e.Scheduler == nil {
		e.getLogger().Info("Scheduler layer not configured, skipping")
		return nil
	}

	schedulers := e.Scheduler.GetAll()
	if len(schedulers) == 0 {
		e.getLogger().Info("No registered Scheduler, skipping")
		return nil
	}

	schedulerMgr, err := container.GetManager[schedulermgr.ISchedulerManager](e.Manager)
	if err != nil {
		return fmt.Errorf("SchedulerManager 未初始化，但存在 %d 个 Scheduler: %w", len(schedulers), err)
	}

	startedCount := 0

	for _, scheduler := range schedulers {
		e.getLogger().Info("Registering scheduler",
			logger.F("scheduler", scheduler.SchedulerName()),
			logger.F("rule", scheduler.GetRule()),
			logger.F("timezone", scheduler.GetTimezone()))

		if err := schedulerMgr.RegisterScheduler(scheduler); err != nil {
			return fmt.Errorf("Failed to register scheduler %s: %w", scheduler.SchedulerName(), err)
		}

		if err := scheduler.OnStart(); err != nil {
			return fmt.Errorf("Failed to start scheduler %s: %w", scheduler.SchedulerName(), err)
		}

		e.logStartup(PhaseStartup, scheduler.SchedulerName()+": started")
		startedCount++
	}

	e.logPhaseEnd(PhaseStartup, "Scheduler layer started", logger.F("count", startedCount))
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

// stopListeners 停止所有监听器
func (e *Engine) stopListeners() []error {
	if e.Listener == nil {
		return nil
	}

	listeners := e.Listener.GetAll()
	var errors []error

	for i := len(listeners) - 1; i >= 0; i-- {
		listener := listeners[i]
		if err := listener.OnStop(); err != nil {
			errors = append(errors, fmt.Errorf("Failed to stop listener %s: %w", listener.ListenerName(), err))
		}
	}

	return errors
}

// stopSchedulers 停止所有定时器
func (e *Engine) stopSchedulers() []error {
	if e.Scheduler == nil {
		return nil
	}

	schedulers := e.Scheduler.GetAll()
	var errors []error

	for i := len(schedulers) - 1; i >= 0; i-- {
		scheduler := schedulers[i]

		if err := scheduler.OnStop(); err != nil {
			errors = append(errors, fmt.Errorf("Failed to stop scheduler %s: %w", scheduler.SchedulerName(), err))
		}

		schedulerMgr, err := container.GetManager[schedulermgr.ISchedulerManager](e.Manager)
		if err != nil {
			errors = append(errors, fmt.Errorf("Failed to get SchedulerManager: %w", err))
			continue
		}

		if err := schedulerMgr.UnregisterScheduler(scheduler); err != nil {
			errors = append(errors, fmt.Errorf("Failed to unregister scheduler %s: %w", scheduler.SchedulerName(), err))
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

	e.logStartup(PhaseShutdown, "Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), e.shutdownTimeout)
	defer cancel()

	if e.httpServer != nil {
		if err := e.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
	}

	e.logPhaseStart(PhaseShutdown, "Stopping all layers")

	middlewareErrors := e.stopMiddlewares()
	e.logStartup(PhaseShutdown, "Middleware layer stopped")

	listenerErrors := e.stopListeners()
	e.logStartup(PhaseShutdown, "Listener layer stopped")

	schedulerErrors := e.stopSchedulers()
	e.logStartup(PhaseShutdown, "Scheduler layer stopped")

	serviceErrors := e.stopServices()
	e.logStartup(PhaseShutdown, "Service layer stopped")

	repositoryErrors := e.stopRepositories()
	e.logStartup(PhaseShutdown, "Repository layer stopped")

	managerErrors := e.stopManagers()
	e.logStartup(PhaseShutdown, "Manager layer stopped")

	allErrors := make([]error, 0, len(middlewareErrors)+len(listenerErrors)+len(schedulerErrors)+len(serviceErrors)+len(repositoryErrors)+len(managerErrors))
	allErrors = append(allErrors, middlewareErrors...)
	allErrors = append(allErrors, listenerErrors...)
	allErrors = append(allErrors, schedulerErrors...)
	allErrors = append(allErrors, serviceErrors...)
	allErrors = append(allErrors, repositoryErrors...)
	allErrors = append(allErrors, managerErrors...)

	totalDuration := time.Since(e.startupStartTime)
	e.logPhaseEnd(PhaseShutdown, "Shutdown completed",
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
