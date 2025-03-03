// Code generated by MockGen. DO NOT EDIT.
// Source: interfaces.go
//
// Generated by this command:
//
//	mockgen -package=redis -destination=mock.go -source=interfaces.go
//

// Package redis is a generated GoMock package.
package redis

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockIProducer is a mock of IProducer interface.
type MockIProducer[T any] struct {
	ctrl     *gomock.Controller
	recorder *MockIProducerMockRecorder[T]
	isgomock struct{}
}

// MockIProducerMockRecorder is the mock recorder for MockIProducer.
type MockIProducerMockRecorder[T any] struct {
	mock *MockIProducer[T]
}

// NewMockIProducer creates a new mock instance.
func NewMockIProducer[T any](ctrl *gomock.Controller) *MockIProducer[T] {
	mock := &MockIProducer[T]{ctrl: ctrl}
	mock.recorder = &MockIProducerMockRecorder[T]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIProducer[T]) EXPECT() *MockIProducerMockRecorder[T] {
	return m.recorder
}

// Close mocks base method.
func (m *MockIProducer[T]) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockIProducerMockRecorder[T]) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIProducer[T])(nil).Close))
}

// Publish mocks base method.
func (m *MockIProducer[T]) Publish(data T) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", data)
	ret0, _ := ret[0].(error)
	return ret0
}

// Publish indicates an expected call of Publish.
func (mr *MockIProducerMockRecorder[T]) Publish(data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockIProducer[T])(nil).Publish), data)
}

// Start mocks base method.
func (m *MockIProducer[T]) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockIProducerMockRecorder[T]) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockIProducer[T])(nil).Start))
}

// MockIGroupConsumer is a mock of IGroupConsumer interface.
type MockIGroupConsumer[T any] struct {
	ctrl     *gomock.Controller
	recorder *MockIGroupConsumerMockRecorder[T]
	isgomock struct{}
}

// MockIGroupConsumerMockRecorder is the mock recorder for MockIGroupConsumer.
type MockIGroupConsumerMockRecorder[T any] struct {
	mock *MockIGroupConsumer[T]
}

// NewMockIGroupConsumer creates a new mock instance.
func NewMockIGroupConsumer[T any](ctrl *gomock.Controller) *MockIGroupConsumer[T] {
	mock := &MockIGroupConsumer[T]{ctrl: ctrl}
	mock.recorder = &MockIGroupConsumerMockRecorder[T]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIGroupConsumer[T]) EXPECT() *MockIGroupConsumerMockRecorder[T] {
	return m.recorder
}

// Close mocks base method.
func (m *MockIGroupConsumer[T]) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockIGroupConsumerMockRecorder[T]) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIGroupConsumer[T])(nil).Close))
}

// Start mocks base method.
func (m *MockIGroupConsumer[T]) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockIGroupConsumerMockRecorder[T]) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockIGroupConsumer[T])(nil).Start))
}

// Subscribe mocks base method.
func (m *MockIGroupConsumer[T]) Subscribe() <-chan *Message[T] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe")
	ret0, _ := ret[0].(<-chan *Message[T])
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockIGroupConsumerMockRecorder[T]) Subscribe() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockIGroupConsumer[T])(nil).Subscribe))
}

// MockIConsumer is a mock of IConsumer interface.
type MockIConsumer[T any] struct {
	ctrl     *gomock.Controller
	recorder *MockIConsumerMockRecorder[T]
	isgomock struct{}
}

// MockIConsumerMockRecorder is the mock recorder for MockIConsumer.
type MockIConsumerMockRecorder[T any] struct {
	mock *MockIConsumer[T]
}

// NewMockIConsumer creates a new mock instance.
func NewMockIConsumer[T any](ctrl *gomock.Controller) *MockIConsumer[T] {
	mock := &MockIConsumer[T]{ctrl: ctrl}
	mock.recorder = &MockIConsumerMockRecorder[T]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIConsumer[T]) EXPECT() *MockIConsumerMockRecorder[T] {
	return m.recorder
}

// Close mocks base method.
func (m *MockIConsumer[T]) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockIConsumerMockRecorder[T]) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIConsumer[T])(nil).Close))
}

// Start mocks base method.
func (m *MockIConsumer[T]) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockIConsumerMockRecorder[T]) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockIConsumer[T])(nil).Start))
}

// Subscribe mocks base method.
func (m *MockIConsumer[T]) Subscribe() <-chan T {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe")
	ret0, _ := ret[0].(<-chan T)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockIConsumerMockRecorder[T]) Subscribe() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockIConsumer[T])(nil).Subscribe))
}

// MockIAutoRenewMutex is a mock of IAutoRenewMutex interface.
type MockIAutoRenewMutex struct {
	ctrl     *gomock.Controller
	recorder *MockIAutoRenewMutexMockRecorder
	isgomock struct{}
}

// MockIAutoRenewMutexMockRecorder is the mock recorder for MockIAutoRenewMutex.
type MockIAutoRenewMutexMockRecorder struct {
	mock *MockIAutoRenewMutex
}

// NewMockIAutoRenewMutex creates a new mock instance.
func NewMockIAutoRenewMutex(ctrl *gomock.Controller) *MockIAutoRenewMutex {
	mock := &MockIAutoRenewMutex{ctrl: ctrl}
	mock.recorder = &MockIAutoRenewMutexMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIAutoRenewMutex) EXPECT() *MockIAutoRenewMutexMockRecorder {
	return m.recorder
}

// Lock mocks base method.
func (m *MockIAutoRenewMutex) Lock(ctx context.Context) (context.Context, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Lock", ctx)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Lock indicates an expected call of Lock.
func (mr *MockIAutoRenewMutexMockRecorder) Lock(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockIAutoRenewMutex)(nil).Lock), ctx)
}

// Unlock mocks base method.
func (m *MockIAutoRenewMutex) Unlock() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unlock")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Unlock indicates an expected call of Unlock.
func (mr *MockIAutoRenewMutexMockRecorder) Unlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unlock", reflect.TypeOf((*MockIAutoRenewMutex)(nil).Unlock))
}

// Valid mocks base method.
func (m *MockIAutoRenewMutex) Valid() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Valid")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Valid indicates an expected call of Valid.
func (mr *MockIAutoRenewMutexMockRecorder) Valid() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Valid", reflect.TypeOf((*MockIAutoRenewMutex)(nil).Valid))
}
