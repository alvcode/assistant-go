// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	entity "assistant-go/internal/layer/entity"

	mock "github.com/stretchr/testify/mock"
)

// MockNoteRepository is an autogenerated mock type for the NoteRepository type
type MockNoteRepository struct {
	mock.Mock
}

type MockNoteRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockNoteRepository) EXPECT() *MockNoteRepository_Expecter {
	return &MockNoteRepository_Expecter{mock: &_m.Mock}
}

// CheckExistsByCategoryIDs provides a mock function with given fields: catIDs
func (_m *MockNoteRepository) CheckExistsByCategoryIDs(catIDs []int) (bool, error) {
	ret := _m.Called(catIDs)

	if len(ret) == 0 {
		panic("no return value specified for CheckExistsByCategoryIDs")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func([]int) (bool, error)); ok {
		return rf(catIDs)
	}
	if rf, ok := ret.Get(0).(func([]int) bool); ok {
		r0 = rf(catIDs)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func([]int) error); ok {
		r1 = rf(catIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockNoteRepository_CheckExistsByCategoryIDs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckExistsByCategoryIDs'
type MockNoteRepository_CheckExistsByCategoryIDs_Call struct {
	*mock.Call
}

// CheckExistsByCategoryIDs is a helper method to define mock.On call
//   - catIDs []int
func (_e *MockNoteRepository_Expecter) CheckExistsByCategoryIDs(catIDs interface{}) *MockNoteRepository_CheckExistsByCategoryIDs_Call {
	return &MockNoteRepository_CheckExistsByCategoryIDs_Call{Call: _e.mock.On("CheckExistsByCategoryIDs", catIDs)}
}

func (_c *MockNoteRepository_CheckExistsByCategoryIDs_Call) Run(run func(catIDs []int)) *MockNoteRepository_CheckExistsByCategoryIDs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]int))
	})
	return _c
}

func (_c *MockNoteRepository_CheckExistsByCategoryIDs_Call) Return(_a0 bool, _a1 error) *MockNoteRepository_CheckExistsByCategoryIDs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockNoteRepository_CheckExistsByCategoryIDs_Call) RunAndReturn(run func([]int) (bool, error)) *MockNoteRepository_CheckExistsByCategoryIDs_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: in
func (_m *MockNoteRepository) Create(in entity.Note) (*entity.Note, error) {
	ret := _m.Called(in)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *entity.Note
	var r1 error
	if rf, ok := ret.Get(0).(func(entity.Note) (*entity.Note, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func(entity.Note) *entity.Note); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Note)
		}
	}

	if rf, ok := ret.Get(1).(func(entity.Note) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockNoteRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockNoteRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - in entity.Note
func (_e *MockNoteRepository_Expecter) Create(in interface{}) *MockNoteRepository_Create_Call {
	return &MockNoteRepository_Create_Call{Call: _e.mock.On("Create", in)}
}

func (_c *MockNoteRepository_Create_Call) Run(run func(in entity.Note)) *MockNoteRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(entity.Note))
	})
	return _c
}

func (_c *MockNoteRepository_Create_Call) Return(_a0 *entity.Note, _a1 error) *MockNoteRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockNoteRepository_Create_Call) RunAndReturn(run func(entity.Note) (*entity.Note, error)) *MockNoteRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteOne provides a mock function with given fields: noteID
func (_m *MockNoteRepository) DeleteOne(noteID int) error {
	ret := _m.Called(noteID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteOne")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(noteID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockNoteRepository_DeleteOne_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteOne'
type MockNoteRepository_DeleteOne_Call struct {
	*mock.Call
}

// DeleteOne is a helper method to define mock.On call
//   - noteID int
func (_e *MockNoteRepository_Expecter) DeleteOne(noteID interface{}) *MockNoteRepository_DeleteOne_Call {
	return &MockNoteRepository_DeleteOne_Call{Call: _e.mock.On("DeleteOne", noteID)}
}

func (_c *MockNoteRepository_DeleteOne_Call) Run(run func(noteID int)) *MockNoteRepository_DeleteOne_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockNoteRepository_DeleteOne_Call) Return(_a0 error) *MockNoteRepository_DeleteOne_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockNoteRepository_DeleteOne_Call) RunAndReturn(run func(int) error) *MockNoteRepository_DeleteOne_Call {
	_c.Call.Return(run)
	return _c
}

// GetById provides a mock function with given fields: ID
func (_m *MockNoteRepository) GetById(ID int) (*entity.Note, error) {
	ret := _m.Called(ID)

	if len(ret) == 0 {
		panic("no return value specified for GetById")
	}

	var r0 *entity.Note
	var r1 error
	if rf, ok := ret.Get(0).(func(int) (*entity.Note, error)); ok {
		return rf(ID)
	}
	if rf, ok := ret.Get(0).(func(int) *entity.Note); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Note)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockNoteRepository_GetById_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetById'
type MockNoteRepository_GetById_Call struct {
	*mock.Call
}

// GetById is a helper method to define mock.On call
//   - ID int
func (_e *MockNoteRepository_Expecter) GetById(ID interface{}) *MockNoteRepository_GetById_Call {
	return &MockNoteRepository_GetById_Call{Call: _e.mock.On("GetById", ID)}
}

func (_c *MockNoteRepository_GetById_Call) Run(run func(ID int)) *MockNoteRepository_GetById_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockNoteRepository_GetById_Call) Return(_a0 *entity.Note, _a1 error) *MockNoteRepository_GetById_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockNoteRepository_GetById_Call) RunAndReturn(run func(int) (*entity.Note, error)) *MockNoteRepository_GetById_Call {
	_c.Call.Return(run)
	return _c
}

// GetMinimalByCategoryIds provides a mock function with given fields: catIds
func (_m *MockNoteRepository) GetMinimalByCategoryIds(catIds []int) ([]*entity.Note, error) {
	ret := _m.Called(catIds)

	if len(ret) == 0 {
		panic("no return value specified for GetMinimalByCategoryIds")
	}

	var r0 []*entity.Note
	var r1 error
	if rf, ok := ret.Get(0).(func([]int) ([]*entity.Note, error)); ok {
		return rf(catIds)
	}
	if rf, ok := ret.Get(0).(func([]int) []*entity.Note); ok {
		r0 = rf(catIds)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*entity.Note)
		}
	}

	if rf, ok := ret.Get(1).(func([]int) error); ok {
		r1 = rf(catIds)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockNoteRepository_GetMinimalByCategoryIds_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMinimalByCategoryIds'
type MockNoteRepository_GetMinimalByCategoryIds_Call struct {
	*mock.Call
}

// GetMinimalByCategoryIds is a helper method to define mock.On call
//   - catIds []int
func (_e *MockNoteRepository_Expecter) GetMinimalByCategoryIds(catIds interface{}) *MockNoteRepository_GetMinimalByCategoryIds_Call {
	return &MockNoteRepository_GetMinimalByCategoryIds_Call{Call: _e.mock.On("GetMinimalByCategoryIds", catIds)}
}

func (_c *MockNoteRepository_GetMinimalByCategoryIds_Call) Run(run func(catIds []int)) *MockNoteRepository_GetMinimalByCategoryIds_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]int))
	})
	return _c
}

func (_c *MockNoteRepository_GetMinimalByCategoryIds_Call) Return(_a0 []*entity.Note, _a1 error) *MockNoteRepository_GetMinimalByCategoryIds_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockNoteRepository_GetMinimalByCategoryIds_Call) RunAndReturn(run func([]int) ([]*entity.Note, error)) *MockNoteRepository_GetMinimalByCategoryIds_Call {
	_c.Call.Return(run)
	return _c
}

// Pin provides a mock function with given fields: noteID
func (_m *MockNoteRepository) Pin(noteID int) error {
	ret := _m.Called(noteID)

	if len(ret) == 0 {
		panic("no return value specified for Pin")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(noteID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockNoteRepository_Pin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Pin'
type MockNoteRepository_Pin_Call struct {
	*mock.Call
}

// Pin is a helper method to define mock.On call
//   - noteID int
func (_e *MockNoteRepository_Expecter) Pin(noteID interface{}) *MockNoteRepository_Pin_Call {
	return &MockNoteRepository_Pin_Call{Call: _e.mock.On("Pin", noteID)}
}

func (_c *MockNoteRepository_Pin_Call) Run(run func(noteID int)) *MockNoteRepository_Pin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockNoteRepository_Pin_Call) Return(_a0 error) *MockNoteRepository_Pin_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockNoteRepository_Pin_Call) RunAndReturn(run func(int) error) *MockNoteRepository_Pin_Call {
	_c.Call.Return(run)
	return _c
}

// UnPin provides a mock function with given fields: noteID
func (_m *MockNoteRepository) UnPin(noteID int) error {
	ret := _m.Called(noteID)

	if len(ret) == 0 {
		panic("no return value specified for UnPin")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(noteID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockNoteRepository_UnPin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UnPin'
type MockNoteRepository_UnPin_Call struct {
	*mock.Call
}

// UnPin is a helper method to define mock.On call
//   - noteID int
func (_e *MockNoteRepository_Expecter) UnPin(noteID interface{}) *MockNoteRepository_UnPin_Call {
	return &MockNoteRepository_UnPin_Call{Call: _e.mock.On("UnPin", noteID)}
}

func (_c *MockNoteRepository_UnPin_Call) Run(run func(noteID int)) *MockNoteRepository_UnPin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockNoteRepository_UnPin_Call) Return(_a0 error) *MockNoteRepository_UnPin_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockNoteRepository_UnPin_Call) RunAndReturn(run func(int) error) *MockNoteRepository_UnPin_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: in
func (_m *MockNoteRepository) Update(in *entity.Note) error {
	ret := _m.Called(in)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*entity.Note) error); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockNoteRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type MockNoteRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - in *entity.Note
func (_e *MockNoteRepository_Expecter) Update(in interface{}) *MockNoteRepository_Update_Call {
	return &MockNoteRepository_Update_Call{Call: _e.mock.On("Update", in)}
}

func (_c *MockNoteRepository_Update_Call) Run(run func(in *entity.Note)) *MockNoteRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*entity.Note))
	})
	return _c
}

func (_c *MockNoteRepository_Update_Call) Return(_a0 error) *MockNoteRepository_Update_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockNoteRepository_Update_Call) RunAndReturn(run func(*entity.Note) error) *MockNoteRepository_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockNoteRepository creates a new instance of MockNoteRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockNoteRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockNoteRepository {
	mock := &MockNoteRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
