package core

const (
	ErrMsgSignErr         = "sign error"
	ErrMsgEmptySignResult = "sign result is empty"
)

// ----------error types ----------

type ErrTypSignErr struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypSignErr) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgSignErr
	}
	return e.ErrMsg
}

func (e *ErrTypSignErr) Unwrap() error {
	return e.Cause
}

// ----------

type ErrTypEmptySignResult struct {
	ErrMsg string
	Cause  error
}

func (e *ErrTypEmptySignResult) Error() string {
	if e.ErrMsg == "" {
		return ErrMsgEmptySignResult
	}
	return e.ErrMsg
}

func (e *ErrTypEmptySignResult) Unwrap() error {
	return e.Cause
}
