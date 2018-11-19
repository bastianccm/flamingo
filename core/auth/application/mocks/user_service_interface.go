// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import domain "flamingo.me/flamingo/core/auth/domain"
import mock "github.com/stretchr/testify/mock"
import sessions "github.com/gorilla/sessions"

// UserServiceInterface is an autogenerated mock type for the UserServiceInterface type
type UserServiceInterface struct {
	mock.Mock
}

// GetUser provides a mock function with given fields: ctx, session
func (_m *UserServiceInterface) GetUser(ctx context.Context, session *sessions.Session) *domain.User {
	ret := _m.Called(ctx, session)

	var r0 *domain.User
	if rf, ok := ret.Get(0).(func(context.Context, *sessions.Session) *domain.User); ok {
		r0 = rf(ctx, session)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}

	return r0
}

// InitUser provides a mock function with given fields: ctx, session
func (_m *UserServiceInterface) InitUser(ctx context.Context, session *sessions.Session) error {
	ret := _m.Called(ctx, session)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *sessions.Session) error); ok {
		r0 = rf(ctx, session)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsLoggedIn provides a mock function with given fields: ctx, session
func (_m *UserServiceInterface) IsLoggedIn(ctx context.Context, session *sessions.Session) bool {
	ret := _m.Called(ctx, session)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, *sessions.Session) bool); ok {
		r0 = rf(ctx, session)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}