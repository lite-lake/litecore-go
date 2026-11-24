package privilegemetrics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/common/metrics"
)

// TestInitRendersProductAndAllMetrics Init 后 Render 含 product 默认标签与全部指标名。
func TestInitRendersProductAndAllMetrics(t *testing.T) {
	metrics.ResetForTest()
	t.Cleanup(metrics.ResetForTest)

	Init("whois")

	names := []string{
		MetricPrivilegeRequestsTotal,
		MetricPrivilegeDeniedTotal,
		MetricUCEntitlementRequestsTotal,
		MetricEntitlementUnmappedTotal,
		MetricKeyOwnerMissingTotal,
		MetricQuotaFailClosedTotal,
		MetricQuotaRefundTotal,
		MetricQuotaPointsTotal,
		MetricQuotaWindowRejectsTotal,
		MetricQuotaRefundAfterDayTotal,
		MetricQuotaReservationStaleTotal,
		MetricSSRTicketRejectTotal,
		MetricKeyFormatNewTotal,
		MetricLegacyKeyUsageTotal,
		MetricKeyFormatRetiredTotal,
		MetricKeyDormantTotal,
		MetricInternalKeyFanoutTotal,
		MetricInternalKeyFanoutPartialTotal,
		MetricSiteBudgetRemovedTotal,
	}
	output := metrics.Render()
	for _, name := range names {
		assert.Contains(t, output, "# TYPE "+name+" ", "指标 %s 未注册", name)
	}

	RequestObserved("free", "allow", "session")
	KeyOwnerMissing()
	output = metrics.Render()
	assert.Contains(t, output, `privilege_requests_total{decision="allow",product="whois",source="session",tier="free"} 1`)
	// 单标签指标只带默认 product 标签，证明 Init 的默认标签生效。
	assert.Contains(t, output, `key_owner_missing_total{product="whois"} 1`)
}

// TestTypedFunctions 类型化函数写入正确的指标名、类型与标签。
func TestTypedFunctions(t *testing.T) {
	metrics.ResetForTest()
	t.Cleanup(metrics.ResetForTest)

	Init("dns")

	Denied("free", "TIER_DISABLED")
	UCEntitlementRequest("fail_closed")
	EntitlementUnmapped("paid_level_9")
	KeyOwnerMissing()
	QuotaFailClosed("redis")
	QuotaRefund("upstream_error")
	QuotaPoints("convert", 2.5)
	QuotaWindowReject("pro", "minute")
	QuotaRefundAfterDay()
	QuotaReservationStale("no_request", 3)
	SSRTicketReject("expired")
	KeyFormatNew()
	LegacyKeyUsage()
	KeyFormatRetired()
	KeyDormant(2)
	InternalKeyFanout("partial")
	InternalKeyFanoutPartial("create")
	SiteBudgetRemoved()

	output := metrics.Render()

	expected := []string{
		`privilege_denied_total{product="dns",reason="TIER_DISABLED",tier="free"} 1`,
		`uc_entitlement_requests_total{product="dns",result="fail_closed"} 1`,
		`entitlement_unmapped_total{product="dns",tierLevel="paid_level_9"} 1`,
		`key_owner_missing_total{product="dns"} 1`,
		`quota_fail_closed_total{product="dns",store="redis"} 1`,
		`quota_refund_total{product="dns",reason="upstream_error"} 1`,
		`quota_points_total{bucket="convert",product="dns"} 2.5`,
		`quota_window_rejects_total{product="dns",tier="pro",window="minute"} 1`,
		`quota_refund_after_day_total{product="dns"} 1`,
		`quota_reservation_stale_total{product="dns",reason="no_request"} 3`,
		`ssr_ticket_reject_total{product="dns",reason="expired"} 1`,
		`key_format_new_total{product="dns"} 1`,
		`legacy_key_usage_total{product="dns"} 1`,
		`key_format_retired_total{product="dns"} 1`,
		`key_dormant_total{product="dns"} 2`,
		`internal_key_fanout_total{product="dns",result="partial"} 1`,
		`internal_key_fanout_partial_total{operation="create",product="dns"} 1`,
		`site_budget_removed_total{product="dns"} 1`,
	}
	for _, line := range expected {
		assert.Contains(t, output, line)
	}
}

// TestInitIdempotent 重复 Init 不产生重复的 HELP/TYPE 行。
func TestInitIdempotent(t *testing.T) {
	metrics.ResetForTest()
	t.Cleanup(metrics.ResetForTest)

	Init("litecreate")
	Init("litecreate")

	output := metrics.Render()
	assert.Equal(t, 1, strings.Count(output, "# TYPE "+MetricPrivilegeRequestsTotal+" counter\n"))
}

// TestMetricNameCatalog 指标名与 PRD 契约逐字一致。
func TestMetricNameCatalog(t *testing.T) {
	cases := map[string]string{
		"privilege_requests_total":          MetricPrivilegeRequestsTotal,
		"privilege_denied_total":            MetricPrivilegeDeniedTotal,
		"uc_entitlement_requests_total":     MetricUCEntitlementRequestsTotal,
		"entitlement_unmapped_total":        MetricEntitlementUnmappedTotal,
		"key_owner_missing_total":           MetricKeyOwnerMissingTotal,
		"quota_fail_closed_total":           MetricQuotaFailClosedTotal,
		"quota_refund_total":                MetricQuotaRefundTotal,
		"quota_points_total":                MetricQuotaPointsTotal,
		"quota_window_rejects_total":        MetricQuotaWindowRejectsTotal,
		"quota_refund_after_day_total":      MetricQuotaRefundAfterDayTotal,
		"quota_reservation_stale_total":     MetricQuotaReservationStaleTotal,
		"ssr_ticket_reject_total":           MetricSSRTicketRejectTotal,
		"key_format_new_total":              MetricKeyFormatNewTotal,
		"legacy_key_usage_total":            MetricLegacyKeyUsageTotal,
		"key_format_retired_total":          MetricKeyFormatRetiredTotal,
		"key_dormant_total":                 MetricKeyDormantTotal,
		"internal_key_fanout_total":         MetricInternalKeyFanoutTotal,
		"internal_key_fanout_partial_total": MetricInternalKeyFanoutPartialTotal,
		"site_budget_removed_total":         MetricSiteBudgetRemovedTotal,
	}
	for want, got := range cases {
		assert.Equal(t, want, got)
	}
}
