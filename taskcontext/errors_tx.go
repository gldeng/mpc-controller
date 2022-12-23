package taskcontext

import (
	"github.com/pkg/errors"
)

const (
	ErrMsgContractBindFail = "failed to bind contract"
	ErrMsgContractCallFail = "failed to call contract"

	ErrMsgTxIssueFail       = "failed to issue tx"
	ErrMsgTxStatusQueryFail = "failed to query tx status"

	ErrMsgTxReverted = "tx reverted"

	ErrMsgTxStatusInvalid  = "tx status invalid"
	ErrMsgTxStatusNotFound = "tx not found"
	ErrMsgTxStatusAborted  = "tx aborted"
	ErrMsgTxStatusDropped  = "tx dropped"
	ErrMsgTxStatusUnknown  = "tx status unknown"

	ErrMsgSharedMemoryFail = "failed to get shared memory"
	ErrMsgUTXOConsumeFail  = "failed to consume UTXO"
	ErrMsgTxInputsInvalid  = "invalid tx inputs"
)

var (
	ErrTxStatusAborted = errors.New(ErrMsgTxStatusAborted)
	ErrTxStatusDropped = errors.New(ErrMsgTxStatusDropped)
)

// ----------error types ----------

type ErrTypContractBindFail struct {
	Msg string
	Pre error
}

func (e *ErrTypContractBindFail) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgContractBindFail
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypContractBindFail) Unwrap() error {
	return e.Pre
}

func (e *ErrTypContractBindFail) Cause() error {
	return e.Pre
}

// ----------

type ErrTypContractCallFail struct {
	Msg string
	Pre error
}

func (e *ErrTypContractCallFail) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgContractCallFail
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypContractCallFail) Unwrap() error {
	return e.Pre
}

func (e *ErrTypContractCallFail) Cause() error {
	return e.Pre
}

// ----------

type ErrTypeTxIssueFail struct {
	Msg string
	Pre error
}

func (e *ErrTypeTxIssueFail) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxIssueFail
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypeTxIssueFail) Unwrap() error {
	return e.Pre
}

func (e *ErrTypeTxIssueFail) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTxStatusQueryFail struct {
	Msg string
	Pre error
}

func (e *ErrTypTxStatusQueryFail) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxStatusQueryFail
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxStatusQueryFail) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxStatusQueryFail) Cause() error {
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

type ErrTypTxStatusInvalid struct {
	Msg string
	Pre error
}

func (e *ErrTypTxStatusInvalid) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxStatusInvalid
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxStatusInvalid) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxStatusInvalid) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTxStatusNotFound struct {
	Msg string
	Pre error
}

func (e *ErrTypTxStatusNotFound) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxStatusNotFound
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxStatusNotFound) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxStatusNotFound) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTxStatusAborted struct {
	Msg string
	Pre error
}

func (e *ErrTypTxStatusAborted) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxStatusAborted
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxStatusAborted) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxStatusAborted) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTxStatusDropped struct {
	Msg string
	Pre error
}

func (e *ErrTypTxStatusDropped) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxStatusDropped
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxStatusDropped) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxStatusDropped) Cause() error {
	return e.Pre
}

// ----------

type ErrTypTxStatusUnknown struct {
	Msg string
	Pre error
}

func (e *ErrTypTxStatusUnknown) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxStatusUnknown
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxStatusUnknown) Unwrap() error {
	return e.Pre
}

func (e *ErrTypTxStatusUnknown) Cause() error {
	return e.Pre
}

// ----------

type ErrTypSharedMemoryFail struct {
	Msg string
	Pre error
}

func (e *ErrTypSharedMemoryFail) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgSharedMemoryFail
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypSharedMemoryFail) Unwrap() error {
	return e.Pre
}

func (e *ErrTypSharedMemoryFail) Cause() error {
	return e.Pre
}

// ----------

type ErrTypeUTXOConsumeFail struct {
	Msg string
	Pre error
}

func (e *ErrTypeUTXOConsumeFail) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgUTXOConsumeFail
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypeUTXOConsumeFail) Unwrap() error {
	return e.Pre
}

func (e *ErrTypeUTXOConsumeFail) Cause() error {
	return e.Pre
}

// ----------

type ErrTypeTxInputsInvalid struct {
	Msg string
	Pre error
}

func (e *ErrTypeTxInputsInvalid) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxInputsInvalid
	}
	if e.Pre == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypeTxInputsInvalid) Unwrap() error {
	return e.Pre
}

func (e *ErrTypeTxInputsInvalid) Cause() error {
	return e.Pre
}
