package liteadmin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidAdminID(t *testing.T) {
	tests := []struct {
		name    string
		adminID string
		want    bool
	}{
		{name: "1 位数字合法", adminID: "1", want: true},
		{name: "18 位数字合法（bigint 上限）", adminID: "123456789012345678", want: true},
		{name: "常规 ID 合法", adminID: "10086", want: true},
		{name: "空串非法", adminID: "", want: false},
		{name: "19 位数字非法（超长）", adminID: "1234567890123456789", want: false},
		{name: "含字母非法", adminID: "12a34", want: false},
		{name: "含符号非法", adminID: "-1", want: false},
		{name: "含空格非法", adminID: " 1", want: false},
		{name: "小数点非法", adminID: "1.0", want: false},
		{name: "全角数字非法", adminID: "１２３", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidAdminID(tt.adminID))
		})
	}
}

func Test协议头常量(t *testing.T) {
	// 头名是协议契约，防手误改值（变更须同步 PRD 附录 A 与 SOP §3.8）
	assert.Equal(t, "X-Lite-Admin-Key", HeaderAdminKey)
	assert.Equal(t, "X-Lite-Admin-Id", HeaderAdminID)
	assert.Equal(t, "X-Lite-Admin-Username", HeaderAdminUsername)
	assert.Equal(t, "X-Lite-Gate-Trace-Id", HeaderGateTraceID)
	assert.Equal(t, "traceparent", HeaderTraceparent)
}

func TestNewDelegatedSession(t *testing.T) {
	t.Run("完整字段", func(t *testing.T) {
		identity := NewDelegatedSession("42", "kent")
		assert.Equal(t, SubjectTypeAdmin, identity.SubjectType)
		assert.Equal(t, "admin", identity.SubjectType)
		assert.Equal(t, "42", identity.SubjectID)
		assert.Equal(t, "kent", identity.Username)
	})
	t.Run("用户名可空", func(t *testing.T) {
		identity := NewDelegatedSession("42", "")
		assert.Equal(t, SubjectTypeAdmin, identity.SubjectType)
		assert.Equal(t, "", identity.Username)
	})
}
