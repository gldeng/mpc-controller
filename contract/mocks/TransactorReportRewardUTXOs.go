// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/ethereum/go-ethereum/core/types"
)

// TransactorReportRewardUTXOs is an autogenerated mock type for the TransactorReportRewardUTXOs type
type TransactorReportRewardUTXOs struct {
	mock.Mock
}

type TransactorReportRewardUTXOs_Expecter struct {
	mock *mock.Mock
}

func (_m *TransactorReportRewardUTXOs) EXPECT() *TransactorReportRewardUTXOs_Expecter {
	return &TransactorReportRewardUTXOs_Expecter{mock: &_m.Mock}
}

// ReportRewardUTXOs provides a mock function with given fields: ctx, AddDelegatorTxID, RewardUTXOIDs
func (_m *TransactorReportRewardUTXOs) ReportRewardUTXOs(ctx context.Context, AddDelegatorTxID [32]byte, RewardUTXOIDs [][32]byte) (*types.Transaction, error) {
	ret := _m.Called(ctx, AddDelegatorTxID, RewardUTXOIDs)

	var r0 *types.Transaction
	if rf, ok := ret.Get(0).(func(context.Context, [32]byte, [][32]byte) *types.Transaction); ok {
		r0 = rf(ctx, AddDelegatorTxID, RewardUTXOIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Transaction)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, [32]byte, [][32]byte) error); ok {
		r1 = rf(ctx, AddDelegatorTxID, RewardUTXOIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TransactorReportRewardUTXOs_ReportRewardUTXOs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReportRewardUTXOs'
type TransactorReportRewardUTXOs_ReportRewardUTXOs_Call struct {
	*mock.Call
}

// ReportRewardUTXOs is a helper method to define mock.On call
//  - ctx context.Context
//  - AddDelegatorTxID [32]byte
//  - RewardUTXOIDs [][32]byte
func (_e *TransactorReportRewardUTXOs_Expecter) ReportRewardUTXOs(ctx interface{}, AddDelegatorTxID interface{}, RewardUTXOIDs interface{}) *TransactorReportRewardUTXOs_ReportRewardUTXOs_Call {
	return &TransactorReportRewardUTXOs_ReportRewardUTXOs_Call{Call: _e.mock.On("ReportRewardUTXOs", ctx, AddDelegatorTxID, RewardUTXOIDs)}
}

func (_c *TransactorReportRewardUTXOs_ReportRewardUTXOs_Call) Run(run func(ctx context.Context, AddDelegatorTxID [32]byte, RewardUTXOIDs [][32]byte)) *TransactorReportRewardUTXOs_ReportRewardUTXOs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([32]byte), args[2].([][32]byte))
	})
	return _c
}

func (_c *TransactorReportRewardUTXOs_ReportRewardUTXOs_Call) Return(_a0 *types.Transaction, _a1 error) *TransactorReportRewardUTXOs_ReportRewardUTXOs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewTransactorReportRewardUTXOs interface {
	mock.TestingT
	Cleanup(func())
}

// NewTransactorReportRewardUTXOs creates a new instance of TransactorReportRewardUTXOs. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTransactorReportRewardUTXOs(t mockConstructorTestingTNewTransactorReportRewardUTXOs) *TransactorReportRewardUTXOs {
	mock := &TransactorReportRewardUTXOs{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
