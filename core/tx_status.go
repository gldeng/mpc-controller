package core

import "fmt"

type TxStatus int

const (
	TxStatusUnknown    TxStatus = 0
	TxStatusCommitted  TxStatus = 4
	TxStatusAborted    TxStatus = 5
	TxStatusProcessing TxStatus = 6
	TxStatusDropped    TxStatus = 8
	TxStatusInvalid    TxStatus = 10
)

func (s TxStatus) String() string {
	switch s {
	case TxStatusUnknown:
		return "Unknown"
	case TxStatusCommitted:
		return "Committed"
	case TxStatusAborted:
		return "Aborted"
	case TxStatusProcessing:
		return "Processing"
	case TxStatusDropped:
		return "Dropped"
	default:
		return "invalid status"
	}
}

func IsPending(status TxStatus) bool {
	if status == TxStatusUnknown {
		return true
	}
	if status == TxStatusProcessing {
		return true
	}
	return false
}

func IsSuccessful(status TxStatus) bool {
	if status == TxStatusCommitted {
		return true
	}
	return false
}

func IsFailed(status TxStatus) bool {
	if status == TxStatusAborted {
		return true
	}
	if status == TxStatusDropped {
		return true
	}
	return false
}

type TxStatusWithReason struct {
	Status TxStatus
	Reason string
}

func (t *TxStatusWithReason) String() string {
	return fmt.Sprintf("%v: %v", t.Status.String(), t.Reason)
}
