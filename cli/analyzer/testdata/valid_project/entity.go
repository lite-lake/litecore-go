package main

type ITestEntity interface {
	GetID() string
}

type TestEntity struct {
	ID string
}
