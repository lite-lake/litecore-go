// Package listeners 定义消息监听器
package listeners

import (
	"context"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

// IMessageAuditListener 留言审核监听器接口
type IMessageAuditListener interface {
	common.IBaseListener
}

type messageAuditListenerImpl struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""` // 日志管理器
}

// NewMessageAuditListener 创建留言审核监听器实例
func NewMessageAuditListener() IMessageAuditListener {
	return &messageAuditListenerImpl{}
}

// ListenerName 返回监听器名称
func (l *messageAuditListenerImpl) ListenerName() string {
	return "MessageAuditListener"
}

// GetQueue 返回监听队列名称
func (l *messageAuditListenerImpl) GetQueue() string {
	return "message.audit"
}

// GetSubscribeOptions 返回订阅配置选项
func (l *messageAuditListenerImpl) GetSubscribeOptions() []common.ISubscribeOption {
	return []common.ISubscribeOption{}
}

// OnStart 监听器启动回调
func (l *messageAuditListenerImpl) OnStart() error {
	l.LoggerMgr.Ins().Info("Message audit listener started")
	return nil
}

// OnStop 监听器停止回调
func (l *messageAuditListenerImpl) OnStop() error {
	l.LoggerMgr.Ins().Info("Message audit listener stopped")
	return nil
}

// Handle 处理收到的消息
func (l *messageAuditListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error {
	l.LoggerMgr.Ins().Info("Received message audit event",
		"message_id", msg.ID(),
		"body", string(msg.Body()),
		"headers", msg.Headers())
	return nil
}

var _ IMessageAuditListener = (*messageAuditListenerImpl)(nil)
var _ common.IBaseListener = (*messageAuditListenerImpl)(nil)
