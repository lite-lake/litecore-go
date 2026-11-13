package notificationmgr

import "github.com/lite-lake/litecore-go/common"

// INotificationManager 服务状态通知管理器接口
type INotificationManager interface {
	common.IBaseManager

	// SendNotification 发送服务状态事件通知
	// event: 事件类型 (starting, started, start_failed, stopping, stopped)
	// details: 附加信息键值对
	SendNotification(event string, details map[string]string) error

	// IsEnabled 是否启用通知
	IsEnabled() bool
}
