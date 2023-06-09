// Code generated by MockGen. DO NOT EDIT.
// Source: post.go

// Package post is a generated GoMock package.
package post

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockPostRepo is a mock of PostRepo interface.
type MockPostRepo struct {
	ctrl     *gomock.Controller
	recorder *MockPostRepoMockRecorder
}

// MockPostRepoMockRecorder is the mock recorder for MockPostRepo.
type MockPostRepoMockRecorder struct {
	mock *MockPostRepo
}

// NewMockPostRepo creates a new mock instance.
func NewMockPostRepo(ctrl *gomock.Controller) *MockPostRepo {
	mock := &MockPostRepo{ctrl: ctrl}
	mock.recorder = &MockPostRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPostRepo) EXPECT() *MockPostRepoMockRecorder {
	return m.recorder
}

// AddComment mocks base method.
func (m *MockPostRepo) AddComment(postID string, comment *Comment) (*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddComment", postID, comment)
	ret0, _ := ret[0].(*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddComment indicates an expected call of AddComment.
func (mr *MockPostRepoMockRecorder) AddComment(postID, comment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddComment", reflect.TypeOf((*MockPostRepo)(nil).AddComment), postID, comment)
}

// AddPost mocks base method.
func (m *MockPostRepo) AddPost(post *Post) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddPost", post)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddPost indicates an expected call of AddPost.
func (mr *MockPostRepoMockRecorder) AddPost(post interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPost", reflect.TypeOf((*MockPostRepo)(nil).AddPost), post)
}

// DeleteComment mocks base method.
func (m *MockPostRepo) DeleteComment(postID, commentID, userID string) (*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteComment", postID, commentID, userID)
	ret0, _ := ret[0].(*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteComment indicates an expected call of DeleteComment.
func (mr *MockPostRepoMockRecorder) DeleteComment(postID, commentID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteComment", reflect.TypeOf((*MockPostRepo)(nil).DeleteComment), postID, commentID, userID)
}

// DeletePost mocks base method.
func (m *MockPostRepo) DeletePost(postID, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePost", postID, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePost indicates an expected call of DeletePost.
func (mr *MockPostRepoMockRecorder) DeletePost(postID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePost", reflect.TypeOf((*MockPostRepo)(nil).DeletePost), postID, userID)
}

// DownvotePost mocks base method.
func (m *MockPostRepo) DownvotePost(postID, userID string) (*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownvotePost", postID, userID)
	ret0, _ := ret[0].(*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownvotePost indicates an expected call of DownvotePost.
func (mr *MockPostRepoMockRecorder) DownvotePost(postID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownvotePost", reflect.TypeOf((*MockPostRepo)(nil).DownvotePost), postID, userID)
}

// GetAll mocks base method.
func (m *MockPostRepo) GetAll() ([]*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll")
	ret0, _ := ret[0].([]*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockPostRepoMockRecorder) GetAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockPostRepo)(nil).GetAll))
}

// GetByCategory mocks base method.
func (m *MockPostRepo) GetByCategory(category string) ([]*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByCategory", category)
	ret0, _ := ret[0].([]*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByCategory indicates an expected call of GetByCategory.
func (mr *MockPostRepoMockRecorder) GetByCategory(category interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByCategory", reflect.TypeOf((*MockPostRepo)(nil).GetByCategory), category)
}

// GetByID mocks base method.
func (m *MockPostRepo) GetByID(postID string) (*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", postID)
	ret0, _ := ret[0].(*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockPostRepoMockRecorder) GetByID(postID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockPostRepo)(nil).GetByID), postID)
}

// GetByUser mocks base method.
func (m *MockPostRepo) GetByUser(username string) ([]*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByUser", username)
	ret0, _ := ret[0].([]*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByUser indicates an expected call of GetByUser.
func (mr *MockPostRepoMockRecorder) GetByUser(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUser", reflect.TypeOf((*MockPostRepo)(nil).GetByUser), username)
}

// UnvotePost mocks base method.
func (m *MockPostRepo) UnvotePost(postID, userID string) (*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnvotePost", postID, userID)
	ret0, _ := ret[0].(*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UnvotePost indicates an expected call of UnvotePost.
func (mr *MockPostRepoMockRecorder) UnvotePost(postID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnvotePost", reflect.TypeOf((*MockPostRepo)(nil).UnvotePost), postID, userID)
}

// UpvotePost mocks base method.
func (m *MockPostRepo) UpvotePost(postID, userID string) (*Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpvotePost", postID, userID)
	ret0, _ := ret[0].(*Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpvotePost indicates an expected call of UpvotePost.
func (mr *MockPostRepoMockRecorder) UpvotePost(postID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpvotePost", reflect.TypeOf((*MockPostRepo)(nil).UpvotePost), postID, userID)
}
