package taskcontext

import (
	"github.com/pkg/errors"
)

const (
	ErrMsgTransactorCreate = "failed to create contract transactor"
	ErrMsgTransactorCall   = "failed to call contract transactor"
	ErrMsgTxReverted       = "tx reverted"
	ErrMsgTxStatusQuery    = "failed to query tx status"
	ErrMsgTxNotFound       = "tx not found, maybe currently pending in message pool"
	ErrMsgTxAborted        = "tx aborted, maybe caused by mpc concurrent requests"
)

var (
	ErrTxAborted = errors.New(ErrMsgTxAborted)
)

// ----------error types ----------

type ErrTypTransactorCreate struct {
	Msg string
	Pre error
}

func (e *ErrTypTransactorCreate) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTransactorCreate
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTransactorCreate) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTransactorCreate) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTransactorCall struct {
	Msg string
	Pre error
}

func (e *ErrTypTransactorCall) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTransactorCall
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTransactorCall) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTransactorCall) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTxReverted struct {
	Msg string
	Pre error
}

func (e *ErrTypTxReverted) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxReverted
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxReverted) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxReverted) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTxStatusQuery struct {
	Msg string
	Pre error
}

func (e *ErrTypTxStatusQuery) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxStatusQuery
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxStatusQuery) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxStatusQuery) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTxNotFound struct {
	Msg string
	Pre error
}

func (e *ErrTypTxNotFound) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxNotFound
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxNotFound) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxNotFound) Cause() error {
	return e.Pre
}
