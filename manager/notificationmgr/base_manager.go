package notificationmgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/lite-lake/litecore-go/common/deployinfo"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

// notificationConfig 通知配置
type notificationConfig struct {
	Enabled bool
	URL     string
	Timeout time.Duration
}

// notificationManager 服务状态通知管理器实现
type notificationManager struct {
	config    *notificationConfig
	loggerMgr loggermgr.ILoggerManager
	appName   string
	appEnv    string
	hostname  string
	client    *http.Client
}

// newNotificationManager 创建通知管理器实例
func newNotificationManager(
	config *notificationConfig,
	loggerMgr loggermgr.ILoggerManager,
	appName string,
	appEnv string,
) *notificationManager {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	return &notificationManager{
		config:    config,
		loggerMgr: loggerMgr,
		appName:   appName,
		appEnv:    appEnv,
		hostname:  hostname,
	}
}

// ManagerName 返回管理器名称
func (n *notificationManager) ManagerName() string {
	return "NotificationManager"
}

// Health 检查管理器健康状态
func (n *notificationManager) Health() error {
	if !n.config.Enabled {
		return nil
	}
	if n.config.URL == "" {
		return fmt.Errorf("notification url is not configured")
	}
	return nil
}

// OnStart 启动通知管理器，初始化 HTTP 客户端
func (n *notificationManager) OnStart() error {
	if !n.config.Enabled {
		n.logger().Info("服务状态通知已禁用")
		return nil
	}

	n.client = &http.Client{Timeout: n.config.Timeout}
	di := deployinfo.Get()
	n.logger().Info("服务状态通知管理器已启动",
		"url", n.config.URL,
		"app", n.appName,
		"app_env", n.appEnv,
		"deploy_env", di.DeployEnv,
		"hostname", n.hostname,
	)
	return nil
}

// OnStop 停止通知管理器
func (n *notificationManager) OnStop() error {
	n.client = nil
	return nil
}

// IsEnabled 是否启用通知
func (n *notificationManager) IsEnabled() bool {
	return n.config.Enabled && n.config.URL != ""
}

// SendNotification 发送服务状态事件通知（fire-and-forget）
func (n *notificationManager) SendNotification(event string, details map[string]string) error {
	if !n.IsEnabled() {
		return nil
	}

	if n.client == nil {
		return fmt.Errorf("notification manager not started")
	}

	content := buildTextContent(event, n.appName, n.appEnv, n.hostname, details)
	payload := buildWeComTextPayload(content)

	go n.doSend(payload)
	return nil
}

// 重试间隔：15s、30s、60s
var retryIntervals = []time.Duration{15 * time.Second, 30 * time.Second, 60 * time.Second}

// doSend 执行实际的 HTTP 发送（异步调用，失败重试 3 次）
func (n *notificationManager) doSend(payload map[string]interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		n.logger().Warn("通知序列化失败", "error", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, n.config.URL, bytes.NewBuffer(body))
	if err != nil {
		n.logger().Warn("通知请求创建失败", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	maxAttempts := len(retryIntervals) + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := n.sendRequest(payload, attempt, maxAttempts)
		if err == nil {
			return
		}
		if attempt < maxAttempts {
			time.Sleep(retryIntervals[attempt-1])
		}
	}
}

// sendRequest 执行单次 HTTP 请求发送
func (n *notificationManager) sendRequest(payload map[string]interface{}, attempt, maxAttempts int) error {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, n.config.URL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		if attempt < maxAttempts {
			n.logger().Warn("通知发送失败，将进行重试",
				"attempt", fmt.Sprintf("%d/%d", attempt, maxAttempts),
				"wait", retryIntervals[attempt-1].String(),
				"error", err,
			)
		} else {
			n.logger().Warn("通知发送失败，已用尽所有重试次数",
				"attempts", maxAttempts,
				"error", err,
			)
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		if attempt < maxAttempts {
			n.logger().Warn("通知返回错误状态码，将进行重试",
				"attempt", fmt.Sprintf("%d/%d", attempt, maxAttempts),
				"wait", retryIntervals[attempt-1].String(),
				"status", resp.StatusCode,
				"body", string(respBody),
			)
		} else {
			n.logger().Warn("通知返回错误状态码，已用尽所有重试次数",
				"attempts", maxAttempts,
				"status", resp.StatusCode,
				"body", string(respBody),
			)
		}
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	n.logger().Debug("通知发送成功", "event", payload["msgtype"])
	return nil
}

// logger 获取日志器
func (n *notificationManager) logger() loggerLogger {
	return loggerLogger{n.loggerMgr}
}

// loggerLogger 日志适配器，避免 nil loggerMgr panic
type loggerLogger struct {
	mgr loggermgr.ILoggerManager
}

func (l loggerLogger) Info(msg string, args ...any) {
	if l.mgr != nil {
		l.mgr.Ins().Info(msg, args...)
	}
}

func (l loggerLogger) Warn(msg string, args ...any) {
	if l.mgr != nil {
		l.mgr.Ins().Warn(msg, args...)
	}
}

func (l loggerLogger) Debug(msg string, args ...any) {
	if l.mgr != nil {
		l.mgr.Ins().Debug(msg, args...)
	}
}

var _ INotificationManager = (*notificationManager)(nil)
