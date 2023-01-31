// Code generated by MockGen. DO NOT EDIT.
// Source: store.go

// Package store is a generated GoMock package.
package store

import (
	context "context"
	reflect "reflect"

	models "github.com/faroshq/faros-ingress/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore.
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance.
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStore) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStoreMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStore)(nil).Close))
}

// CreateConnection mocks base method.
func (m *MockStore) CreateConnection(arg0 context.Context, arg1 models.Connection) (*models.Connection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateConnection", arg0, arg1)
	ret0, _ := ret[0].(*models.Connection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateConnection indicates an expected call of CreateConnection.
func (mr *MockStoreMockRecorder) CreateConnection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateConnection", reflect.TypeOf((*MockStore)(nil).CreateConnection), arg0, arg1)
}

// CreateUser mocks base method.
func (m *MockStore) CreateUser(arg0 context.Context, arg1 models.User) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0, arg1)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockStoreMockRecorder) CreateUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockStore)(nil).CreateUser), arg0, arg1)
}

// DeleteConnection mocks base method.
func (m *MockStore) DeleteConnection(arg0 context.Context, arg1 models.Connection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteConnection", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteConnection indicates an expected call of DeleteConnection.
func (mr *MockStoreMockRecorder) DeleteConnection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteConnection", reflect.TypeOf((*MockStore)(nil).DeleteConnection), arg0, arg1)
}

// DeleteUser mocks base method.
func (m *MockStore) DeleteUser(arg0 context.Context, arg1 models.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockStoreMockRecorder) DeleteUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockStore)(nil).DeleteUser), arg0, arg1)
}

// GetConnection mocks base method.
func (m *MockStore) GetConnection(arg0 context.Context, arg1 models.Connection) (*models.Connection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConnection", arg0, arg1)
	ret0, _ := ret[0].(*models.Connection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConnection indicates an expected call of GetConnection.
func (mr *MockStoreMockRecorder) GetConnection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConnection", reflect.TypeOf((*MockStore)(nil).GetConnection), arg0, arg1)
}

// GetUser mocks base method.
func (m *MockStore) GetUser(arg0 context.Context, arg1 models.User) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0, arg1)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockStoreMockRecorder) GetUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockStore)(nil).GetUser), arg0, arg1)
}

// ListAllConnections mocks base method.
func (m *MockStore) ListAllConnections(ctx context.Context) ([]models.Connection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAllConnections", ctx)
	ret0, _ := ret[0].([]models.Connection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAllConnections indicates an expected call of ListAllConnections.
func (mr *MockStoreMockRecorder) ListAllConnections(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAllConnections", reflect.TypeOf((*MockStore)(nil).ListAllConnections), ctx)
}

// ListConnections mocks base method.
func (m *MockStore) ListConnections(arg0 context.Context, arg1 models.Connection) ([]models.Connection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListConnections", arg0, arg1)
	ret0, _ := ret[0].([]models.Connection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListConnections indicates an expected call of ListConnections.
func (mr *MockStoreMockRecorder) ListConnections(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListConnections", reflect.TypeOf((*MockStore)(nil).ListConnections), arg0, arg1)
}

// ListUsers mocks base method.
func (m *MockStore) ListUsers(arg0 context.Context, arg1 models.User) ([]models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsers", arg0, arg1)
	ret0, _ := ret[0].([]models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsers indicates an expected call of ListUsers.
func (mr *MockStoreMockRecorder) ListUsers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsers", reflect.TypeOf((*MockStore)(nil).ListUsers), arg0, arg1)
}

// RawDB mocks base method.
func (m *MockStore) RawDB() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RawDB")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// RawDB indicates an expected call of RawDB.
func (mr *MockStoreMockRecorder) RawDB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RawDB", reflect.TypeOf((*MockStore)(nil).RawDB))
}

// Status mocks base method.
func (m *MockStore) Status() (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Status")
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Status indicates an expected call of Status.
func (mr *MockStoreMockRecorder) Status() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Status", reflect.TypeOf((*MockStore)(nil).Status))
}

// SubscribeChanges mocks base method.
func (m *MockStore) SubscribeChanges(ctx context.Context, callback func(*models.Event) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribeChanges", ctx, callback)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubscribeChanges indicates an expected call of SubscribeChanges.
func (mr *MockStoreMockRecorder) SubscribeChanges(ctx, callback interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeChanges", reflect.TypeOf((*MockStore)(nil).SubscribeChanges), ctx, callback)
}

// UpdateConnection mocks base method.
func (m *MockStore) UpdateConnection(arg0 context.Context, arg1 models.Connection) (*models.Connection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateConnection", arg0, arg1)
	ret0, _ := ret[0].(*models.Connection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateConnection indicates an expected call of UpdateConnection.
func (mr *MockStoreMockRecorder) UpdateConnection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateConnection", reflect.TypeOf((*MockStore)(nil).UpdateConnection), arg0, arg1)
}

// UpdateConnectionLastSeen mocks base method.
func (m *MockStore) UpdateConnectionLastSeen(arg0 context.Context, arg1 models.Connection, arg2 models.ConnectionState) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateConnectionLastSeen", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateConnectionLastSeen indicates an expected call of UpdateConnectionLastSeen.
func (mr *MockStoreMockRecorder) UpdateConnectionLastSeen(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateConnectionLastSeen", reflect.TypeOf((*MockStore)(nil).UpdateConnectionLastSeen), arg0, arg1, arg2)
}

// UpdateUser mocks base method.
func (m *MockStore) UpdateUser(arg0 context.Context, arg1 models.User) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", arg0, arg1)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockStoreMockRecorder) UpdateUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockStore)(nil).UpdateUser), arg0, arg1)
}
