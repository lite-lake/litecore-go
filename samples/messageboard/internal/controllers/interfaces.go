// Package controllers 定义 HTTP 控制器接口
package controllers

import "com.litelake.litecore/common"

// IGetMessagesController 获取留言控制器接口
type IGetMessagesController interface {
	common.BaseController
}

// ICreateMessageController 创建留言控制器接口
type ICreateMessageController interface {
	common.BaseController
}

// IAdminLoginController 管理员登录控制器接口
type IAdminLoginController interface {
	common.BaseController
}

// IGetAllMessagesController 获取所有留言控制器接口
type IGetAllMessagesController interface {
	common.BaseController
}

// IUpdateStatusController 更新状态控制器接口
type IUpdateStatusController interface {
	common.BaseController
}

// IDeleteMessageController 删除留言控制器接口
type IDeleteMessageController interface {
	common.BaseController
}
