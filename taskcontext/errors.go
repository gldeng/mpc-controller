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
	Msg string
	Pre error
}

func (e *ErrTypCreateTransactor) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgCreateTransactor
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypCreateTransactor) Cause() error {
	return e.Pre
}

func (e *ErrTypCreateTransactor) Unwrap() error {
	return e.Pre
}

// ----------

type ErrTypCallTransactor struct {
	Msg string
	Pre error
}

func (e *ErrTypCallTransactor) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgCallTransactor
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypCallTransactor) Cause() error {
	return e.Pre
}

func (e *ErrTypCallTransactor) Unwrap() error {
	return e.Pre
}

// ----------

type ErrTypTxAborted struct {
	Msg string
	Pre error
}

func (e *ErrTypTxAborted) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgTxAborted
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypTxAborted) Cause() error {
	return e.Pre
}

func (e *ErrTypTxAborted) Unwrap() error {
	return e.Pre
}

// ----------

type ErrTypExecutionReverted struct {
	Msg string
	Pre error
}

func (e *ErrTypExecutionReverted) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgExecutionReverted
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypExecutionReverted) Cause() error {
	return e.Pre
}

func (e *ErrTypExecutionReverted) Unwrap() error {
	return e.Pre
}

// ----------

type ErrTypReceiptNotFound struct {
	Msg string
	Pre error
}

func (e *ErrTypReceiptNotFound) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgReceiptNotFound
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypReceiptNotFound) Cause() error {
	return e.Pre
}

func (e *ErrTypReceiptNotFound) Unwrap() error {
	return e.Pre
}

// ----------

type ErrTypQueryReceipt struct {
	Msg string
	Pre error
}

func (e *ErrTypQueryReceipt) Error() string {
	if e.Msg == "" {
		e.Msg = ErrMsgQueryReceipt
	}
	return e.Msg + ": " + e.Pre.Error()
}

func (e *ErrTypQueryReceipt) Cause() error {
	return e.Pre
}

func (e *ErrTypQueryReceipt) Unwrap() error {
	return e.Pre
}
