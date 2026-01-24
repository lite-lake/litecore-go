package listeners

import (
	"context"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IMessageAuditListener interface {
	common.IBaseListener
}

type messageAuditListenerImpl struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewMessageAuditListener() IMessageAuditListener {
	return &messageAuditListenerImpl{}
}

func (l *messageAuditListenerImpl) ListenerName() string {
	return "messageAuditListenerImpl"
}

func (l *messageAuditListenerImpl) GetQueue() string {
	return "message.audit"
}

func (l *messageAuditListenerImpl) GetSubscribeOptions() []common.ISubscribeOption {
	return []common.ISubscribeOption{}
}

func (l *messageAuditListenerImpl) OnStart() error {
	if l.LoggerMgr != nil {
		l.LoggerMgr.Ins().Info("消息审核监听器已启动")
	}
	return nil
}

func (l *messageAuditListenerImpl) OnStop() error {
	if l.LoggerMgr != nil {
		l.LoggerMgr.Ins().Info("消息审核监听器已停止")
	}
	return nil
}

func (l *messageAuditListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error {
	if l.LoggerMgr != nil {
		l.LoggerMgr.Ins().Info("收到消息审核事件",
			"message_id", msg.ID(),
			"body", string(msg.Body()),
			"headers", msg.Headers())
	}
	return nil
}

var _ IMessageAuditListener = (*messageAuditListenerImpl)(nil)
var _ common.IBaseListener = (*messageAuditListenerImpl)(nil)
