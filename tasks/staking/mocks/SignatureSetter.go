// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// SignatureSetter is an autogenerated mock type for the SignatureSetter type
type SignatureSetter struct {
	mock.Mock
}

// SetAddDelegatorTxSig provides a mock function with given fields: sig
func (_m *SignatureSetter) SetAddDelegatorTxSig(sig [65]byte) error {
	ret := _m.Called(sig)

	var r0 error
	if rf, ok := ret.Get(0).(func([65]byte) error); ok {
		r0 = rf(sig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetExportTxSig provides a mock function with given fields: sig
func (_m *SignatureSetter) SetExportTxSig(sig [65]byte) error {
	ret := _m.Called(sig)

	var r0 error
	if rf, ok := ret.Get(0).(func([65]byte) error); ok {
		r0 = rf(sig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetImportTxSig provides a mock function with given fields: sig
func (_m *SignatureSetter) SetImportTxSig(sig [65]byte) error {
	ret := _m.Called(sig)

	var r0 error
	if rf, ok := ret.Get(0).(func([65]byte) error); ok {
		r0 = rf(sig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewSignatureSetter interface {
	mock.TestingT
	Cleanup(func())
}

// NewSignatureSetter creates a new instance of SignatureSetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSignatureSetter(t mockConstructorTestingTNewSignatureSetter) *SignatureSetter {
	mock := &SignatureSetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
