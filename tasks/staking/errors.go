package staking

import (
	"fmt"
)

const (
	ErrMsgNonceRegress = "nonce regress"
	ErrMsgNonceJump    = "nonce jump"

	ErrMsgStakeStartTimeWillExpireIn5Seconds  = "stake start time will expire in 5 seconds"
	ErrMsgStakeStartTimeWillExpireIn10Seconds = "stake start time will expire in 10 seconds"
	ErrMsgStakeStartTimeWillExpireIn20Seconds = "stake start time will expire in 20 seconds"
	ErrMsgStakeStartTimeWillExpireIn40Seconds = "stake start time will expire in 40 seconds"
	ErrMsgStakeStartTimeWillExpireIn60Seconds = "stake start time will expire in 60 seconds"
)

// ----------error types ----------

type ErrTypNonceRegress struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypNonceRegress) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgNonceRegress + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
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
		return ErrMsgNonceJump + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypeNonceJump) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypStakeStartTimeWillExpireIn5Seconds struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypStakeStartTimeWillExpireIn5Seconds) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgStakeStartTimeWillExpireIn5Seconds + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypStakeStartTimeWillExpireIn5Seconds) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypStakeStartTimeWillExpireIn10Seconds struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypStakeStartTimeWillExpireIn10Seconds) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgStakeStartTimeWillExpireIn10Seconds + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypStakeStartTimeWillExpireIn10Seconds) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypStakeStartTimeWillExpireIn20Seconds struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypStakeStartTimeWillExpireIn20Seconds) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgStakeStartTimeWillExpireIn20Seconds + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypStakeStartTimeWillExpireIn20Seconds) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypStakeStartTimeWillExpireIn40Seconds struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypStakeStartTimeWillExpireIn40Seconds) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgStakeStartTimeWillExpireIn40Seconds + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypStakeStartTimeWillExpireIn40Seconds) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypStakeStartTimeWillExpireIn60Seconds struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypStakeStartTimeWillExpireIn60Seconds) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgStakeStartTimeWillExpireIn60Seconds + fmt.Sprintf(":%v", e.Cause)
	}
	return e.ErrMsg + fmt.Sprintf(":%v", e.Cause)
}

func (e *ErrTypStakeStartTimeWillExpireIn60Seconds) Unwrap() error {
	return e.Cause
}
