package taskcontext

const (
	ErrMsgCreateTransactor  = "failed to create contract transactor"
	ErrMsgCallTransactor    = "failed to call transactor"
	ErrMsgTxAborted         = "tx aborted"
	ErrMsgExecutionReverted = "contract execution reverted"
	ErrMsgReceiptNotFound   = "receipt not found, maybe tx is pending"
	ErrMsgQueryReceipt      = "failed to query receipt "
)

// ----------error types ----------

type ErrTypCreateTransactor struct {
	Msg   string
	Cause error
}

func (e *ErrTypCreateTransactor) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgCreateTransactor
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypCreateTransactor) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypCallTransactor struct {
	Msg   string
	Cause error
}

func (e *ErrTypCallTransactor) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgCallTransactor
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypCallTransactor) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypTxAborted struct {
	Msg   string
	Cause error
}

func (e *ErrTypTxAborted) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxAborted
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypTxAborted) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypExecutionReverted struct {
	Msg   string
	Cause error
}

func (e *ErrTypExecutionReverted) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgExecutionReverted
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypExecutionReverted) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypReceiptNotFound struct {
	Msg   string
	Cause error
}

func (e *ErrTypReceiptNotFound) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgReceiptNotFound
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypReceiptNotFound) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypQueryReceipt struct {
	Msg   string
	Cause error
}

func (e *ErrTypQueryReceipt) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgQueryReceipt
	}
	if e.Cause == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Cause.Error()
}

func (e *ErrTypQueryReceipt) Unwrap() error {
	return e.Cause
}
