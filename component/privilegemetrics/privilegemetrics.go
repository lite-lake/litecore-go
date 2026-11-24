// Package privilegemetrics 定义八产品统一的权限/配额/Key 生命周期指标，并提供类型化打点函数。
//
// 目的：让各产品以同一套指标名、help 与标签集打点，避免拼写漂移；所有 series 通过
// common/metrics 的进程内注册表输出为 Prometheus 文本（/metrics）。
//
// 使用方式：
//
//	privilegemetrics.Init("whois")            // 进程启动期：设置 product 默认标签并注册全部指标
//	privilegemetrics.RequestObserved("free", "allow", "session")
//
// 标签基数约束：只允许有界枚举（tier/decision/source/reason/store/window/bucket/
// result/operation/tierLevel），禁止 IP、key_id、request_id、user_id 等无界维度。
package privilegemetrics

import "github.com/lite-lake/litecore-go/common/metrics"

// 指标名（对外契约，勿改名）。
const (
	// MetricPrivilegeRequestsTotal 身份解析请求总数。
	MetricPrivilegeRequestsTotal = "privilege_requests_total"
	// MetricPrivilegeDeniedTotal 权限拒绝总数。
	MetricPrivilegeDeniedTotal = "privilege_denied_total"
	// MetricUCEntitlementRequestsTotal UC 权益请求总数。
	MetricUCEntitlementRequestsTotal = "uc_entitlement_requests_total"
	// MetricEntitlementUnmappedTotal 权益档位未映射总数。
	MetricEntitlementUnmappedTotal = "entitlement_unmapped_total"
	// MetricKeyOwnerMissingTotal API Key 归属缺失总数。
	MetricKeyOwnerMissingTotal = "key_owner_missing_total"
	// MetricQuotaFailClosedTotal 配额存储故障 fail-closed 总数。
	MetricQuotaFailClosedTotal = "quota_fail_closed_total"
	// MetricQuotaRefundTotal 配额退款总数。
	MetricQuotaRefundTotal = "quota_refund_total"
	// MetricQuotaPointsTotal 配额点数扣减总量。
	MetricQuotaPointsTotal = "quota_points_total"
	// MetricQuotaWindowRejectsTotal 多窗口限流拒绝总数。
	MetricQuotaWindowRejectsTotal = "quota_window_rejects_total"
	// MetricQuotaRefundAfterDayTotal 跨日配额退款总数。
	MetricQuotaRefundAfterDayTotal = "quota_refund_after_day_total"
	// MetricQuotaReservationStaleTotal 配额预留清扫发现的陈旧条目数。
	MetricQuotaReservationStaleTotal = "quota_reservation_stale_total"
	// MetricSSRTicketRejectTotal SSR 票据拒绝总数。
	MetricSSRTicketRejectTotal = "ssr_ticket_reject_total"
	// MetricKeyFormatNewTotal 新建 API Key 新格式计数。
	MetricKeyFormatNewTotal = "key_format_new_total"
	// MetricLegacyKeyUsageTotal 旧格式 API Key 命中计数。
	MetricLegacyKeyUsageTotal = "legacy_key_usage_total"
	// MetricKeyFormatRetiredTotal 旧格式 API Key 退役后命中计数。
	MetricKeyFormatRetiredTotal = "key_format_retired_total"
	// MetricKeyDormantTotal 休眠 API Key 数量。
	MetricKeyDormantTotal = "key_dormant_total"
	// MetricInternalKeyFanoutTotal AC 跨产品创建/吊销内部 Key 次数。
	MetricInternalKeyFanoutTotal = "internal_key_fanout_total"
	// MetricInternalKeyFanoutPartialTotal AC 内部 Key 部分失败次数。
	MetricInternalKeyFanoutPartialTotal = "internal_key_fanout_partial_total"
	// MetricSiteBudgetRemovedTotal Create 站点预算迁移移除确认（一次性）。
	MetricSiteBudgetRemovedTotal = "site_budget_removed_total"
)

// Init 设置 product 默认标签并注册全部指标（幂等，可在启动期安全重复调用）。
//
// product 取值固定为产品短名（whois / ipregion / dns / nsfwdetector / liteconvert /
// websitescreenshot / litetranslate / litecreate），不得使用域名或环境后缀。
func Init(product string) {
	metrics.Register(MetricPrivilegeRequestsTotal, "身份解析请求总数（按档位/决策/来源）", metrics.Counter)
	metrics.Register(MetricPrivilegeDeniedTotal, "权限拒绝总数（按档位/原因）", metrics.Counter)
	metrics.Register(MetricUCEntitlementRequestsTotal, "UC 权益请求总数（按结果：ok/stale/fail_closed/fallback）", metrics.Counter)
	metrics.Register(MetricEntitlementUnmappedTotal, "权益档位未映射总数（按 tierLevel）", metrics.Counter)
	metrics.Register(MetricKeyOwnerMissingTotal, "API Key 归属缺失总数", metrics.Counter)
	metrics.Register(MetricQuotaFailClosedTotal, "配额存储故障 fail-closed 总数（按存储）", metrics.Counter)
	metrics.Register(MetricQuotaRefundTotal, "配额退款总数（按原因）", metrics.Counter)
	metrics.Register(MetricQuotaPointsTotal, "配额点数扣减总量（按桶）", metrics.Counter)
	metrics.Register(MetricQuotaWindowRejectsTotal, "多窗口限流拒绝总数（按档位/窗口）", metrics.Counter)
	metrics.Register(MetricQuotaRefundAfterDayTotal, "跨日配额退款总数", metrics.Counter)
	metrics.Register(MetricQuotaReservationStaleTotal, "配额预留清扫发现的陈旧条目数（按原因快照：no_request/settle_failed/live_pending/ttl_expired/state_mismatch）", metrics.Gauge)
	metrics.Register(MetricSSRTicketRejectTotal, "SSR 票据拒绝总数（按原因）", metrics.Counter)
	metrics.Register(MetricKeyFormatNewTotal, "新建 API Key 新格式计数", metrics.Counter)
	metrics.Register(MetricLegacyKeyUsageTotal, "旧格式 API Key 命中计数", metrics.Counter)
	metrics.Register(MetricKeyFormatRetiredTotal, "旧格式 API Key 退役后命中计数", metrics.Counter)
	metrics.Register(MetricKeyDormantTotal, "休眠 API Key 数量", metrics.Gauge)
	metrics.Register(MetricInternalKeyFanoutTotal, "AC 跨产品创建/吊销内部 Key 次数（按结果）", metrics.Counter)
	metrics.Register(MetricInternalKeyFanoutPartialTotal, "AC 内部 Key 部分失败次数（按操作）", metrics.Counter)
	metrics.Register(MetricSiteBudgetRemovedTotal, "Create 站点预算迁移移除确认（一次性）", metrics.Gauge)

	metrics.SetDefaultLabels(metrics.Labels{"product": product})
}

// RequestObserved 记录一次身份解析决策。
func RequestObserved(tier, decision, source string) {
	metrics.Add(MetricPrivilegeRequestsTotal, 1, metrics.Labels{
		"tier":     tier,
		"decision": decision,
		"source":   source,
	})
}

// Denied 记录一次权限拒绝。
func Denied(tier, reason string) {
	metrics.Add(MetricPrivilegeDeniedTotal, 1, metrics.Labels{
		"tier":   tier,
		"reason": reason,
	})
}

// UCEntitlementRequest 记录一次 UC 权益请求结果。
func UCEntitlementRequest(result string) {
	metrics.Add(MetricUCEntitlementRequestsTotal, 1, metrics.Labels{"result": result})
}

// EntitlementUnmapped 记录一次权益档位未映射。
func EntitlementUnmapped(tierLevel string) {
	metrics.Add(MetricEntitlementUnmappedTotal, 1, metrics.Labels{"tierLevel": tierLevel})
}

// KeyOwnerMissing 记录一次 API Key 归属缺失。
func KeyOwnerMissing() {
	metrics.Add(MetricKeyOwnerMissingTotal, 1, nil)
}

// QuotaFailClosed 记录一次配额存储故障 fail-closed。
func QuotaFailClosed(store string) {
	metrics.Add(MetricQuotaFailClosedTotal, 1, metrics.Labels{"store": store})
}

// QuotaRefund 记录一次配额退款。
func QuotaRefund(reason string) {
	metrics.Add(MetricQuotaRefundTotal, 1, metrics.Labels{"reason": reason})
}

// QuotaPoints 记录按桶扣减的配额点数。
func QuotaPoints(bucket string, points float64) {
	metrics.Add(MetricQuotaPointsTotal, points, metrics.Labels{"bucket": bucket})
}

// QuotaWindowReject 记录一次多窗口限流拒绝。
func QuotaWindowReject(tier, window string) {
	metrics.Add(MetricQuotaWindowRejectsTotal, 1, metrics.Labels{
		"tier":   tier,
		"window": window,
	})
}

// QuotaRefundAfterDay 记录一次跨日配额退款。
func QuotaRefundAfterDay() {
	metrics.Add(MetricQuotaRefundAfterDayTotal, 1, nil)
}

// QuotaReservationStale 设置按原因分组的配额预留陈旧条目快照数。
// reason 为有界枚举（no_request/settle_failed/live_pending/ttl_expired/state_mismatch），禁止 request_id/IP/用户数据。
func QuotaReservationStale(reason string, n float64) {
	metrics.Set(MetricQuotaReservationStaleTotal, n, metrics.Labels{"reason": reason})
}

// SSRTicketReject 记录一次 SSR 票据拒绝。
func SSRTicketReject(reason string) {
	metrics.Add(MetricSSRTicketRejectTotal, 1, metrics.Labels{"reason": reason})
}

// KeyFormatNew 记录一次新建 API Key 新格式判定。
func KeyFormatNew() {
	metrics.Add(MetricKeyFormatNewTotal, 1, nil)
}

// LegacyKeyUsage 记录一次旧格式 API Key 命中。
func LegacyKeyUsage() {
	metrics.Add(MetricLegacyKeyUsageTotal, 1, nil)
}

// KeyFormatRetired 记录一次旧格式 API Key 退役后命中。
func KeyFormatRetired() {
	metrics.Add(MetricKeyFormatRetiredTotal, 1, nil)
}

// KeyDormant 设置休眠 API Key 数量。
func KeyDormant(n float64) {
	metrics.Set(MetricKeyDormantTotal, n, nil)
}

// InternalKeyFanout 记录一次 AC 跨产品创建/吊销内部 Key 的结果。
func InternalKeyFanout(result string) {
	metrics.Add(MetricInternalKeyFanoutTotal, 1, metrics.Labels{"result": result})
}

// InternalKeyFanoutPartial 记录一次 AC 内部 Key 部分失败。
func InternalKeyFanoutPartial(op string) {
	metrics.Add(MetricInternalKeyFanoutPartialTotal, 1, metrics.Labels{"operation": op})
}

// SiteBudgetRemoved 设置 Create 站点预算迁移移除确认值（一次性）。
func SiteBudgetRemoved() {
	metrics.Set(MetricSiteBudgetRemovedTotal, 1, nil)
}
