package transactor

import "fmt"

const (
	ErrMsgQuorumAlreadyReached = "quorum already reached"
	ErrMsgAttemptToRejoin      = "attempt to rejoin"

	ErrMsgTransactionFailed = "transaction failed"
)

// ----------error types ----------

type ErrTypQuorumAlreadyReached struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypQuorumAlreadyReached) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgQuorumAlreadyReached + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypQuorumAlreadyReached) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypAttemptToRejoin struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypAttemptToRejoin) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgAttemptToRejoin + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypAttemptToRejoin) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypTransactionFailed struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypTransactionFailed) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgTransactionFailed + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypTransactionFailed) Unwrap() error {
	return e.Cause
}
