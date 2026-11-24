package common

const (
	HTTPStatusContinue           = 100 // HTTP/1.1: Continue
	HTTPStatusSwitchingProtocols = 101 // HTTP/1.1: Switching Protocols

	HTTPStatusOK                   = 200 // HTTP/1.1: OK
	HTTPStatusCreated              = 201 // HTTP/1.1: Created
	HTTPStatusAccepted             = 202 // HTTP/1.1: Accepted
	HTTPStatusNonAuthoritativeInfo = 203 // HTTP/1.1: Non-Authoritative Information
	HTTPStatusNoContent            = 204 // HTTP/1.1: No Content
	HTTPStatusResetContent         = 205 // HTTP/1.1: Reset Content
	HTTPStatusPartialContent       = 206 // HTTP/1.1: Partial Content

	HTTPStatusMultipleChoices   = 300 // HTTP/1.1: Multiple Choices
	HTTPStatusMovedPermanently  = 301 // HTTP/1.1: Moved Permanently
	HTTPStatusFound             = 302 // HTTP/1.1: Found
	HTTPStatusSeeOther          = 303 // HTTP/1.1: See Other
	HTTPStatusNotModified       = 304 // HTTP/1.1: Not Modified
	HTTPStatusUseProxy          = 305 // HTTP/1.1: Use Proxy
	HTTPStatusTemporaryRedirect = 307 // HTTP/1.1: Temporary Redirect
	HTTPStatusPermanentRedirect = 308 // HTTP/1.1: Permanent Redirect

	HTTPStatusBadRequest                  = 400 // HTTP/1.1: Bad Request
	HTTPStatusUnauthorized                = 401 // HTTP/1.1: Unauthorized
	HTTPStatusPaymentRequired             = 402 // HTTP/1.1: Payment Required
	HTTPStatusForbidden                   = 403 // HTTP/1.1: Forbidden
	HTTPStatusNotFound                    = 404 // HTTP/1.1: Not Found
	HTTPStatusMethodNotAllowed            = 405 // HTTP/1.1: Method Not Allowed
	HTTPStatusNotAcceptable               = 406 // HTTP/1.1: Not Acceptable
	HTTPStatusProxyAuthRequired           = 407 // HTTP/1.1: Proxy Authentication Required
	HTTPStatusRequestTimeout              = 408 // HTTP/1.1: Request Timeout
	HTTPStatusConflict                    = 409 // HTTP/1.1: Conflict
	HTTPStatusGone                        = 410 // HTTP/1.1: Gone
	HTTPStatusLengthRequired              = 411 // HTTP/1.1: Length Required
	HTTPStatusPreconditionFailed          = 412 // HTTP/1.1: Precondition Failed
	HTTPStatusPayloadTooLarge             = 413 // HTTP/1.1: Payload Too Large
	HTTPStatusURITooLong                  = 414 // HTTP/1.1: URI Too Long
	HTTPStatusUnsupportedMediaType        = 415 // HTTP/1.1: Unsupported Media Type
	HTTPStatusRangeNotSatisfiable         = 416 // HTTP/1.1: Range Not Satisfiable
	HTTPStatusExpectationFailed           = 417 // HTTP/1.1: Expectation Failed
	HTTPStatusTeapot                      = 418 // HTTP/1.1: I'm a teapot
	HTTPStatusMisdirectedRequest          = 421 // HTTP/1.1: Misdirected Request
	HTTPStatusUnprocessableEntity         = 422 // HTTP/1.1: Unprocessable Entity
	HTTPStatusLocked                      = 423 // HTTP/1.1: Locked
	HTTPStatusFailedDependency            = 424 // HTTP/1.1: Failed Dependency
	HTTPStatusTooEarly                    = 425 // HTTP/1.1: Too Early
	HTTPStatusUpgradeRequired             = 426 // HTTP/1.1: Upgrade Required
	HTTPStatusPreconditionRequired        = 428 // HTTP/1.1: Precondition Required
	HTTPStatusTooManyRequests             = 429 // HTTP/1.1: Too Many Requests
	HTTPStatusRequestHeaderFieldsTooLarge = 431 // HTTP/1.1: Request Header Fields Too Large
	HTTPStatusUnavailableForLegalReasons  = 451 // HTTP/1.1: Unavailable For Legal Reasons

	HTTPStatusInternalServerError           = 500 // HTTP/1.1: Internal Server Error
	HTTPStatusNotImplemented                = 501 // HTTP/1.1: Not Implemented
	HTTPStatusBadGateway                    = 502 // HTTP/1.1: Bad Gateway
	HTTPStatusServiceUnavailable            = 503 // HTTP/1.1: Service Unavailable
	HTTPStatusGatewayTimeout                = 504 // HTTP/1.1: Gateway Timeout
	HTTPStatusHTTPVersionNotSupported       = 505 // HTTP/1.1: HTTP Version Not Supported
	HTTPStatusVariantAlsoNegotiates         = 506 // HTTP/1.1: Variant Also Negotiates
	HTTPStatusInsufficientStorage           = 507 // HTTP/1.1: Insufficient Storage
	HTTPStatusLoopDetected                  = 508 // HTTP/1.1: Loop Detected
	HTTPStatusNotExtended                   = 510 // HTTP/1.1: Not Extended
	HTTPStatusNetworkAuthenticationRequired = 511 // HTTP/1.1: Network Authentication Required
)
