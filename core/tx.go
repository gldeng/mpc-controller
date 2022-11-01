package core

type TxStatus = int

const (
	TxStatusUnknown    TxStatus = 0
	TxStatusCommitted  TxStatus = 4
	TxStatusAborted    TxStatus = 5
	TxStatusProcessing TxStatus = 6
	TxStatusDropped    TxStatus = 8
)

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
