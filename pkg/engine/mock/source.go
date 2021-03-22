// Code generated by MockGen. DO NOT EDIT.
// Source: ./source.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	source "github.com/Checkmarx/kics/pkg/engine/source"
	model "github.com/Checkmarx/kics/pkg/model"
	gomock "github.com/golang/mock/gomock"
)

// MockQueriesSource is a mock of QueriesSource interface.
type MockQueriesSource struct {
	ctrl     *gomock.Controller
	recorder *MockQueriesSourceMockRecorder
}

// MockQueriesSourceMockRecorder is the mock recorder for MockQueriesSource.
type MockQueriesSourceMockRecorder struct {
	mock *MockQueriesSource
}

// NewMockQueriesSource creates a new mock instance.
func NewMockQueriesSource(ctrl *gomock.Controller) *MockQueriesSource {
	mock := &MockQueriesSource{ctrl: ctrl}
	mock.recorder = &MockQueriesSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQueriesSource) EXPECT() *MockQueriesSourceMockRecorder {
	return m.recorder
}

// GetGenericQuery mocks base method.
func (m *MockQueriesSource) GetGenericQuery(platform string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGenericQuery", platform)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGenericQuery indicates an expected call of GetGenericQuery.
func (mr *MockQueriesSourceMockRecorder) GetGenericQuery(platform interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGenericQuery", reflect.TypeOf((*MockQueriesSource)(nil).GetGenericQuery), platform)
}

// GetQueries mocks base method.
func (m *MockQueriesSource) GetQueries(excludeQueries source.ExcludeQueries) ([]model.QueryMetadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetQueries", excludeQueries)
	ret0, _ := ret[0].([]model.QueryMetadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetQueries indicates an expected call of GetQueries.
func (mr *MockQueriesSourceMockRecorder) GetQueries(excludeQueries interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetQueries", reflect.TypeOf((*MockQueriesSource)(nil).GetQueries), excludeQueries)
}
