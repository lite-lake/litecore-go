package listeners

import (
	"context"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IMessageCreatedListener interface {
	common.IBaseListener
}

type messageCreatedListenerImpl struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewMessageCreatedListener() IMessageCreatedListener {
	return &messageCreatedListenerImpl{}
}

func (l *messageCreatedListenerImpl) ListenerName() string {
	return "MessageCreatedListener"
}

func (l *messageCreatedListenerImpl) GetQueue() string {
	return "message.created"
}

func (l *messageCreatedListenerImpl) GetSubscribeOptions() []common.ISubscribeOption {
	return []common.ISubscribeOption{}
}

func (l *messageCreatedListenerImpl) OnStart() error {
	if l.LoggerMgr != nil {
		l.LoggerMgr.Ins().Info("消息创建监听器已启动")
	}
	return nil
}

func (l *messageCreatedListenerImpl) OnStop() error {
	if l.LoggerMgr != nil {
		l.LoggerMgr.Ins().Info("消息创建监听器已停止")
	}
	return nil
}

func (l *messageCreatedListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error {
	if l.LoggerMgr != nil {
		l.LoggerMgr.Ins().Info("收到消息创建事件",
			"message_id", msg.ID(),
			"body", string(msg.Body()),
			"headers", msg.Headers())
	}
	return nil
}

var _ IMessageCreatedListener = (*messageCreatedListenerImpl)(nil)
var _ common.IBaseListener = (*messageCreatedListenerImpl)(nil)
