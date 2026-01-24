package schedulermgr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lite-lake/litecore-go/common"
)

type schedulerTask struct {
	scheduler   common.IBaseScheduler
	expr        *cronExpression
	cancel      context.CancelFunc
	cancelled   bool
	cancelMutex sync.RWMutex
}

type schedulerManagerImpl struct {
	config   *CronConfig
	tasks    map[string]*schedulerTask
	tasksMap map[common.IBaseScheduler]*schedulerTask
	mu       sync.RWMutex
}

func NewSchedulerManagerCronImpl(config *CronConfig) ISchedulerManager {
	return &schedulerManagerImpl{
		config:   config,
		tasks:    make(map[string]*schedulerTask),
		tasksMap: make(map[common.IBaseScheduler]*schedulerTask),
	}
}

func (s *schedulerManagerImpl) ManagerName() string {
	return "SchedulerManager"
}

func (s *schedulerManagerImpl) Health() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return nil
}

func (s *schedulerManagerImpl) OnStart() error {
	return nil
}

func (s *schedulerManagerImpl) OnStop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		task.cancelMutex.Lock()
		if !task.cancelled {
			task.cancelled = true
			task.cancel()
		}
		task.cancelMutex.Unlock()
	}

	return nil
}

func (s *schedulerManagerImpl) ValidateScheduler(scheduler common.IBaseScheduler) error {
	if scheduler == nil {
		return fmt.Errorf("scheduler cannot be nil")
	}

	rule := scheduler.GetRule()
	if rule == "" {
		return fmt.Errorf("scheduler rule cannot be empty")
	}

	timezone := scheduler.GetTimezone()
	_, err := parseCrontab(rule, timezone)
	if err != nil {
		return fmt.Errorf("invalid crontab expression: %w", err)
	}

	return nil
}

func (s *schedulerManagerImpl) RegisterScheduler(scheduler common.IBaseScheduler) error {
	if scheduler == nil {
		return fmt.Errorf("scheduler cannot be nil")
	}

	if err := s.ValidateScheduler(scheduler); err != nil {
		return err
	}

	name := scheduler.SchedulerName()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[name]; exists {
		return fmt.Errorf("scheduler %s already registered", name)
	}

	rule := scheduler.GetRule()
	timezone := scheduler.GetTimezone()

	expr, err := parseCrontab(rule, timezone)
	if err != nil {
		return fmt.Errorf("failed to parse crontab expression: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	task := &schedulerTask{
		scheduler: scheduler,
		expr:      expr,
		cancel:    cancel,
		cancelled: false,
	}

	s.tasks[name] = task
	s.tasksMap[scheduler] = task

	go s.runSchedulerTask(ctx, task)

	return nil
}

func (s *schedulerManagerImpl) UnregisterScheduler(scheduler common.IBaseScheduler) error {
	if scheduler == nil {
		return fmt.Errorf("scheduler cannot be nil")
	}

	name := scheduler.SchedulerName()

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("scheduler %s not found", name)
	}

	task.cancelMutex.Lock()
	if !task.cancelled {
		task.cancelled = true
		task.cancel()
	}
	task.cancelMutex.Unlock()

	delete(s.tasks, name)
	delete(s.tasksMap, scheduler)

	return nil
}

func (s *schedulerManagerImpl) runSchedulerTask(ctx context.Context, task *schedulerTask) {
	nextTime := task.expr.getNextExecutionTime(time.Now())
	if nextTime.IsZero() {
		return
	}

	ticker := time.NewTimer(time.Until(nextTime))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()

			task.cancelMutex.RLock()
			isCancelled := task.cancelled
			task.cancelMutex.RUnlock()

			if isCancelled {
				return
			}

			tickID := nextTime.Unix()
			s.executeTask(task.scheduler, tickID)

			nextTime = task.expr.getNextExecutionTime(now)
			if nextTime.IsZero() {
				return
			}
			ticker.Reset(time.Until(nextTime))
		}
	}
}

func (s *schedulerManagerImpl) executeTask(scheduler common.IBaseScheduler, tickID int64) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				var err error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("panic recovered: %v", r)
				}
				fmt.Printf("[Scheduler] %s panic: %v\n", scheduler.SchedulerName(), err)
			}
		}()

		if err := scheduler.OnTick(tickID); err != nil {
			fmt.Printf("[Scheduler] %s OnTick error: %v\n", scheduler.SchedulerName(), err)
		}
	}()
}

var _ ISchedulerManager = (*schedulerManagerImpl)(nil)
var _ common.IBaseManager = (*schedulerManagerImpl)(nil)
