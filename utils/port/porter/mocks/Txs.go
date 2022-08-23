// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Txs is an autogenerated mock type for the Txs type
type Txs struct {
	mock.Mock
}

type Txs_Expecter struct {
	mock *mock.Mock
}

func (_m *Txs) EXPECT() *Txs_Expecter {
	return &Txs_Expecter{mock: &_m.Mock}
}

// ExportTxHash provides a mock function with given fields:
func (_m *Txs) ExportTxHash() ([]byte, error) {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Txs_ExportTxHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExportTxHash'
type Txs_ExportTxHash_Call struct {
	*mock.Call
}

// ExportTxHash is a helper method to define mock.On call
func (_e *Txs_Expecter) ExportTxHash() *Txs_ExportTxHash_Call {
	return &Txs_ExportTxHash_Call{Call: _e.mock.On("ExportTxHash")}
}

func (_c *Txs_ExportTxHash_Call) Run(run func()) *Txs_ExportTxHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txs_ExportTxHash_Call) Return(_a0 []byte, _a1 error) *Txs_ExportTxHash_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// ImportTxHash provides a mock function with given fields:
func (_m *Txs) ImportTxHash() ([]byte, error) {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Txs_ImportTxHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ImportTxHash'
type Txs_ImportTxHash_Call struct {
	*mock.Call
}

// ImportTxHash is a helper method to define mock.On call
func (_e *Txs_Expecter) ImportTxHash() *Txs_ImportTxHash_Call {
	return &Txs_ImportTxHash_Call{Call: _e.mock.On("ImportTxHash")}
}

func (_c *Txs_ImportTxHash_Call) Run(run func()) *Txs_ImportTxHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txs_ImportTxHash_Call) Return(_a0 []byte, _a1 error) *Txs_ImportTxHash_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// SetExportTxSig provides a mock function with given fields: exportTxSig
func (_m *Txs) SetExportTxSig(exportTxSig [65]byte) error {
	ret := _m.Called(exportTxSig)

	var r0 error
	if rf, ok := ret.Get(0).(func([65]byte) error); ok {
		r0 = rf(exportTxSig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Txs_SetExportTxSig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetExportTxSig'
type Txs_SetExportTxSig_Call struct {
	*mock.Call
}

// SetExportTxSig is a helper method to define mock.On call
//  - exportTxSig [65]byte
func (_e *Txs_Expecter) SetExportTxSig(exportTxSig interface{}) *Txs_SetExportTxSig_Call {
	return &Txs_SetExportTxSig_Call{Call: _e.mock.On("SetExportTxSig", exportTxSig)}
}

func (_c *Txs_SetExportTxSig_Call) Run(run func(exportTxSig [65]byte)) *Txs_SetExportTxSig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([65]byte))
	})
	return _c
}

func (_c *Txs_SetExportTxSig_Call) Return(_a0 error) *Txs_SetExportTxSig_Call {
	_c.Call.Return(_a0)
	return _c
}

// SetImportTxSig provides a mock function with given fields: importTxSig
func (_m *Txs) SetImportTxSig(importTxSig [65]byte) error {
	ret := _m.Called(importTxSig)

	var r0 error
	if rf, ok := ret.Get(0).(func([65]byte) error); ok {
		r0 = rf(importTxSig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Txs_SetImportTxSig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetImportTxSig'
type Txs_SetImportTxSig_Call struct {
	*mock.Call
}

// SetImportTxSig is a helper method to define mock.On call
//  - importTxSig [65]byte
func (_e *Txs_Expecter) SetImportTxSig(importTxSig interface{}) *Txs_SetImportTxSig_Call {
	return &Txs_SetImportTxSig_Call{Call: _e.mock.On("SetImportTxSig", importTxSig)}
}

func (_c *Txs_SetImportTxSig_Call) Run(run func(importTxSig [65]byte)) *Txs_SetImportTxSig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([65]byte))
	})
	return _c
}

func (_c *Txs_SetImportTxSig_Call) Return(_a0 error) *Txs_SetImportTxSig_Call {
	_c.Call.Return(_a0)
	return _c
}

// SignedExportTxBytes provides a mock function with given fields:
func (_m *Txs) SignedExportTxBytes() ([]byte, error) {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Txs_SignedExportTxBytes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SignedExportTxBytes'
type Txs_SignedExportTxBytes_Call struct {
	*mock.Call
}

// SignedExportTxBytes is a helper method to define mock.On call
func (_e *Txs_Expecter) SignedExportTxBytes() *Txs_SignedExportTxBytes_Call {
	return &Txs_SignedExportTxBytes_Call{Call: _e.mock.On("SignedExportTxBytes")}
}

func (_c *Txs_SignedExportTxBytes_Call) Run(run func()) *Txs_SignedExportTxBytes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txs_SignedExportTxBytes_Call) Return(_a0 []byte, _a1 error) *Txs_SignedExportTxBytes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// SignedImportTxBytes provides a mock function with given fields:
func (_m *Txs) SignedImportTxBytes() ([]byte, error) {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Txs_SignedImportTxBytes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SignedImportTxBytes'
type Txs_SignedImportTxBytes_Call struct {
	*mock.Call
}

// SignedImportTxBytes is a helper method to define mock.On call
func (_e *Txs_Expecter) SignedImportTxBytes() *Txs_SignedImportTxBytes_Call {
	return &Txs_SignedImportTxBytes_Call{Call: _e.mock.On("SignedImportTxBytes")}
}

func (_c *Txs_SignedImportTxBytes_Call) Run(run func()) *Txs_SignedImportTxBytes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txs_SignedImportTxBytes_Call) Return(_a0 []byte, _a1 error) *Txs_SignedImportTxBytes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewTxs interface {
	mock.TestingT
	Cleanup(func())
}

// NewTxs creates a new instance of Txs. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTxs(t mockConstructorTestingTNewTxs) *Txs {
	mock := &Txs{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
