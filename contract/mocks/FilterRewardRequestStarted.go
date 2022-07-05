// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	context "context"

	contract "github.com/avalido/mpc-controller/contract"
	mock "github.com/stretchr/testify/mock"
)

// FilterRewardRequestStarted is an autogenerated mock type for the FilterRewardRequestStarted type
type FilterRewardRequestStarted struct {
	mock.Mock
}

type FilterRewardRequestStarted_Expecter struct {
	mock *mock.Mock
}

func (_m *FilterRewardRequestStarted) EXPECT() *FilterRewardRequestStarted_Expecter {
	return &FilterRewardRequestStarted_Expecter{mock: &_m.Mock}
}

// WatchRewardRequestStarted provides a mock function with given fields: ctx, addDelegatorTxID
func (_m *FilterRewardRequestStarted) WatchRewardRequestStarted(ctx context.Context, addDelegatorTxID [][32]byte) (<-chan *contract.MpcManagerExportRewardRequestStarted, error) {
	ret := _m.Called(ctx, addDelegatorTxID)

	var r0 <-chan *contract.MpcManagerExportRewardRequestStarted
	if rf, ok := ret.Get(0).(func(context.Context, [][32]byte) <-chan *contract.MpcManagerExportRewardRequestStarted); ok {
		r0 = rf(ctx, addDelegatorTxID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *contract.MpcManagerExportRewardRequestStarted)
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

// FilterRewardRequestStarted_WatchRewardRequestStarted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WatchRewardRequestStarted'
type FilterRewardRequestStarted_WatchRewardRequestStarted_Call struct {
	*mock.Call
}

// WatchRewardRequestStarted is a helper method to define mock.On call
//  - ctx context.Context
//  - addDelegatorTxID [][32]byte
func (_e *FilterRewardRequestStarted_Expecter) WatchRewardRequestStarted(ctx interface{}, addDelegatorTxID interface{}) *FilterRewardRequestStarted_WatchRewardRequestStarted_Call {
	return &FilterRewardRequestStarted_WatchRewardRequestStarted_Call{Call: _e.mock.On("WatchRewardRequestStarted", ctx, addDelegatorTxID)}
}

func (_c *FilterRewardRequestStarted_WatchRewardRequestStarted_Call) Run(run func(ctx context.Context, addDelegatorTxID [][32]byte)) *FilterRewardRequestStarted_WatchRewardRequestStarted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([][32]byte))
	})
	return _c
}

func (_c *FilterRewardRequestStarted_WatchRewardRequestStarted_Call) Return(_a0 <-chan *contract.MpcManagerExportRewardRequestStarted, _a1 error) *FilterRewardRequestStarted_WatchRewardRequestStarted_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewFilterRewardRequestStarted interface {
	mock.TestingT
	Cleanup(func())
}

// NewFilterRewardRequestStarted creates a new instance of FilterRewardRequestStarted. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFilterRewardRequestStarted(t mockConstructorTestingTNewFilterRewardRequestStarted) *FilterRewardRequestStarted {
	mock := &FilterRewardRequestStarted{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
