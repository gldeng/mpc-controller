package keygen

const (
	StatusInit Status = iota
	StatusKeygenReqSent
	StatusKeygenReqDone
	StatusTxSent
	StatusDone
)

type Status int
