package container

import (
	"com.litelake.litecore/common"
	"github.com/gin-gonic/gin"
)

// IMockService Mock 服务接口
type IMockService interface {
	common.IBaseService
}

// IMockRepository Mock 仓储接口
type IMockRepository interface {
	common.IBaseRepository
}

// IMockController Mock 控制器接口
type IMockController interface {
	common.IBaseController
}

// IMockMiddleware Mock 中间件接口
type IMockMiddleware interface {
	common.IBaseMiddleware
}

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
	Config common.IBaseConfigProvider `inject:""`
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
	name    string
	Config  common.IBaseConfigProvider `inject:""`
	Manager common.IBaseManager        `inject:""`
	Entity  common.IBaseEntity         `inject:""`
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

var _ IMockRepository = (*MockRepository)(nil)

// MockService Mock 服务
type MockService struct {
	name   string
	Config common.IBaseConfigProvider `inject:""`
	Repo   common.IBaseRepository     `inject:""`
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

var _ IMockService = (*MockService)(nil)

// MockController Mock 控制器
type MockController struct {
	name    string
	Service common.IBaseService `inject:""`
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

var _ IMockController = (*MockController)(nil)

// MockMiddleware Mock 中间件
type MockMiddleware struct {
	name    string
	Service common.IBaseService `inject:""`
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

var _ IMockMiddleware = (*MockMiddleware)(nil)
