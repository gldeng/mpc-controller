package staking

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	ErrMsgNonceRegress = "nonce regress"
	ErrMsgNonceJump    = "nonce jump"
)

var (
	ErrNonceRegress = errors.New("nonce regress")
	ErrNonceJump    = errors.New("nonce jump[")
)

// ----------error types ----------

type ErrTypNonceRegress struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypNonceRegress) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgNonceRegress + fmt.Sprintf(": %v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %v", e.Cause)
}

func (e *ErrTypNonceRegress) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypeNonceJump struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypeNonceJump) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgNonceJump + fmt.Sprintf(": %v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(": %v", e.Cause)
}

func (e *ErrTypeNonceJump) Unwrap() error {
	return e.Cause
}
