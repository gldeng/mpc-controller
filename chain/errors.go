package chain

import "fmt"

const (
	ErrMsgInsufficientFunds = "insufficient funds"
	ErrMsgInvalidNonce      = "invalid nonce"

	ErrMsgConflictAtomicInputs  = "conflicting atomic inputs"
	ErrMsgTxHasNoImportedInputs = "tx has no imported inputs"

	ErrMsgNotFound             = "not found"
	ErrMsgSharedMemoryNotFound = "shared memory not found"
	ErrMsgConsumedUTXONotFound = "consumed UTXO not found"
)

// ----------error types ----------

type ErrTypInsufficientFunds struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypInsufficientFunds) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgInsufficientFunds + fmt.Sprintf(": %+v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %+v", e.Cause)
}

func (e *ErrTypInsufficientFunds) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypInvalidNonce struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypInvalidNonce) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgInvalidNonce + fmt.Sprintf(": %+v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %+v", e.Cause)
}

func (e *ErrTypInvalidNonce) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypConflictAtomicInputs struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypConflictAtomicInputs) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgConflictAtomicInputs + fmt.Sprintf(": %+v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %+v", e.Cause)
}

func (e *ErrTypConflictAtomicInputs) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypTxHasNoImportedInputs struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypTxHasNoImportedInputs) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgTxHasNoImportedInputs + fmt.Sprintf(": %+v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %+v", e.Cause)
}

func (e *ErrTypTxHasNoImportedInputs) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypNotFound struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypNotFound) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgNotFound + fmt.Sprintf(": %+v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %+v", e.Cause)
}

func (e *ErrTypNotFound) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypSharedMemoryNotFound struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypSharedMemoryNotFound) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgSharedMemoryNotFound + fmt.Sprintf(": %+v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %+v", e.Cause)
}

func (e *ErrTypSharedMemoryNotFound) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypConsumedUTXONotFound struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypConsumedUTXONotFound) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgConsumedUTXONotFound + fmt.Sprintf(": %+v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %+v", e.Cause)
}

func (e *ErrTypConsumedUTXONotFound) Unwrap() error {
	return e.Cause
}
