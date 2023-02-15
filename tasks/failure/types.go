package failure

import "fmt"

const (
	StatusInit Status = iota
	StatusTxSent
	StatusDone
)

type Status int

func (s Status) String() string {
	switch s {
	case StatusInit:
		return "Init"
	case StatusTxSent:
		return "TxSet"
	case StatusDone:
		return "Don"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

var (
	ErrMsgTimedOut = "task timeout"
)
