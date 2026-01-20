package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPStatusCodes_1xx信息响应(t *testing.T) {
	assert.Equal(t, 100, HTTPStatusContinue)
	assert.Equal(t, 101, HTTPStatusSwitchingProtocols)
}

func TestHTTPStatusCodes_2xx成功响应(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   int
	}{
		{"OK", HTTPStatusOK, 200},
		{"Created", HTTPStatusCreated, 201},
		{"Accepted", HTTPStatusAccepted, 202},
		{"Non-Authoritative Info", HTTPStatusNonAuthoritativeInfo, 203},
		{"No Content", HTTPStatusNoContent, 204},
		{"Reset Content", HTTPStatusResetContent, 205},
		{"Partial Content", HTTPStatusPartialContent, 206},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status)
		})
	}
}

func TestHTTPStatusCodes_3xx重定向(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   int
	}{
		{"Multiple Choices", HTTPStatusMultipleChoices, 300},
		{"Moved Permanently", HTTPStatusMovedPermanently, 301},
		{"Found", HTTPStatusFound, 302},
		{"See Other", HTTPStatusSeeOther, 303},
		{"Not Modified", HTTPStatusNotModified, 304},
		{"Use Proxy", HTTPStatusUseProxy, 305},
		{"Temporary Redirect", HTTPStatusTemporaryRedirect, 307},
		{"Permanent Redirect", HTTPStatusPermanentRedirect, 308},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status)
		})
	}
}

func TestHTTPStatusCodes_4xx客户端错误(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   int
	}{
		{"Bad Request", HTTPStatusBadRequest, 400},
		{"Unauthorized", HTTPStatusUnauthorized, 401},
		{"Payment Required", HTTPStatusPaymentRequired, 402},
		{"Forbidden", HTTPStatusForbidden, 403},
		{"Not Found", HTTPStatusNotFound, 404},
		{"Method Not Allowed", HTTPStatusMethodNotAllowed, 405},
		{"Not Acceptable", HTTPStatusNotAcceptable, 406},
		{"Proxy Auth Required", HTTPStatusProxyAuthRequired, 407},
		{"Request Timeout", HTTPStatusRequestTimeout, 408},
		{"Conflict", HTTPStatusConflict, 409},
		{"Gone", HTTPStatusGone, 410},
		{"Length Required", HTTPStatusLengthRequired, 411},
		{"Precondition Failed", HTTPStatusPreconditionFailed, 412},
		{"Payload Too Large", HTTPStatusPayloadTooLarge, 413},
		{"URI Too Long", HTTPStatusURITooLong, 414},
		{"Unsupported Media Type", HTTPStatusUnsupportedMediaType, 415},
		{"Range Not Satisfiable", HTTPStatusRangeNotSatisfiable, 416},
		{"Expectation Failed", HTTPStatusExpectationFailed, 417},
		{"I'm a teapot", HTTPStatusTeapot, 418},
		{"Misdirected Request", HTTPStatusMisdirectedRequest, 421},
		{"Unprocessable Entity", HTTPStatusUnprocessableEntity, 422},
		{"Locked", HTTPStatusLocked, 423},
		{"Failed Dependency", HTTPStatusFailedDependency, 424},
		{"Too Early", HTTPStatusTooEarly, 425},
		{"Upgrade Required", HTTPStatusUpgradeRequired, 426},
		{"Precondition Required", HTTPStatusPreconditionRequired, 428},
		{"Too Many Requests", HTTPStatusTooManyRequests, 429},
		{"Request Header Fields Too Large", HTTPStatusRequestHeaderFieldsTooLarge, 431},
		{"Unavailable For Legal Reasons", HTTPStatusUnavailableForLegalReasons, 451},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status)
		})
	}
}

func TestHTTPStatusCodes_5xx服务器错误(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   int
	}{
		{"Internal Server Error", HTTPStatusInternalServerError, 500},
		{"Not Implemented", HTTPStatusNotImplemented, 501},
		{"Bad Gateway", HTTPStatusBadGateway, 502},
		{"Service Unavailable", HTTPStatusServiceUnavailable, 503},
		{"Gateway Timeout", HTTPStatusGatewayTimeout, 504},
		{"HTTP Version Not Supported", HTTPStatusHTTPVersionNotSupported, 505},
		{"Variant Also Negotiates", HTTPStatusVariantAlsoNegotiates, 506},
		{"Insufficient Storage", HTTPStatusInsufficientStorage, 507},
		{"Loop Detected", HTTPStatusLoopDetected, 508},
		{"Not Extended", HTTPStatusNotExtended, 510},
		{"Network Authentication Required", HTTPStatusNetworkAuthenticationRequired, 511},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status)
		})
	}
}

func TestHTTPStatusCodes_常用状态码(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   int
	}{
		{"成功", HTTPStatusOK, 200},
		{"未找到", HTTPStatusNotFound, 404},
		{"内部服务器错误", HTTPStatusInternalServerError, 500},
		{"错误请求", HTTPStatusBadRequest, 400},
		{"未授权", HTTPStatusUnauthorized, 401},
		{"禁止访问", HTTPStatusForbidden, 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status)
		})
	}
}

func TestHTTPStatusCodes_常量唯一性(t *testing.T) {
	uniqueCodes := make(map[int]string)

	codes := []struct {
		name   string
		status int
	}{
		{"Continue", HTTPStatusContinue},
		{"OK", HTTPStatusOK},
		{"Created", HTTPStatusCreated},
		{"BadRequest", HTTPStatusBadRequest},
		{"NotFound", HTTPStatusNotFound},
		{"InternalServerError", HTTPStatusInternalServerError},
	}

	for _, code := range codes {
		if existing, exists := uniqueCodes[code.status]; exists {
			t.Errorf("状态码 %d 被多个常量使用: %s 和 %s", code.status, existing, code.name)
		}
		uniqueCodes[code.status] = code.name
	}
}
