package taskcontext

import (
	"github.com/pkg/errors"
)

const (
	ErrMsgTransactorCreate = "failed to create contract transactor"
	ErrMsgTransactorCall   = "failed to call transactor"
	ErrMsgTxReverted       = "tx reverted"
	ErrMsgTxAborted        = "tx aborted"
	ErrMsgTxReceiptQuery   = "failed to query tx receipt "
	ErrMsgTxNotFound       = "tx not found, maybe it is pending"
)

var (
	ErrTxAborted = errors.New(ErrMsgTxAborted)
)

// ----------error types ----------

type ErrTypTransactorCreate struct {
	Msg   string
	Cause error
}

func (e *ErrTypTransactorCreate) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTransactorCreate
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypTransactorCreate) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypTransactorCall struct {
	Msg   string
	Cause error
}

func (e *ErrTypTransactorCall) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTransactorCall
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypTransactorCall) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypTxReverted struct {
	Msg   string
	Cause error
}

func (e *ErrTypTxReverted) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxReverted
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypTxReverted) Unwrap() error {
	return e.Cause
}

//// ----------
//
//type ErrTypTxAborted struct {
//	Msg   string
//	Cause error
//}
//
//func (e *ErrTypTxAborted) Error() string {
//	if e.Msg == "" {
//		e.Msg = ErrMsgTxAborted
//	}
//	if e.Cause == nil {
//		return e.Msg
//	}
//	return e.Msg + ": " + e.Cause.Error()
//}
//
//func (e *ErrTypTxAborted) Unwrap() error {
//	return e.Cause
//}

// ----------

type ErrTypTxReceiptQuery struct {
	Msg   string
	Cause error
}

func (e *ErrTypTxReceiptQuery) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxReceiptQuery
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypTxReceiptQuery) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypTxNotFound struct {
	Msg   string
	Cause error
}

func (e *ErrTypTxNotFound) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxNotFound
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypTxNotFound) Unwrap() error {
	return e.Cause
}
