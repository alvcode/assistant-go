// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	entity "assistant-go/internal/layer/entity"

	mock "github.com/stretchr/testify/mock"
)

// MockUserRepository is an autogenerated mock type for the UserRepository type
type MockUserRepository struct {
	mock.Mock
}

type MockUserRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserRepository) EXPECT() *MockUserRepository_Expecter {
	return &MockUserRepository_Expecter{mock: &_m.Mock}
}

// ChangePassword provides a mock function with given fields: userID, newPassword
func (_m *MockUserRepository) ChangePassword(userID int, newPassword string) error {
	ret := _m.Called(userID, newPassword)

	if len(ret) == 0 {
		panic("no return value specified for ChangePassword")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int, string) error); ok {
		r0 = rf(userID, newPassword)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockUserRepository_ChangePassword_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChangePassword'
type MockUserRepository_ChangePassword_Call struct {
	*mock.Call
}

// ChangePassword is a helper method to define mock.On call
//   - userID int
//   - newPassword string
func (_e *MockUserRepository_Expecter) ChangePassword(userID interface{}, newPassword interface{}) *MockUserRepository_ChangePassword_Call {
	return &MockUserRepository_ChangePassword_Call{Call: _e.mock.On("ChangePassword", userID, newPassword)}
}

func (_c *MockUserRepository_ChangePassword_Call) Run(run func(userID int, newPassword string)) *MockUserRepository_ChangePassword_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int), args[1].(string))
	})
	return _c
}

func (_c *MockUserRepository_ChangePassword_Call) Return(_a0 error) *MockUserRepository_ChangePassword_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockUserRepository_ChangePassword_Call) RunAndReturn(run func(int, string) error) *MockUserRepository_ChangePassword_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: in
func (_m *MockUserRepository) Create(in entity.User) (*entity.User, error) {
	ret := _m.Called(in)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *entity.User
	var r1 error
	if rf, ok := ret.Get(0).(func(entity.User) (*entity.User, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func(entity.User) *entity.User); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.User)
		}
	}

	if rf, ok := ret.Get(1).(func(entity.User) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockUserRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - in entity.User
func (_e *MockUserRepository_Expecter) Create(in interface{}) *MockUserRepository_Create_Call {
	return &MockUserRepository_Create_Call{Call: _e.mock.On("Create", in)}
}

func (_c *MockUserRepository_Create_Call) Run(run func(in entity.User)) *MockUserRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(entity.User))
	})
	return _c
}

func (_c *MockUserRepository_Create_Call) Return(_a0 *entity.User, _a1 error) *MockUserRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserRepository_Create_Call) RunAndReturn(run func(entity.User) (*entity.User, error)) *MockUserRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: userID
func (_m *MockUserRepository) Delete(userID int) error {
	ret := _m.Called(userID)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockUserRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockUserRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - userID int
func (_e *MockUserRepository_Expecter) Delete(userID interface{}) *MockUserRepository_Delete_Call {
	return &MockUserRepository_Delete_Call{Call: _e.mock.On("Delete", userID)}
}

func (_c *MockUserRepository_Delete_Call) Run(run func(userID int)) *MockUserRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockUserRepository_Delete_Call) Return(_a0 error) *MockUserRepository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockUserRepository_Delete_Call) RunAndReturn(run func(int) error) *MockUserRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteUserTokensByID provides a mock function with given fields: userID
func (_m *MockUserRepository) DeleteUserTokensByID(userID int) error {
	ret := _m.Called(userID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUserTokensByID")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockUserRepository_DeleteUserTokensByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteUserTokensByID'
type MockUserRepository_DeleteUserTokensByID_Call struct {
	*mock.Call
}

// DeleteUserTokensByID is a helper method to define mock.On call
//   - userID int
func (_e *MockUserRepository_Expecter) DeleteUserTokensByID(userID interface{}) *MockUserRepository_DeleteUserTokensByID_Call {
	return &MockUserRepository_DeleteUserTokensByID_Call{Call: _e.mock.On("DeleteUserTokensByID", userID)}
}

func (_c *MockUserRepository_DeleteUserTokensByID_Call) Run(run func(userID int)) *MockUserRepository_DeleteUserTokensByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockUserRepository_DeleteUserTokensByID_Call) Return(_a0 error) *MockUserRepository_DeleteUserTokensByID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockUserRepository_DeleteUserTokensByID_Call) RunAndReturn(run func(int) error) *MockUserRepository_DeleteUserTokensByID_Call {
	_c.Call.Return(run)
	return _c
}

// Find provides a mock function with given fields: login
func (_m *MockUserRepository) Find(login string) (*entity.User, error) {
	ret := _m.Called(login)

	if len(ret) == 0 {
		panic("no return value specified for Find")
	}

	var r0 *entity.User
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*entity.User, error)); ok {
		return rf(login)
	}
	if rf, ok := ret.Get(0).(func(string) *entity.User); ok {
		r0 = rf(login)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.User)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserRepository_Find_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Find'
type MockUserRepository_Find_Call struct {
	*mock.Call
}

// Find is a helper method to define mock.On call
//   - login string
func (_e *MockUserRepository_Expecter) Find(login interface{}) *MockUserRepository_Find_Call {
	return &MockUserRepository_Find_Call{Call: _e.mock.On("Find", login)}
}

func (_c *MockUserRepository_Find_Call) Run(run func(login string)) *MockUserRepository_Find_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockUserRepository_Find_Call) Return(_a0 *entity.User, _a1 error) *MockUserRepository_Find_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserRepository_Find_Call) RunAndReturn(run func(string) (*entity.User, error)) *MockUserRepository_Find_Call {
	_c.Call.Return(run)
	return _c
}

// FindById provides a mock function with given fields: id
func (_m *MockUserRepository) FindById(id int) (*entity.User, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for FindById")
	}

	var r0 *entity.User
	var r1 error
	if rf, ok := ret.Get(0).(func(int) (*entity.User, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(int) *entity.User); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.User)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserRepository_FindById_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindById'
type MockUserRepository_FindById_Call struct {
	*mock.Call
}

// FindById is a helper method to define mock.On call
//   - id int
func (_e *MockUserRepository_Expecter) FindById(id interface{}) *MockUserRepository_FindById_Call {
	return &MockUserRepository_FindById_Call{Call: _e.mock.On("FindById", id)}
}

func (_c *MockUserRepository_FindById_Call) Run(run func(id int)) *MockUserRepository_FindById_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockUserRepository_FindById_Call) Return(_a0 *entity.User, _a1 error) *MockUserRepository_FindById_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserRepository_FindById_Call) RunAndReturn(run func(int) (*entity.User, error)) *MockUserRepository_FindById_Call {
	_c.Call.Return(run)
	return _c
}

// FindUserToken provides a mock function with given fields: token
func (_m *MockUserRepository) FindUserToken(token string) (*entity.UserToken, error) {
	ret := _m.Called(token)

	if len(ret) == 0 {
		panic("no return value specified for FindUserToken")
	}

	var r0 *entity.UserToken
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*entity.UserToken, error)); ok {
		return rf(token)
	}
	if rf, ok := ret.Get(0).(func(string) *entity.UserToken); ok {
		r0 = rf(token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.UserToken)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserRepository_FindUserToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindUserToken'
type MockUserRepository_FindUserToken_Call struct {
	*mock.Call
}

// FindUserToken is a helper method to define mock.On call
//   - token string
func (_e *MockUserRepository_Expecter) FindUserToken(token interface{}) *MockUserRepository_FindUserToken_Call {
	return &MockUserRepository_FindUserToken_Call{Call: _e.mock.On("FindUserToken", token)}
}

func (_c *MockUserRepository_FindUserToken_Call) Run(run func(token string)) *MockUserRepository_FindUserToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockUserRepository_FindUserToken_Call) Return(_a0 *entity.UserToken, _a1 error) *MockUserRepository_FindUserToken_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserRepository_FindUserToken_Call) RunAndReturn(run func(string) (*entity.UserToken, error)) *MockUserRepository_FindUserToken_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveTokensByDateExpired provides a mock function with given fields: time
func (_m *MockUserRepository) RemoveTokensByDateExpired(time int) error {
	ret := _m.Called(time)

	if len(ret) == 0 {
		panic("no return value specified for RemoveTokensByDateExpired")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(time)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockUserRepository_RemoveTokensByDateExpired_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveTokensByDateExpired'
type MockUserRepository_RemoveTokensByDateExpired_Call struct {
	*mock.Call
}

// RemoveTokensByDateExpired is a helper method to define mock.On call
//   - time int
func (_e *MockUserRepository_Expecter) RemoveTokensByDateExpired(time interface{}) *MockUserRepository_RemoveTokensByDateExpired_Call {
	return &MockUserRepository_RemoveTokensByDateExpired_Call{Call: _e.mock.On("RemoveTokensByDateExpired", time)}
}

func (_c *MockUserRepository_RemoveTokensByDateExpired_Call) Run(run func(time int)) *MockUserRepository_RemoveTokensByDateExpired_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockUserRepository_RemoveTokensByDateExpired_Call) Return(_a0 error) *MockUserRepository_RemoveTokensByDateExpired_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockUserRepository_RemoveTokensByDateExpired_Call) RunAndReturn(run func(int) error) *MockUserRepository_RemoveTokensByDateExpired_Call {
	_c.Call.Return(run)
	return _c
}

// SetUserToken provides a mock function with given fields: in
func (_m *MockUserRepository) SetUserToken(in entity.UserToken) (*entity.UserToken, error) {
	ret := _m.Called(in)

	if len(ret) == 0 {
		panic("no return value specified for SetUserToken")
	}

	var r0 *entity.UserToken
	var r1 error
	if rf, ok := ret.Get(0).(func(entity.UserToken) (*entity.UserToken, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func(entity.UserToken) *entity.UserToken); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.UserToken)
		}
	}

	if rf, ok := ret.Get(1).(func(entity.UserToken) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserRepository_SetUserToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetUserToken'
type MockUserRepository_SetUserToken_Call struct {
	*mock.Call
}

// SetUserToken is a helper method to define mock.On call
//   - in entity.UserToken
func (_e *MockUserRepository_Expecter) SetUserToken(in interface{}) *MockUserRepository_SetUserToken_Call {
	return &MockUserRepository_SetUserToken_Call{Call: _e.mock.On("SetUserToken", in)}
}

func (_c *MockUserRepository_SetUserToken_Call) Run(run func(in entity.UserToken)) *MockUserRepository_SetUserToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(entity.UserToken))
	})
	return _c
}

func (_c *MockUserRepository_SetUserToken_Call) Return(_a0 *entity.UserToken, _a1 error) *MockUserRepository_SetUserToken_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserRepository_SetUserToken_Call) RunAndReturn(run func(entity.UserToken) (*entity.UserToken, error)) *MockUserRepository_SetUserToken_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockUserRepository creates a new instance of MockUserRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserRepository {
	mock := &MockUserRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
