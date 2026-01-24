// Package listeners 定义消息监听器
package listeners

import (
	"context"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

// IMessageCreatedListener 留言创建监听器接口
type IMessageCreatedListener interface {
	common.IBaseListener
}

type messageCreatedListenerImpl struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewMessageCreatedListener 创建留言创建监听器实例
func NewMessageCreatedListener() IMessageCreatedListener {
	return &messageCreatedListenerImpl{}
}

// ListenerName 返回监听器名称
func (l *messageCreatedListenerImpl) ListenerName() string {
	return "MessageCreatedListener"
}

// GetQueue 返回监听队列名称
func (l *messageCreatedListenerImpl) GetQueue() string {
	return "message.created"
}

// GetSubscribeOptions 返回订阅配置选项
func (l *messageCreatedListenerImpl) GetSubscribeOptions() []common.ISubscribeOption {
	return []common.ISubscribeOption{}
}

// OnStart 监听器启动回调
func (l *messageCreatedListenerImpl) OnStart() error {
	l.LoggerMgr.Ins().Info("Message created listener started")
	return nil
}

// OnStop 监听器停止回调
func (l *messageCreatedListenerImpl) OnStop() error {
	l.LoggerMgr.Ins().Info("Message created listener stopped")
	return nil
}

// Handle 处理收到的消息
func (l *messageCreatedListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error {
	l.LoggerMgr.Ins().Info("Received message created event",
		"message_id", msg.ID(),
		"body", string(msg.Body()),
		"headers", msg.Headers())
	return nil
}

var _ IMessageCreatedListener = (*messageCreatedListenerImpl)(nil)
var _ common.IBaseListener = (*messageCreatedListenerImpl)(nil)
