// Code generated by MockGen. DO NOT EDIT.
// Source: ./space.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	model "webserver/internal/model"

	gomock "github.com/golang/mock/gomock"
)

// MockspaceRepo is a mock of spaceRepo interface.
type MockspaceRepo struct {
	ctrl     *gomock.Controller
	recorder *MockspaceRepoMockRecorder
}

// MockspaceRepoMockRecorder is the mock recorder for MockspaceRepo.
type MockspaceRepoMockRecorder struct {
	mock *MockspaceRepo
}

// NewMockspaceRepo creates a new mock instance.
func NewMockspaceRepo(ctrl *gomock.Controller) *MockspaceRepo {
	mock := &MockspaceRepo{ctrl: ctrl}
	mock.recorder = &MockspaceRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockspaceRepo) EXPECT() *MockspaceRepoMockRecorder {
	return m.recorder
}

// CreateNote mocks base method.
func (m *MockspaceRepo) CreateNote(ctx context.Context, note model.CreateNoteRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNote", ctx, note)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateNote indicates an expected call of CreateNote.
func (mr *MockspaceRepoMockRecorder) CreateNote(ctx, note interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNote", reflect.TypeOf((*MockspaceRepo)(nil).CreateNote), ctx, note)
}

// GetAllNotesBySpaceID mocks base method.
func (m *MockspaceRepo) GetAllNotesBySpaceID(ctx context.Context, spaceID int64) ([]model.GetNote, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllNotesBySpaceID", ctx, spaceID)
	ret0, _ := ret[0].([]model.GetNote)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllNotesBySpaceID indicates an expected call of GetAllNotesBySpaceID.
func (mr *MockspaceRepoMockRecorder) GetAllNotesBySpaceID(ctx, spaceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllNotesBySpaceID", reflect.TypeOf((*MockspaceRepo)(nil).GetAllNotesBySpaceID), ctx, spaceID)
}

// GetAllNotesBySpaceIDFull mocks base method.
func (m *MockspaceRepo) GetAllNotesBySpaceIDFull(ctx context.Context, spaceID int64) ([]model.Note, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllNotesBySpaceIDFull", ctx, spaceID)
	ret0, _ := ret[0].([]model.Note)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllNotesBySpaceIDFull indicates an expected call of GetAllNotesBySpaceIDFull.
func (mr *MockspaceRepoMockRecorder) GetAllNotesBySpaceIDFull(ctx, spaceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllNotesBySpaceIDFull", reflect.TypeOf((*MockspaceRepo)(nil).GetAllNotesBySpaceIDFull), ctx, spaceID)
}

// GetSpaceByID mocks base method.
func (m *MockspaceRepo) GetSpaceByID(ctx context.Context, id int) (model.Space, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSpaceByID", ctx, id)
	ret0, _ := ret[0].(model.Space)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSpaceByID indicates an expected call of GetSpaceByID.
func (mr *MockspaceRepoMockRecorder) GetSpaceByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSpaceByID", reflect.TypeOf((*MockspaceRepo)(nil).GetSpaceByID), ctx, id)
}

// UpdateNote mocks base method.
func (m *MockspaceRepo) UpdateNote(ctx context.Context, update model.UpdateNote) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNote", ctx, update)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNote indicates an expected call of UpdateNote.
func (mr *MockspaceRepoMockRecorder) UpdateNote(ctx, update interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNote", reflect.TypeOf((*MockspaceRepo)(nil).UpdateNote), ctx, update)
}
