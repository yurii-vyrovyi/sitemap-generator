// Code generated by MockGen. DO NOT EDIT.
// Source: core.go

// Package core is a generated GoMock package.
package core

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockPageLoader is a mock of PageLoader interface.
type MockPageLoader struct {
	ctrl     *gomock.Controller
	recorder *MockPageLoaderMockRecorder
}

// MockPageLoaderMockRecorder is the mock recorder for MockPageLoader.
type MockPageLoaderMockRecorder struct {
	mock *MockPageLoader
}

// NewMockPageLoader creates a new mock instance.
func NewMockPageLoader(ctrl *gomock.Controller) *MockPageLoader {
	mock := &MockPageLoader{ctrl: ctrl}
	mock.recorder = &MockPageLoaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPageLoader) EXPECT() *MockPageLoaderMockRecorder {
	return m.recorder
}

// GetPageLinks mocks base method.
func (m *MockPageLoader) GetPageLinks(arg0 context.Context, arg1 string) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPageLinks", arg0, arg1)
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetPageLinks indicates an expected call of GetPageLinks.
func (mr *MockPageLoaderMockRecorder) GetPageLinks(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPageLinks", reflect.TypeOf((*MockPageLoader)(nil).GetPageLinks), arg0, arg1)
}

// MockReporter is a mock of Reporter interface.
type MockReporter struct {
	ctrl     *gomock.Controller
	recorder *MockReporterMockRecorder
}

// MockReporterMockRecorder is the mock recorder for MockReporter.
type MockReporterMockRecorder struct {
	mock *MockReporter
}

// NewMockReporter creates a new mock instance.
func NewMockReporter(ctrl *gomock.Controller) *MockReporter {
	mock := &MockReporter{ctrl: ctrl}
	mock.recorder = &MockReporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReporter) EXPECT() *MockReporterMockRecorder {
	return m.recorder
}

// Save mocks base method.
func (m *MockReporter) Save(root *PageItem) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", root)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockReporterMockRecorder) Save(root interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockReporter)(nil).Save), root)
}
