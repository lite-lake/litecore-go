package common

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 分页默认值常量，所有使用分页查询的接口统一遵循
const (
	DefaultPage     = 1   // 默认页码
	DefaultPageSize = 20  // 默认每页条数
	MaxPageSize     = 100 // 每页条数上限，防止单次查询数据量过大
)

// ParsePageQuery 从 gin 请求中解析分页参数（page / page_size），并做钳制。
// 缺失或非法值时回退默认值：page→DefaultPage，page_size→DefaultPageSize；
// page_size 超过 MaxPageSize 时钳制为 MaxPageSize。
func ParsePageQuery(ctx *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize)))
	return NormalizePage(page, pageSize)
}

// NormalizePage 校验并钳制分页参数：
// page < 1 → DefaultPage；pageSize < 1 → DefaultPageSize；pageSize > MaxPageSize → MaxPageSize。
func NormalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

// Paginate 返回 GORM 分页 Scope，计算 Offset/Limit 并应用到查询链。
// 传入的 page/pageSize 应已通过 NormalizePage 钳制，本函数不再校验。
func Paginate(page, pageSize int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((page - 1) * pageSize).Limit(pageSize)
	}
}

// PageList 分页查询：先统计总数（Count），再查询当前页数据（Find），返回 total。
// 传入的 db 应已包含过滤条件（Where 等）与排序（Order），本函数追加分页条件；
// 内部自动做分页参数钳制，page/pageSize 可直接传原始值。
func PageList(db *gorm.DB, out interface{}, page, pageSize int) (int64, error) {
	page, pageSize = NormalizePage(page, pageSize)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}

	if err := db.Scopes(Paginate(page, pageSize)).Find(out).Error; err != nil {
		return 0, err
	}

	return total, nil
}
