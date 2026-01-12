// Package middlewares 定义 HTTP 中间件接口
package middlewares

import "com.litelake.litecore/common"

// IAuthMiddleware 认证中间件接口
type IAuthMiddleware interface {
	common.BaseMiddleware
}
