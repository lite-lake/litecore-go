package common

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testPaginationEntity 分页测试用实体
type testPaginationEntity struct {
	BaseEntityWithTimestamps
	Name string `json:"name"`
}

func (e *testPaginationEntity) EntityName() string { return "TestPaginationEntity" }
func (e *testPaginationEntity) TableName() string  { return "test_pagination_entities" }
func (e *testPaginationEntity) GetId() string      { return e.ID }

func newPageTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Dialector{DriverName: "sqlite", DSN: ":memory:"}, &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&testPaginationEntity{}))
	return db
}

func TestParsePageQuery(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		wantPage     int
		wantPageSize int
	}{
		{name: "缺省参数", query: "", wantPage: 1, wantPageSize: 20},
		{name: "正常参数", query: "page=3&page_size=50", wantPage: 3, wantPageSize: 50},
		{name: "page 为 0", query: "page=0&page_size=20", wantPage: 1, wantPageSize: 20},
		{name: "page 为负数", query: "page=-2&page_size=20", wantPage: 1, wantPageSize: 20},
		{name: "page_size 为 0", query: "page=1&page_size=0", wantPage: 1, wantPageSize: 20},
		{name: "page_size 超上限", query: "page=1&page_size=1000", wantPage: 1, wantPageSize: 100},
		{name: "page_size 恰好上限", query: "page=1&page_size=100", wantPage: 1, wantPageSize: 100},
		{name: "非法数值", query: "page=abc&page_size=xyz", wantPage: 1, wantPageSize: 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			req := httptest.NewRequest("GET", "/test?"+tt.query, nil)
			ctx.Request = req

			page, pageSize := ParsePageQuery(ctx)
			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantPageSize, pageSize)
		})
	}
}

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{name: "正常值不变", page: 5, pageSize: 30, wantPage: 5, wantPageSize: 30},
		{name: "page 为 0", page: 0, pageSize: 20, wantPage: 1, wantPageSize: 20},
		{name: "page 为负数", page: -1, pageSize: 20, wantPage: 1, wantPageSize: 20},
		{name: "pageSize 为 0", page: 1, pageSize: 0, wantPage: 1, wantPageSize: 20},
		{name: "pageSize 为负数", page: 1, pageSize: -5, wantPage: 1, wantPageSize: 20},
		{name: "pageSize 超上限", page: 1, pageSize: 200, wantPage: 1, wantPageSize: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, pageSize := NormalizePage(tt.page, tt.pageSize)
			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantPageSize, pageSize)
		})
	}
}

func TestPaginate(t *testing.T) {
	db := newPageTestDB(t)

	rows := []testPaginationEntity{}
	for i := 0; i < 5; i++ {
		rows = append(rows, testPaginationEntity{Name: "seed"})
	}
	assert.NoError(t, db.Create(&rows).Error)

	got := []testPaginationEntity{}
	assert.NoError(t, db.Model(&testPaginationEntity{}).Scopes(Paginate(2, 2)).Find(&got).Error)
	assert.Len(t, got, 2)
}

func TestPageList(t *testing.T) {
	db := newPageTestDB(t)

	rows := []testPaginationEntity{}
	for i := 0; i < 25; i++ {
		rows = append(rows, testPaginationEntity{Name: "seed-" + string(rune('a'+i%3))})
	}
	assert.NoError(t, db.Create(&rows).Error)

	t.Run("第 1 页每页 10 条", func(t *testing.T) {
		got := []testPaginationEntity{}
		total, err := PageList(db.Model(&testPaginationEntity{}), &got, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(25), total)
		assert.Len(t, got, 10)
	})

	t.Run("第 3 页每页 10 条", func(t *testing.T) {
		got := []testPaginationEntity{}
		total, err := PageList(db.Model(&testPaginationEntity{}), &got, 3, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(25), total)
		assert.Len(t, got, 5)
	})

	t.Run("pageSize 超上限自动钳制", func(t *testing.T) {
		got := []testPaginationEntity{}
		total, err := PageList(db.Model(&testPaginationEntity{}), &got, 1, 1000)
		assert.NoError(t, err)
		assert.Equal(t, int64(25), total)
		assert.Len(t, got, 25)
	})

	t.Run("过滤条件保留", func(t *testing.T) {
		got := []testPaginationEntity{}
		total, err := PageList(db.Model(&testPaginationEntity{}).Where("name = ?", "seed-a"), &got, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(9), total)
		assert.Len(t, got, 9)
	})
}
