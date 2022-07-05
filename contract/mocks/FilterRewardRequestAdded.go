// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	context "context"

	contract "github.com/avalido/mpc-controller/contract"
	mock "github.com/stretchr/testify/mock"
)

// FilterRewardRequestAdded is an autogenerated mock type for the FilterRewardRequestAdded type
type FilterRewardRequestAdded struct {
	mock.Mock
}

type FilterRewardRequestAdded_Expecter struct {
	mock *mock.Mock
}

func (_m *FilterRewardRequestAdded) EXPECT() *FilterRewardRequestAdded_Expecter {
	return &FilterRewardRequestAdded_Expecter{mock: &_m.Mock}
}

// WatchRewardRequestAdded provides a mock function with given fields: ctx, addDelegatorTxID
func (_m *FilterRewardRequestAdded) WatchRewardRequestAdded(ctx context.Context, addDelegatorTxID [][32]byte) (<-chan *contract.MpcManagerExportRewardRequestAdded, error) {
	ret := _m.Called(ctx, addDelegatorTxID)

	var r0 <-chan *contract.MpcManagerExportRewardRequestAdded
	if rf, ok := ret.Get(0).(func(context.Context, [][32]byte) <-chan *contract.MpcManagerExportRewardRequestAdded); ok {
		r0 = rf(ctx, addDelegatorTxID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *contract.MpcManagerExportRewardRequestAdded)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, [][32]byte) error); ok {
		r1 = rf(ctx, addDelegatorTxID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FilterRewardRequestAdded_WatchRewardRequestAdded_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WatchRewardRequestAdded'
type FilterRewardRequestAdded_WatchRewardRequestAdded_Call struct {
	*mock.Call
}

// WatchRewardRequestAdded is a helper method to define mock.On call
//  - ctx context.Context
//  - addDelegatorTxID [][32]byte
func (_e *FilterRewardRequestAdded_Expecter) WatchRewardRequestAdded(ctx interface{}, addDelegatorTxID interface{}) *FilterRewardRequestAdded_WatchRewardRequestAdded_Call {
	return &FilterRewardRequestAdded_WatchRewardRequestAdded_Call{Call: _e.mock.On("WatchRewardRequestAdded", ctx, addDelegatorTxID)}
}

func (_c *FilterRewardRequestAdded_WatchRewardRequestAdded_Call) Run(run func(ctx context.Context, addDelegatorTxID [][32]byte)) *FilterRewardRequestAdded_WatchRewardRequestAdded_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([][32]byte))
	})
	return _c
}

func (_c *FilterRewardRequestAdded_WatchRewardRequestAdded_Call) Return(_a0 <-chan *contract.MpcManagerExportRewardRequestAdded, _a1 error) *FilterRewardRequestAdded_WatchRewardRequestAdded_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewFilterRewardRequestAdded interface {
	mock.TestingT
	Cleanup(func())
}

// NewFilterRewardRequestAdded creates a new instance of FilterRewardRequestAdded. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFilterRewardRequestAdded(t mockConstructorTestingTNewFilterRewardRequestAdded) *FilterRewardRequestAdded {
	mock := &FilterRewardRequestAdded{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
