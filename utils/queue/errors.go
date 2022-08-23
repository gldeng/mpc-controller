package queue

import "fmt"

const (
	ErrMsgQueueIsFull = "queue is full"
)

// ----------error types ----------

type ErrTypQueueIsFull struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypQueueIsFull) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgQueueIsFull + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypQueueIsFull) Unwrap() error {
	return e.Cause
}

// ----------
