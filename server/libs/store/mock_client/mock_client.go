// Code generated by MockGen. DO NOT EDIT.
// Source: /home/zhuofeng/go/src/gitlab.yunshan.net/yunshan/droplet-libs/vendor/github.com/influxdata/influxdb/client/v2/client.go
// mockgen -source=/home/zhuofeng/go/src/gitlab.yunshan.net/yunshan/droplet-libs/vendor/github.com/influxdata/influxdb/client/v2/client.go > mock_client.go

// Package mock_client is a generated GoMock package.
package mock_client

import (
	gomock "github.com/golang/mock/gomock"
	v2 "github.com/influxdata/influxdb/client/v2"
	reflect "reflect"
	time "time"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// Ping mocks base method
func (m *MockClient) Ping(timeout time.Duration) (time.Duration, string, error) {
	ret := m.ctrl.Call(m, "Ping", timeout)
	ret0, _ := ret[0].(time.Duration)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Ping indicates an expected call of Ping
func (mr *MockClientMockRecorder) Ping(timeout interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockClient)(nil).Ping), timeout)
}

// Write mocks base method
func (m *MockClient) Write(bp v2.BatchPoints) error {
	ret := m.ctrl.Call(m, "Write", bp)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write mocks base method
func (m *MockClient) WriteDirect(db, rp string, data []byte) error {
	ret := m.ctrl.Call(m, "WriteDirect", db, rp, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write
func (mr *MockClientMockRecorder) Write(bp interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockClient)(nil).Write), bp)
}

// Query mocks base method
func (m *MockClient) Query(q v2.Query) (*v2.Response, error) {
	ret := m.ctrl.Call(m, "Query", q)
	ret0, _ := ret[0].(*v2.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockClientMockRecorder) Query(q interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockClient)(nil).Query), q)
}

// QueryAsChunk mocks base method
func (m *MockClient) QueryAsChunk(q v2.Query) (*v2.ChunkedResponse, error) {
	ret := m.ctrl.Call(m, "QueryAsChunk", q)
	ret0, _ := ret[0].(*v2.ChunkedResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryAsChunk indicates an expected call of QueryAsChunk
func (mr *MockClientMockRecorder) QueryAsChunk(q interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryAsChunk", reflect.TypeOf((*MockClient)(nil).QueryAsChunk), q)
}

// Close mocks base method
func (m *MockClient) Close() error {
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockClientMockRecorder) Close() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockClient)(nil).Close))
}

// MockBatchPoints is a mock of BatchPoints interface
type MockBatchPoints struct {
	ctrl     *gomock.Controller
	recorder *MockBatchPointsMockRecorder
}

// MockBatchPointsMockRecorder is the mock recorder for MockBatchPoints
type MockBatchPointsMockRecorder struct {
	mock *MockBatchPoints
}

// NewMockBatchPoints creates a new mock instance
func NewMockBatchPoints(ctrl *gomock.Controller) *MockBatchPoints {
	mock := &MockBatchPoints{ctrl: ctrl}
	mock.recorder = &MockBatchPointsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBatchPoints) EXPECT() *MockBatchPointsMockRecorder {
	return m.recorder
}

// AddPoint mocks base method
func (m *MockBatchPoints) AddPoint(p *v2.Point) {
	m.ctrl.Call(m, "AddPoint", p)
}

// AddPoint indicates an expected call of AddPoint
func (mr *MockBatchPointsMockRecorder) AddPoint(p interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPoint", reflect.TypeOf((*MockBatchPoints)(nil).AddPoint), p)
}

// AddPoints mocks base method
func (m *MockBatchPoints) AddPoints(ps []*v2.Point) {
	m.ctrl.Call(m, "AddPoints", ps)
}

// AddPoints indicates an expected call of AddPoints
func (mr *MockBatchPointsMockRecorder) AddPoints(ps interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPoints", reflect.TypeOf((*MockBatchPoints)(nil).AddPoints), ps)
}

// Points mocks base method
func (m *MockBatchPoints) Points() []*v2.Point {
	ret := m.ctrl.Call(m, "Points")
	ret0, _ := ret[0].([]*v2.Point)
	return ret0
}

// Points indicates an expected call of Points
func (mr *MockBatchPointsMockRecorder) Points() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Points", reflect.TypeOf((*MockBatchPoints)(nil).Points))
}

// Precision mocks base method
func (m *MockBatchPoints) Precision() string {
	ret := m.ctrl.Call(m, "Precision")
	ret0, _ := ret[0].(string)
	return ret0
}

// Precision indicates an expected call of Precision
func (mr *MockBatchPointsMockRecorder) Precision() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Precision", reflect.TypeOf((*MockBatchPoints)(nil).Precision))
}

// SetPrecision mocks base method
func (m *MockBatchPoints) SetPrecision(s string) error {
	ret := m.ctrl.Call(m, "SetPrecision", s)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetPrecision indicates an expected call of SetPrecision
func (mr *MockBatchPointsMockRecorder) SetPrecision(s interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPrecision", reflect.TypeOf((*MockBatchPoints)(nil).SetPrecision), s)
}

// Database mocks base method
func (m *MockBatchPoints) Database() string {
	ret := m.ctrl.Call(m, "Database")
	ret0, _ := ret[0].(string)
	return ret0
}

// Database indicates an expected call of Database
func (mr *MockBatchPointsMockRecorder) Database() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Database", reflect.TypeOf((*MockBatchPoints)(nil).Database))
}

// SetDatabase mocks base method
func (m *MockBatchPoints) SetDatabase(s string) {
	m.ctrl.Call(m, "SetDatabase", s)
}

// SetDatabase indicates an expected call of SetDatabase
func (mr *MockBatchPointsMockRecorder) SetDatabase(s interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDatabase", reflect.TypeOf((*MockBatchPoints)(nil).SetDatabase), s)
}

// WriteConsistency mocks base method
func (m *MockBatchPoints) WriteConsistency() string {
	ret := m.ctrl.Call(m, "WriteConsistency")
	ret0, _ := ret[0].(string)
	return ret0
}

// WriteConsistency indicates an expected call of WriteConsistency
func (mr *MockBatchPointsMockRecorder) WriteConsistency() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteConsistency", reflect.TypeOf((*MockBatchPoints)(nil).WriteConsistency))
}

// SetWriteConsistency mocks base method
func (m *MockBatchPoints) SetWriteConsistency(s string) {
	m.ctrl.Call(m, "SetWriteConsistency", s)
}

// SetWriteConsistency indicates an expected call of SetWriteConsistency
func (mr *MockBatchPointsMockRecorder) SetWriteConsistency(s interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWriteConsistency", reflect.TypeOf((*MockBatchPoints)(nil).SetWriteConsistency), s)
}

// RetentionPolicy mocks base method
func (m *MockBatchPoints) RetentionPolicy() string {
	ret := m.ctrl.Call(m, "RetentionPolicy")
	ret0, _ := ret[0].(string)
	return ret0
}

// RetentionPolicy indicates an expected call of RetentionPolicy
func (mr *MockBatchPointsMockRecorder) RetentionPolicy() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetentionPolicy", reflect.TypeOf((*MockBatchPoints)(nil).RetentionPolicy))
}

// SetRetentionPolicy mocks base method
func (m *MockBatchPoints) SetRetentionPolicy(s string) {
	m.ctrl.Call(m, "SetRetentionPolicy", s)
}

// SetRetentionPolicy indicates an expected call of SetRetentionPolicy
func (mr *MockBatchPointsMockRecorder) SetRetentionPolicy(s interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRetentionPolicy", reflect.TypeOf((*MockBatchPoints)(nil).SetRetentionPolicy), s)
}
