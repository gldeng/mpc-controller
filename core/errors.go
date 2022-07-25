package core

const (
	ErrMsgSignErr = "sign error"
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
