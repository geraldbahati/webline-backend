package services

import (
	context "context"
	reflect "reflect"

	redis "github.com/go-redis/redis/v8"
	gomock "github.com/golang/mock/gomock"
)

// MockCacheService is a mock of CacheService interface.
type MockCacheService struct {
	ctrl     *gomock.Controller
	recorder *MockCacheServiceMockRecorder
}

// MockCacheServiceMockRecorder is the mock recorder for MockCacheService.
type MockCacheServiceMockRecorder struct {
	mock *MockCacheService
}

// NewMockCacheService creates a new mock instance.
func NewMockCacheService(ctrl *gomock.Controller) *MockCacheService {
	mock := &MockCacheService{ctrl: ctrl}
	mock.recorder = &MockCacheServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheService) EXPECT() *MockCacheServiceMockRecorder {
	return m.recorder
}

// Decr mocks base method.
func (m *MockCacheService) Decr(ctx context.Context, key string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Decr", ctx, key)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Decr indicates an expected call of Decr.
func (mr *MockCacheServiceMockRecorder) Decr(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Decr", reflect.TypeOf((*MockCacheService)(nil).Decr), ctx, key)
}

// Delete mocks base method.
func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, key)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockCacheServiceMockRecorder) Delete(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockCacheService)(nil).Delete), ctx, key)
}

// Get mocks base method.
func (m *MockCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, key, dest)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockCacheServiceMockRecorder) Get(ctx, key, dest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCacheService)(nil).Get), ctx, key, dest)
}

// HGet mocks base method.
func (m *MockCacheService) HGet(ctx context.Context, key, field string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HGet", ctx, key, field)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HGet indicates an expected call of HGet.
func (mr *MockCacheServiceMockRecorder) HGet(ctx, key, field interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HGet", reflect.TypeOf((*MockCacheService)(nil).HGet), ctx, key, field)
}

// HSet mocks base method.
func (m *MockCacheService) HSet(ctx context.Context, key, field string, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HSet", ctx, key, field, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// HSet indicates an expected call of HSet.
func (mr *MockCacheServiceMockRecorder) HSet(ctx, key, field, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HSet", reflect.TypeOf((*MockCacheService)(nil).HSet), ctx, key, field, value)
}

// Incr mocks base method.
func (m *MockCacheService) Incr(ctx context.Context, key string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Incr", ctx, key)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Incr indicates an expected call of Incr.
func (mr *MockCacheServiceMockRecorder) Incr(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Incr", reflect.TypeOf((*MockCacheService)(nil).Incr), ctx, key)
}

// Pipeline mocks base method.
func (m *MockCacheService) Pipeline(ctx context.Context, fn func(redis.Pipeliner) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pipeline", ctx, fn)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pipeline indicates an expected call of Pipeline.
func (mr *MockCacheServiceMockRecorder) Pipeline(ctx, fn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pipeline", reflect.TypeOf((*MockCacheService)(nil).Pipeline), ctx, fn)
}

// SAdd mocks base method.
func (m *MockCacheService) SAdd(ctx context.Context, key string, members ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, key}
	varargs = append(varargs, members...)
	ret := m.ctrl.Call(m, "SAdd", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// SAdd indicates an expected call of SAdd.
func (mr *MockCacheServiceMockRecorder) SAdd(ctx, key interface{}, members ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, key}, members...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SAdd", reflect.TypeOf((*MockCacheService)(nil).SAdd), varargs...)
}

// SMembers mocks base method.
func (m *MockCacheService) SMembers(ctx context.Context, key string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SMembers", ctx, key)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SMembers indicates an expected call of SMembers.
func (mr *MockCacheServiceMockRecorder) SMembers(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SMembers", reflect.TypeOf((*MockCacheService)(nil).SMembers), ctx, key)
}

// Set mocks base method.
func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, key, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockCacheServiceMockRecorder) Set(ctx, key, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCacheService)(nil).Set), ctx, key, value)
}

// ZAdd mocks base method.
func (m *MockCacheService) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, key}
	for _, a := range members {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ZAdd", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// ZAdd indicates an expected call of ZAdd.
func (mr *MockCacheServiceMockRecorder) ZAdd(ctx, key interface{}, members ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, key}, members...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ZAdd", reflect.TypeOf((*MockCacheService)(nil).ZAdd), varargs...)
}

// ZRange mocks base method.
func (m *MockCacheService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ZRange", ctx, key, start, stop)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ZRange indicates an expected call of ZRange.
func (mr *MockCacheServiceMockRecorder) ZRange(ctx, key, start, stop interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ZRange", reflect.TypeOf((*MockCacheService)(nil).ZRange), ctx, key, start, stop)
}