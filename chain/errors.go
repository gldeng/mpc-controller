package chain

const (
	ErrMsgInsufficientFunds = "insufficient funds"
	ErrMsgInvalidNonce      = "invalid nonce"

	ErrMsgConflictAtomicInputs  = "conflicting atomic inputs"
	ErrMsgTxHasNoImportedInputs = "tx has no imported inputs"

	ErrMsgNotFound             = "due to: not found"
	ErrMsgSharedMemoryNotFound = "shared memory: not found"
)

// ----------error types ----------

type ErrTypInsufficientFunds struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypInsufficientFunds) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgInsufficientFunds
	}
	return e.ErrMsg
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
		return ErrMsgInvalidNonce
	}
	return e.ErrMsg
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
		return ErrMsgConflictAtomicInputs
	}
	return e.ErrMsg
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
		return ErrMsgTxHasNoImportedInputs
	}
	return e.ErrMsg
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
		return ErrMsgNotFound
	}
	return e.ErrMsg
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
		return ErrMsgSharedMemoryNotFound
	}
	return e.ErrMsg
}

func (e *ErrTypSharedMemoryNotFound) Unwrap() error {
	return e.Cause
}
