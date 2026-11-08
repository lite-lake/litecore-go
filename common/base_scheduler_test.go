package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockScheduler struct {
	name     string
	rule     string
	timezone string
	onTick   func(tickID int64) error
	startErr error
	stopErr  error
}

func (m *mockScheduler) SchedulerName() string {
	return m.name
}

func (m *mockScheduler) GetRule() string {
	return m.rule
}

func (m *mockScheduler) GetTimezone() string {
	return m.timezone
}

func (m *mockScheduler) OnTick(tickID int64) error {
	if m.onTick != nil {
		return m.onTick(tickID)
	}
	return nil
}

func (m *mockScheduler) OnStart() error {
	return m.startErr
}

func (m *mockScheduler) OnStop() error {
	return m.stopErr
}

type cleanupScheduler struct{}

func (c *cleanupScheduler) SchedulerName() string {
	return "cleanupScheduler"
}

func (c *cleanupScheduler) GetRule() string {
	return "0 0 2 * * *"
}

func (c *cleanupScheduler) GetTimezone() string {
	return "Asia/Shanghai"
}

func (c *cleanupScheduler) OnTick(tickID int64) error {
	return nil
}

func (c *cleanupScheduler) OnStart() error {
	return nil
}

func (c *cleanupScheduler) OnStop() error {
	return nil
}

type failingScheduler struct{}

func (f *failingScheduler) SchedulerName() string {
	return "failingScheduler"
}

func (f *failingScheduler) GetRule() string {
	return "* * * * * *"
}

func (f *failingScheduler) GetTimezone() string {
	return ""
}

func (f *failingScheduler) OnTick(tickID int64) error {
	return errors.New("定时任务执行失败")
}

func (f *failingScheduler) OnStart() error {
	return errors.New("定时器启动失败")
}

func (f *failingScheduler) OnStop() error {
	return errors.New("定时器停止失败")
}

func TestIBaseScheduler_基础接口实现(t *testing.T) {
	scheduler := &mockScheduler{
		name:     "TestScheduler",
		rule:     "0 */5 * * * *",
		timezone: "",
	}

	assert.Equal(t, "TestScheduler", scheduler.SchedulerName())
	assert.Equal(t, "0 */5 * * * *", scheduler.GetRule())
	assert.Equal(t, "", scheduler.GetTimezone())
	assert.NoError(t, scheduler.OnStart())
	assert.NoError(t, scheduler.OnStop())
}

func TestIBaseScheduler_OnTick方法(t *testing.T) {
	tests := []struct {
		name      string
		scheduler IBaseScheduler
		tickID    int64
		wantErr   bool
	}{
		{
			name:      "正常执行定时任务",
			scheduler: &mockScheduler{name: "Normal", rule: "0 */5 * * * *"},
			tickID:    1234567890,
			wantErr:   false,
		},
		{
			name:      "清理定时器",
			scheduler: &cleanupScheduler{},
			tickID:    9876543210,
			wantErr:   false,
		},
		{
			name:      "执行失败",
			scheduler: &failingScheduler{},
			tickID:    1111111111,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scheduler.OnTick(tt.tickID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseScheduler_生命周期方法(t *testing.T) {
	tests := []struct {
		name      string
		scheduler IBaseScheduler
		wantErr   bool
	}{
		{
			name:      "正常启动和停止",
			scheduler: &mockScheduler{name: "LifecycleTest", rule: "0 */5 * * * *"},
			wantErr:   false,
		},
		{
			name:      "启动失败的定时器",
			scheduler: &failingScheduler{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scheduler.OnStart()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = tt.scheduler.OnStop()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseScheduler_空实现(t *testing.T) {
	tests := []struct {
		name      string
		scheduler IBaseScheduler
	}{
		{
			name:      "空定时器实例",
			scheduler: &mockScheduler{},
		},
		{
			name:      "带有空名称的定时器",
			scheduler: &mockScheduler{name: "", rule: "0 */5 * * * *"},
		},
		{
			name:      "带有空规则的定时器",
			scheduler: &mockScheduler{name: "Test", rule: ""},
		},
		{
			name:      "带有空时区的定时器",
			scheduler: &mockScheduler{name: "Test", rule: "0 */5 * * * *", timezone: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.scheduler.SchedulerName())
			assert.NotNil(t, tt.scheduler.GetRule())
			assert.NotNil(t, tt.scheduler.GetTimezone())
		})
	}
}

func TestIBaseScheduler_接口组合(t *testing.T) {
	scheduler := &mockScheduler{
		name:     "CombinedScheduler",
		rule:     "0 0 2 * * *",
		timezone: "UTC",
	}

	var iface IBaseScheduler = scheduler
	assert.Equal(t, "CombinedScheduler", iface.SchedulerName())
	assert.Equal(t, "0 0 2 * * *", iface.GetRule())
	assert.Equal(t, "UTC", iface.GetTimezone())
}

func TestIBaseScheduler_Crontab规则(t *testing.T) {
	tests := []struct {
		name      string
		scheduler IBaseScheduler
		expected  string
	}{
		{
			name:      "每5分钟执行",
			scheduler: &mockScheduler{name: "FiveMin", rule: "0 */5 * * * *"},
			expected:  "0 */5 * * * *",
		},
		{
			name:      "每天凌晨2点",
			scheduler: &mockScheduler{name: "Daily", rule: "0 0 2 * * *"},
			expected:  "0 0 2 * * *",
		},
		{
			name:      "每周一凌晨",
			scheduler: &mockScheduler{name: "Weekly", rule: "0 0 * * * 1"},
			expected:  "0 0 * * * 1",
		},
		{
			name:      "每月1号",
			scheduler: &mockScheduler{name: "Monthly", rule: "0 0 0 1 * *"},
			expected:  "0 0 0 1 * *",
		},
		{
			name:      "每秒执行",
			scheduler: &mockScheduler{name: "EverySecond", rule: "* * * * * *"},
			expected:  "* * * * * *",
		},
		{
			name:      "空规则",
			scheduler: &mockScheduler{name: "EmptyRule", rule: ""},
			expected:  "",
		},
		{
			name:      "清理定时器规则",
			scheduler: &cleanupScheduler{},
			expected:  "0 0 2 * * *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.scheduler.GetRule())
		})
	}
}

func TestIBaseScheduler_时区配置(t *testing.T) {
	tests := []struct {
		name      string
		scheduler IBaseScheduler
		expected  string
	}{
		{
			name:      "上海时区",
			scheduler: &mockScheduler{name: "Shanghai", rule: "0 */5 * * * *", timezone: "Asia/Shanghai"},
			expected:  "Asia/Shanghai",
		},
		{
			name:      "UTC时区",
			scheduler: &mockScheduler{name: "UTC", rule: "0 */5 * * * *", timezone: "UTC"},
			expected:  "UTC",
		},
		{
			name:      "纽约时区",
			scheduler: &mockScheduler{name: "NewYork", rule: "0 */5 * * * *", timezone: "America/New_York"},
			expected:  "America/New_York",
		},
		{
			name:      "空时区（本地时间）",
			scheduler: &mockScheduler{name: "Local", rule: "0 */5 * * * *", timezone: ""},
			expected:  "",
		},
		{
			name:      "清理定时器时区",
			scheduler: &cleanupScheduler{},
			expected:  "Asia/Shanghai",
		},
		{
			name:      "失败定时器时区",
			scheduler: &failingScheduler{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.scheduler.GetTimezone())
		})
	}
}

func TestIBaseScheduler_定时任务ID(t *testing.T) {
	tests := []struct {
		name      string
		scheduler IBaseScheduler
		tickIDs   []int64
	}{
		{
			name:      "正常TickID",
			scheduler: &mockScheduler{name: "TickTest", rule: "0 */5 * * * *"},
			tickIDs:   []int64{0, 1234567890, 9876543210, 9999999999},
		},
		{
			name:      "负数TickID",
			scheduler: &mockScheduler{name: "NegativeTick", rule: "0 */5 * * * *"},
			tickIDs:   []int64{-1, -100, -999999},
		},
		{
			name:      "零TickID",
			scheduler: &mockScheduler{name: "ZeroTick", rule: "0 */5 * * * *"},
			tickIDs:   []int64{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, tickID := range tt.tickIDs {
				err := tt.scheduler.OnTick(tickID)
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseScheduler_自定义OnTick(t *testing.T) {
	var receivedTickID int64

	scheduler := &mockScheduler{
		name:     "CustomTickScheduler",
		rule:     "0 */5 * * * *",
		timezone: "",
		onTick: func(tickID int64) error {
			receivedTickID = tickID
			return nil
		},
	}

	testTickID := int64(1234567890)
	err := scheduler.OnTick(testTickID)

	assert.NoError(t, err)
	assert.Equal(t, testTickID, receivedTickID)
}
