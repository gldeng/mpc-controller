package taskcontext

const (
	errMsgFailedToCreateContractTransactor = "failed to create contract transactor"
	errMsgContractExecutionReverted        = "contract execution reverted"
	errMsgTxReceiptNotFound                = "tx receipt not found, maybe it is pending"
)

// ----------error types ----------

type ErrTypFailedToCreateContractTransactor struct {
	msg string
	pre error
}

func (e *ErrTypFailedToCreateContractTransactor) Error() string {
	if e.msg == "" {
		e.msg = errMsgFailedToCreateContractTransactor
	}
	return e.msg + ": " + e.pre.Error()
}

func (e *ErrTypFailedToCreateContractTransactor) Cause() error {
	return e.pre
}

func (e *ErrTypFailedToCreateContractTransactor) Unwrap() error {
	return e.pre
}

// ----------

type ErrTypContractExecutionReverted struct {
	msg string
	pre error
}

func (e *ErrTypContractExecutionReverted) Error() string {
	if e.msg == "" {
		e.msg = errMsgContractExecutionReverted
	}
	return e.msg + ": " + e.pre.Error()
}

func (e *ErrTypContractExecutionReverted) Cause() error {
	return e.pre
}

func (e *ErrTypContractExecutionReverted) Unwrap() error {
	return e.pre
}

// ----------

type ErrTypTxReceiptNotFound struct {
	msg string
	pre error
}

func (e *ErrTypTxReceiptNotFound) Error() string {
	if e.msg == "" {
		e.msg = errMsgTxReceiptNotFound
	}
	return e.msg + ": " + e.pre.Error()
}

func (e *ErrTypTxReceiptNotFound) Cause() error {
	return e.pre
}

func (e *ErrTypTxReceiptNotFound) Unwrap() error {
	return e.pre
}
