package services

type ITestService interface {
	Execute() error
}

type TestServiceImpl struct{}

func NewTestService() ITestService {
	return &TestServiceImpl{}
}
