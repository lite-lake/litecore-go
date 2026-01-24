package common

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMessageListener struct {
	id      string
	body    []byte
	headers map[string]any
}

func (m *mockMessageListener) ID() string {
	return m.id
}

func (m *mockMessageListener) Body() []byte {
	return m.body
}

func (m *mockMessageListener) Headers() map[string]any {
	return m.headers
}

type mockListener struct {
	name     string
	queue    string
	options  []ISubscribeOption
	handle   func(ctx context.Context, msg IMessageListener) error
	startErr error
	stopErr  error
}

func (m *mockListener) ListenerName() string {
	return m.name
}

func (m *mockListener) GetQueue() string {
	return m.queue
}

func (m *mockListener) GetSubscribeOptions() []ISubscribeOption {
	return m.options
}

func (m *mockListener) Handle(ctx context.Context, msg IMessageListener) error {
	if m.handle != nil {
		return m.handle(ctx, msg)
	}
	return nil
}

func (m *mockListener) OnStart() error {
	return m.startErr
}

func (m *mockListener) OnStop() error {
	return m.stopErr
}

type userListener struct{}

func (u *userListener) ListenerName() string {
	return "userListenerImpl"
}

func (u *userListener) GetQueue() string {
	return "user.created"
}

func (u *userListener) GetSubscribeOptions() []ISubscribeOption {
	return []ISubscribeOption{
		"option1",
		"option2",
	}
}

func (u *userListener) Handle(ctx context.Context, msg IMessageListener) error {
	return nil
}

func (u *userListener) OnStart() error {
	return nil
}

func (u *userListener) OnStop() error {
	return nil
}

type failingListener struct{}

func (f *failingListener) ListenerName() string {
	return "failingListenerImpl"
}

func (f *failingListener) GetQueue() string {
	return "failing.queue"
}

func (f *failingListener) GetSubscribeOptions() []ISubscribeOption {
	return nil
}

func (f *failingListener) Handle(ctx context.Context, msg IMessageListener) error {
	return errors.New("消息处理失败")
}

func (f *failingListener) OnStart() error {
	return errors.New("监听器启动失败")
}

func (f *failingListener) OnStop() error {
	return errors.New("监听器停止失败")
}

func TestIBaseListener_基础接口实现(t *testing.T) {
	listener := &mockListener{
		name:    "TestListener",
		queue:   "test.queue",
		options: []ISubscribeOption{"option1"},
	}

	assert.Equal(t, "TestListener", listener.ListenerName())
	assert.Equal(t, "test.queue", listener.GetQueue())
	assert.Equal(t, []ISubscribeOption{"option1"}, listener.GetSubscribeOptions())
	assert.NoError(t, listener.OnStart())
	assert.NoError(t, listener.OnStop())
}

func TestIBaseListener_Handle方法(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		listener IBaseListener
		msg      IMessageListener
		wantErr  bool
	}{
		{
			name:     "正常处理消息",
			listener: &mockListener{name: "Normal", queue: "normal.queue"},
			msg: &mockMessageListener{
				id:   "msg1",
				body: []byte("test message"),
			},
			wantErr: false,
		},
		{
			name:     "用户监听器处理消息",
			listener: &userListener{},
			msg: &mockMessageListener{
				id:   "msg2",
				body: []byte("user created"),
			},
			wantErr: false,
		},
		{
			name:     "处理失败",
			listener: &failingListener{},
			msg: &mockMessageListener{
				id:   "msg3",
				body: []byte("error message"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.listener.Handle(ctx, tt.msg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseListener_生命周期方法(t *testing.T) {
	tests := []struct {
		name     string
		listener IBaseListener
		wantErr  bool
	}{
		{
			name:     "正常启动和停止",
			listener: &mockListener{name: "LifecycleTest", queue: "test.queue"},
			wantErr:  false,
		},
		{
			name:     "启动失败的监听器",
			listener: &failingListener{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.listener.OnStart()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = tt.listener.OnStop()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseListener_空实现(t *testing.T) {
	tests := []struct {
		name     string
		listener IBaseListener
	}{
		{
			name:     "空监听器实例",
			listener: &mockListener{},
		},
		{
			name:     "带有空名称的监听器",
			listener: &mockListener{name: "", queue: "test.queue"},
		},
		{
			name:     "带有空队列的监听器",
			listener: &mockListener{name: "Test", queue: ""},
		},
		{
			name:     "带有空选项的监听器",
			listener: &mockListener{name: "Test", queue: "test.queue", options: []ISubscribeOption{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.listener.ListenerName())
			assert.NotNil(t, tt.listener.GetQueue())
		})
	}
}

func TestIBaseListener_接口组合(t *testing.T) {
	listener := &mockListener{
		name:    "CombinedListener",
		queue:   "combined.queue",
		options: []ISubscribeOption{"opt1", "opt2"},
	}

	var iface IBaseListener = listener
	assert.Equal(t, "CombinedListener", iface.ListenerName())
	assert.Equal(t, "combined.queue", iface.GetQueue())
	assert.Equal(t, []ISubscribeOption{"opt1", "opt2"}, iface.GetSubscribeOptions())
}

func TestIBaseListener_队列名称(t *testing.T) {
	tests := []struct {
		name     string
		listener IBaseListener
		expected string
	}{
		{
			name:     "用户队列",
			listener: &mockListener{name: "UserListener", queue: "user.created"},
			expected: "user.created",
		},
		{
			name:     "消息队列",
			listener: &mockListener{name: "MessageListener", queue: "message.send"},
			expected: "message.send",
		},
		{
			name:     "事件队列",
			listener: &mockListener{name: "EventListener", queue: "event.triggered"},
			expected: "event.triggered",
		},
		{
			name:     "空字符串队列",
			listener: &mockListener{name: "EmptyListener", queue: ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.listener.GetQueue())
		})
	}
}

func TestIMessageListener_消息接口(t *testing.T) {
	tests := []struct {
		name    string
		msg     IMessageListener
		wantID  string
		wantLen int
	}{
		{
			name: "正常消息",
			msg: &mockMessageListener{
				id:      "msg123",
				body:    []byte("test body"),
				headers: map[string]any{"key": "value"},
			},
			wantID:  "msg123",
			wantLen: 9,
		},
		{
			name: "空消息体",
			msg: &mockMessageListener{
				id:      "msg456",
				body:    []byte{},
				headers: map[string]any{},
			},
			wantID:  "msg456",
			wantLen: 0,
		},
		{
			name: "空ID",
			msg: &mockMessageListener{
				id:      "",
				body:    []byte("body"),
				headers: nil,
			},
			wantID:  "",
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantID, tt.msg.ID())
			assert.Equal(t, tt.wantLen, len(tt.msg.Body()))
			headers := tt.msg.Headers()
			if tt.name != "空ID" {
				assert.NotNil(t, headers)
			}
		})
	}
}

func TestIBaseListener_订阅选项(t *testing.T) {
	tests := []struct {
		name        string
		listener    IBaseListener
		expectedLen int
	}{
		{
			name:        "多个选项",
			listener:    &mockListener{name: "Test", queue: "test.queue", options: []ISubscribeOption{"opt1", "opt2", "opt3"}},
			expectedLen: 3,
		},
		{
			name:        "单个选项",
			listener:    &mockListener{name: "Test", queue: "test.queue", options: []ISubscribeOption{"opt1"}},
			expectedLen: 1,
		},
		{
			name:        "无选项",
			listener:    &mockListener{name: "Test", queue: "test.queue", options: []ISubscribeOption{}},
			expectedLen: 0,
		},
		{
			name:        "nil选项",
			listener:    &mockListener{name: "Test", queue: "test.queue", options: nil},
			expectedLen: 0,
		},
		{
			name:        "用户监听器选项",
			listener:    &userListener{},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := tt.listener.GetSubscribeOptions()
			if options == nil {
				assert.Equal(t, tt.expectedLen, 0)
			} else {
				assert.Equal(t, tt.expectedLen, len(options))
			}
		})
	}
}
