package container

import (
	"github.com/gin-gonic/gin"
	"com.litelake.litecore/common"
)

// MockConfigProvider Mock 配置提供者
type MockConfigProvider struct {
	name string
}

func (m *MockConfigProvider) ConfigProviderName() string {
	return m.name
}

func (m *MockConfigProvider) Get(key string) (any, error) {
	return nil, nil
}

func (m *MockConfigProvider) Has(key string) bool {
	return false
}

// MockManager Mock 管理器
type MockManager struct {
	name   string
	Config common.BaseConfigProvider `inject:""`
}

func (m *MockManager) ManagerName() string {
	return m.name
}

func (m *MockManager) Health() error {
	return nil
}

func (m *MockManager) OnStart() error {
	return nil
}

func (m *MockManager) OnStop() error {
	return nil
}

// MockEntity Mock 实体
type MockEntity struct {
	name string
	id   string
}

func (m *MockEntity) EntityName() string {
	return m.name
}

func (m *MockEntity) TableName() string {
	return "mock_entities"
}

func (m *MockEntity) GetId() string {
	return m.id
}

// MockRepository Mock 存储库
type MockRepository struct {
	name      string
	Config    common.BaseConfigProvider `inject:""`
	Manager   common.BaseManager       `inject:""`
	Entity    common.BaseEntity        `inject:""`
}

func (m *MockRepository) RepositoryName() string {
	return m.name
}

func (m *MockRepository) OnStart() error {
	return nil
}

func (m *MockRepository) OnStop() error {
	return nil
}

// MockService Mock 服务
type MockService struct {
	name   string
	Config common.BaseConfigProvider `inject:""`
	Repo   common.BaseRepository    `inject:""`
}

func (m *MockService) ServiceName() string {
	return m.name
}

func (m *MockService) OnStart() error {
	return nil
}

func (m *MockService) OnStop() error {
	return nil
}

// MockController Mock 控制器
type MockController struct {
	name    string
	Service common.BaseService `inject:""`
}

func (m *MockController) ControllerName() string {
	return m.name
}

func (m *MockController) GetRouter() string {
	return "/mock [GET]"
}

func (m *MockController) Handle(ctx *gin.Context) {
	// Mock 实现
}

// MockMiddleware Mock 中间件
type MockMiddleware struct {
	name    string
	Service common.BaseService `inject:""`
}

func (m *MockMiddleware) MiddlewareName() string {
	return m.name
}

func (m *MockMiddleware) Order() int {
	return 0
}

func (m *MockMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func (m *MockMiddleware) OnStart() error {
	return nil
}

func (m *MockMiddleware) OnStop() error {
	return nil
}
